package worker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/travis-ci/imaged/db"
	"github.com/travis-ci/imaged/storage"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"io/ioutil"
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
	j.db().StartBuild(ctx, j.Build)

	l := log.WithFields(log.Fields{
		"build_id": j.Build.ID,
		"name":     j.Build.Name,
		"revision": j.Build.Revision,
	})
	l.Info("started build")

	// Assume the build fails unless we get to the end and mark it successful
	j.Build.Status = db.BuildStatusFailed
	defer j.db().FinishBuild(ctx, j.Build)

	rev, err := j.resetRepository(ctx)
	if err != nil {
		return err
	}
	l.WithField("resolved", rev).Info("checked out templates")

	j.Build.FullRevision = &rev
	j.db().UpdateBuild(ctx, j.Build)

	dir, err := ioutil.TempDir("", "imaged-build")
	if err != nil {
		return errors.Wrap(err, "could not create build output directory")
	}
	j.outputDir = dir
	defer os.RemoveAll(dir)
	l.WithField("out_dir", dir).Debug("created build output directory")

	logFile, err := os.Create(j.outputFile("build.log"))
	if err != nil {
		return errors.Wrap(err, "could not create build log file")
	}
	j.log = logFile
	defer logFile.Close()
	l.Debug("created build log")

	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()

	if err := j.installSecrets(ctx); err != nil {
		return err
	}
	l.Info("installed secrets file")

	cmd := exec.CommandContext(ctx, j.packer(), "version")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err = cmd.Run(); err != nil {
		return errors.Wrap(err, "could not print Packer version")
	}
	logWriter.Flush()
	l.Debug("printed packer version")

	template, err := j.convertTemplateToJSON()
	if err != nil {
		return err
	}
	l.Debug("converted template from YAML to JSON")

	recordsDir := filepath.Join(dir, "records")
	if err = os.Mkdir(recordsDir, 0777); err != nil {
		return err
	}
	l.WithField("records_path", recordsDir).Debug("created custom records directory")

	packerSucceeded := true
	cmd = exec.CommandContext(ctx, j.packer(), "build", "-color=false", "-var", "records_path="+recordsDir, template)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Dir = j.templatesDir()
	if err = cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(logWriter, "packer exited with non-zero status code: %v\n", cmd.ProcessState)
			l.WithError(err).Info("packer build failed")
			packerSucceeded = false
		} else {
			return errors.Wrap(err, "could not run Packer build")
		}
	} else {
		l.Info("packer build succeeded")
	}
	logWriter.Flush()
	logFile.Sync()

	if err = j.createRecords(ctx, l, recordsDir); err != nil {
		return err
	}

	if packerSucceeded {
		j.Build.Status = db.BuildStatusSucceeded
	}
	l.WithField("status", j.Build.Status).Info("build finished")

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

func (j *Job) createRecords(ctx context.Context, l *log.Entry, recordsDir string) error {
	records, err := ioutil.ReadDir(recordsDir)
	if err != nil {
		return err
	}

	for _, f := range records {
		rlog := l.WithField("file", f.Name())

		path := filepath.Join(recordsDir, f.Name())
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "could not open record file upload")
		}
		defer file.Close()

		if r, err := j.createRecord(ctx, file); err != nil {
			rlog.WithError(err).Error("failed to upload record")
		} else {
			rlog.WithField("record_id", r.ID).Info("uploaded record")
		}
	}

	logRead, err := os.Open(j.log.Name())
	if err != nil {
		return errors.Wrap(err, "could not open log file for reading")
	}
	defer logRead.Close()

	if r, err := j.createRecord(ctx, logRead); err != nil {
		l.WithError(err).Error("failed to upload build log")
	} else {
		l.WithField("record_id", r.ID).Info("uploaded build log")
	}
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
