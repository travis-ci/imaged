package main

import (
	"github.com/travis-ci/imaged"
	rpc "github.com/travis-ci/imaged/rpc/images"
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
	}

	app.Action = Run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Run starts the imaged server listening for API requests.
func Run(c *cli.Context) error {
	dbURL := c.String("database")
	server, err := imaged.NewServer(dbURL)
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
