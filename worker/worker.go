package worker

import (
	"context"
	"github.com/pkg/errors"
	"github.com/travis-ci/imaged/db"
	"github.com/travis-ci/imaged/storage"
	"gopkg.in/src-d/go-git.v4"
	"log"
)

// Worker waits for and runs a single Packer build at a time.
type Worker struct {
	jobs   chan Job
	config Config
	repo   *git.Repository
}

// Config contains options for configuring a new Worker.
type Config struct {
	// TemplatesPath is the local path of the Git clone of the Packer templates.
	TemplatesPath string
	// TemplatesURL is the URL where the Packer templates should be cloned from.
	TemplatesURL string
	// Packer is the path to the Packer executable.
	Packer string
	// DB is the database connection jobs should use.
	DB *db.Connection
	// Storage is the storage jobs should use to upload records.
	Storage *storage.Storage
}

// New creates a new worker ready to run jobs.
func New(c Config) (*Worker, error) {
	w := &Worker{
		jobs:   make(chan Job),
		config: c,
	}

	if err := w.initTemplates(); err != nil {
		return nil, err
	}

	return w, nil
}

// Send asks the worker to run a job.
//
// Returns an error if a job is already running on the worker.
func (w *Worker) Send(j Job) error {
	j.worker = w

	select {
	case w.jobs <- j:
		return nil
	default:
		return errors.New("a job is already running on this worker")
	}
}

// Run waits for new jobs and runs them as they come in.
//
// It should be called in a goroutine.
func (w *Worker) Run() {
	for j := range w.jobs {
		ctx := context.Background()

		if err := j.Execute(ctx); err != nil {
			log.Printf("Error running job: %v", err)
		}
	}
}

func (w *Worker) initTemplates() error {
	if w.config.TemplatesPath == "" {
		return errors.New("a templates path is required when creating a worker")
	}

	r, err := git.PlainOpen(w.config.TemplatesPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			r, err = w.cloneTemplates()
			if err != nil {
				return err
			}
		} else {
			return errors.Wrap(err, "could not open existing templates repo")
		}
	}

	w.repo = r
	if err = w.updateTemplates(); err != nil {
		return err
	}

	return nil
}

func (w *Worker) cloneTemplates() (*git.Repository, error) {
	if w.config.TemplatesURL == "" {
		return nil, errors.New("a templates URL is required when templates are not already cloned")
	}

	r, err := git.PlainClone(w.config.TemplatesPath, false, &git.CloneOptions{
		URL: w.config.TemplatesURL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not clone templates repo")
	}

	return r, nil
}

func (w *Worker) updateTemplates() error {
	if err := w.repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return errors.Wrap(err, "could not fetch latest commits for templates repo")
		}
	}

	return nil
}
