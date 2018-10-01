package worker

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"os"
)

// Worker waits for and runs a single Packer build at a time.
type Worker struct {
	jobs   chan Job
	config Config
}

// Config contains options for configuring a new Worker.
type Config struct {
	// TemplatesPath is the local path of the Git clone of the Packer templates.
	TemplatesPath string
	// TemplatesURL is the URL where the Packer templates should be cloned from.
	TemplatesURL string
	// Packer is the path to the Packer executable.
	Packer string
}

// New creates a new worker ready to run jobs.
func New(c Config) (*Worker, error) {
	if c.TemplatesPath == "" {
		return nil, errors.New("a templates path is required when creating a worker")
	}

	if _, err := os.Stat(c.TemplatesPath); err != nil {
		return nil, errors.Wrap(err, "could not access templates path")
	}

	return &Worker{
		jobs:   make(chan Job),
		config: c,
	}, nil
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
