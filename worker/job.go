package worker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/travis-ci/imaged"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Job describes a Packer build that the worker needs to run.
type Job struct {
	Build *imaged.Build

	worker    *Worker
	log       *os.File
	outputDir string
}

// Execute runs a single job.
func (j *Job) Execute(ctx context.Context) error {
	log.Printf("Build %d: building template '%s' at revision '%s'", j.Build.ID, j.Build.Name, j.Build.Revision)

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

	cmd := exec.CommandContext(ctx, j.packer(), "version")
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err = cmd.Run(); err != nil {
		return errors.Wrap(err, "could not print Packer version")
	}
	logWriter.Flush()

	// TODO: ensure the templates are at the right revision and in a clean state

	template, err := j.convertTemplateToJSON()
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, j.packer(), "build", template)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Dir = j.templatesDir()
	if err = cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(logWriter, "packer exited with non-zero status code: %v\n", cmd.ProcessState)
		} else {
			return errors.Wrap(err, "could not run Packer build")
		}
	}
	logWriter.Flush()
	logFile.Sync()

	// TODO: replace this with attaching the log as a record
	logContents, err := ioutil.ReadFile(logFile.Name())
	os.Stdout.Write(logContents)
	log.Printf("Build %d completed", j.Build.ID)

	return nil
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
