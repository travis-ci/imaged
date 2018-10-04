package worker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/travis-ci/imaged/db"
	"github.com/travis-ci/imaged/storage"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Job describes a Packer build that the worker needs to run.
type Job struct {
	Build *db.Build

	worker    *Worker
	log       *os.File
	outputDir string
}

// Execute runs a single job.
func (j *Job) Execute(ctx context.Context) error {
	log.Printf("Build %d: building template '%s' at revision '%s'", j.Build.ID, j.Build.Name, j.Build.Revision)
	j.db().StartBuild(ctx, j.Build)

	// Assume the build fails unless we get to the end and mark it successful
	j.Build.Status = db.BuildStatusFailed
	defer j.db().FinishBuild(ctx, j.Build)

	rev, err := j.resetRepository(ctx)
	if err != nil {
		return err
	}

	log.Printf("Build %d: checked out revision %s", j.Build.ID, rev)
	j.Build.FullRevision = &rev
	j.db().UpdateBuild(ctx, j.Build)

	dir, err := ioutil.TempDir("", "imaged-build")
	if err != nil {
		return errors.Wrap(err, "could not create build output directory")
	}
	j.outputDir = dir
	defer os.RemoveAll(dir)

	logFile, err := os.Create(j.outputFile("build.log"))
	if err != nil {
		return errors.Wrap(err, "could not create build log file")
	}
	j.log = logFile
	defer logFile.Close()

	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()

	if err := j.installSecrets(ctx); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, j.packer(), "version")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err = cmd.Run(); err != nil {
		return errors.Wrap(err, "could not print Packer version")
	}
	logWriter.Flush()

	template, err := j.convertTemplateToJSON()
	if err != nil {
		return err
	}

	packerSucceeded := true
	cmd = exec.CommandContext(ctx, j.packer(), "build", "-color=false", template)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Dir = j.templatesDir()
	if err = cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(logWriter, "packer exited with non-zero status code: %v\n", cmd.ProcessState)
			packerSucceeded = false
		} else {
			return errors.Wrap(err, "could not run Packer build")
		}
	}
	logWriter.Flush()
	logFile.Sync()

	logRead, err := os.Open(logFile.Name())
	if err != nil {
		return errors.Wrap(err, "could not open log file for reading")
	}
	defer logRead.Close()

	j.createRecord(ctx, logRead)

	if packerSucceeded {
		j.Build.Status = db.BuildStatusSucceeded
	}
	log.Printf("Build %d completed", j.Build.ID)

	return nil
}

func (j *Job) storage() *storage.Storage {
	return j.worker.config.Storage
}

func (j *Job) db() *db.Connection {
	return j.worker.config.DB
}

func (j *Job) repo() *git.Repository {
	return j.worker.repo
}

func (j *Job) packer() string {
	return j.worker.config.Packer
}

func (j *Job) templatesDir() string {
	return j.worker.config.TemplatesPath
}

func (j *Job) outputFile(name string) string {
	return filepath.Join(j.outputDir, name)
}

func (j *Job) resetRepository(ctx context.Context) (string, error) {
	// Fetch any new commits since the process started
	if err := j.worker.updateTemplates(); err != nil {
		return "", err
	}

	// We need to resolve the reference they gave us
	rev := "origin/" + j.Build.Revision
	h, err := j.repo().ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			h, err = j.repo().ResolveRevision(plumbing.Revision(j.Build.Revision))
		}

		if err != nil {
			return "", errors.Wrap(err, "could not resolve reference in templates repo")
		}
	}

	// Check out the resolved revision, discarding any local changes
	w, err := j.repo().Worktree()
	if err != nil {
		return "", errors.Wrap(err, "could not get worktree for templates repo")
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  *h,
		Force: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "could not checkout templates revision")
	}

	rev = h.String()
	return rev, nil
}

func (j *Job) convertTemplateToJSON() (string, error) {
	ymlPath := filepath.Join(j.templatesDir(), "templates", j.Build.Name+".yml")
	yml, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		return "", errors.Wrap(err, "could not read template YAML")
	}

	jsonPath := j.outputFile(j.Build.Name + ".json")

	json, err := yaml.YAMLToJSON(yml)
	if err != nil {
		return "", errors.Wrap(err, "could not convert template YAML to JSON")
	}

	if err = ioutil.WriteFile(jsonPath, json, 0644); err != nil {
		return "", errors.Wrap(err, "could not create file for template JSON")
	}

	return jsonPath, nil
}

func (j *Job) installSecrets(ctx context.Context) error {
	srcPath := j.worker.config.AnsibleSecretsFile
	destPath := filepath.Join(j.templatesDir(), "linux_playbooks", "secrets.yml")

	src, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "could not open secrets file")
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return errors.Wrap(err, "could not create new secrets file")
	}
	defer dest.Close()

	if _, err = io.Copy(dest, src); err != nil {
		return errors.Wrap(err, "could not copy secrets file contents")
	}

	dest.Sync()

	return nil
}

func (j *Job) createRecord(ctx context.Context, f *os.File) (*db.Record, error) {
	name := filepath.Base(f.Name())
	key := j.Build.RecordKey(name)
	if _, err := j.storage().Upload(ctx, key, f); err != nil {
		return nil, errors.Wrap(err, "could not upload file to S3")
	}

	record, err := j.db().CreateRecord(ctx, j.Build, name, key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create record for uploaded file")
	}

	return record, nil
}
