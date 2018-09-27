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

	app.Action = Run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Run starts the imaged server listening for API requests.
func Run(c *cli.Context) error {
	server := &imaged.Server{}
	handler := rpc.NewImagesServer(server, nil)
	return http.ListenAndServe(":8080", handler)
}
