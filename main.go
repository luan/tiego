package main

import (
	"os"

	"github.com/luan/tiego/commands"
	"github.com/tmtk75/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tiego"
	app.Usage = "manages tiego workstations and shell sessions"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "teapot, t",
			Value:  "http://127.0.0.1:8080",
			Usage:  "address of the Teapot to use",
			EnvVar: "TEAPOT",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "creates a workstation",
			Action:    commands.CreateWokstation,
			Args:      "<name>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "docker-image, d",
					Value: "docker:///ubuntu#trusty",
					Usage: "docker image available on hub.docker.com",
				},
			},
		},
		{
			Name:      "delete",
			ShortName: "d",
			Usage:     "deltes a workstation",
			Action:    commands.DeleteWokstation,
			Args:      "<name>",
		},
	}

	app.Run(os.Args)
}
