package main

import (
	rpc "github.com/travis-ci/imaged/rpc/images"
	"github.com/travis-ci/imaged/server"
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
			Usage:  "Local path where Packer templates Git repo should be checked out",
			EnvVar: "IMAGED_TEMPLATES_PATH",
			Value:  "/templates",
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
	worker, err := worker.New(worker.Config{
		TemplatesPath: c.String("templates-path"),
	})
	if err != nil {
		return err
	}

	go worker.Run()

	server, err := server.New(server.Config{
		DatabaseURL:  c.String("database"),
		RecordBucket: c.String("bucket"),
		Worker:       worker,
	})
	if err != nil {
		return err
	}

	if c.Bool("migrate") {
		if err = server.DB.Migrate(); err != nil {
			return err
		}
	}

	handler := rpc.NewImagesServer(server, nil)
	return http.ListenAndServe(":8080", handler)
}
