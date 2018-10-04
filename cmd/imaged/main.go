package main

import (
	"github.com/travis-ci/imaged/db"
	rpc "github.com/travis-ci/imaged/rpc/images"
	"github.com/travis-ci/imaged/server"
	"github.com/travis-ci/imaged/storage"
	"github.com/travis-ci/imaged/worker"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "imaged"
	app.Description = "build Packer images at your request"
	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:   "migrate",
			Usage:  "run database migrations before starting the server",
			EnvVar: "IMAGED_RUN_MIGRATIONS",
		},
		cli.StringFlag{
			Name:   "database, d",
			Usage:  "URL for connecting to the PostgreSQL database",
			EnvVar: "IMAGED_DATABASE_URL",
		},
		cli.StringFlag{
			Name:   "bucket, b",
			Usage:  "S3 bucket name for storing build records",
			EnvVar: "IMAGED_RECORD_BUCKET",
		},
		cli.StringFlag{
			Name:   "templates-path",
			Usage:  "local path where Packer templates Git repo should be checked out",
			EnvVar: "IMAGED_TEMPLATES_PATH",
			Value:  "/templates",
		},
		cli.StringFlag{
			Name:   "templates-url",
			Usage:  "URL for Git repo containing Packer templates",
			EnvVar: "IMAGED_TEMPLATES_URL",
		},
		cli.StringFlag{
			Name:   "packer",
			Usage:  "path to the Packer executable",
			EnvVar: "IMAGED_PACKER_PATH",
			Value:  "/bin/packer",
		},
		cli.StringFlag{
			Name:   "secrets",
			Usage:  "local path to a file containing secrets for Linux Ansible playbooks",
			EnvVar: "IMAGED_ANSIBLE_SECRETS_FILE",
		},
	}

	app.Action = Run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Run starts the imaged server listening for API requests.
func Run(c *cli.Context) error {
	db, err := db.NewConnection(c.String("database"))
	if err != nil {
		return err
	}

	storage, err := storage.New(c.String("bucket"))
	if err != nil {
		return err
	}

	worker, err := worker.New(worker.Config{
		TemplatesPath:      c.String("templates-path"),
		TemplatesURL:       c.String("templates-url"),
		Packer:             c.String("packer"),
		AnsibleSecretsFile: c.String("secrets"),
		DB:                 db,
		Storage:            storage,
	})
	if err != nil {
		return err
	}

	go worker.Run()

	server := &server.Server{
		DB:      db,
		Storage: storage,
		Worker:  worker,
	}

	if c.Bool("migrate") {
		if err = db.Migrate(); err != nil {
			return err
		}
	}

	handler := rpc.NewImagesServer(server, nil)
	return http.ListenAndServe(":8080", handler)
}
