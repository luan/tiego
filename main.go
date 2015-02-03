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
			Action:    commands.CreateWorkstation,
			Args:      "<name>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "docker-image, d",
					Value: "docker:///ubuntu#trusty",
					Usage: "docker image available on hub.docker.com",
				},
				cli.IntFlag{
					Name:  "cpu, c",
					Value: 1,
					Usage: "cpu weight allocated for workstation in mb",
				},
				cli.IntFlag{
					Name:  "disk, k",
					Value: 2048,
					Usage: "disk space allocated for workstation in mb",
				},
				cli.IntFlag{
					Name:  "memory, m",
					Value: 256,
					Usage: "memory allocated for workstation in mb",
				},
			},
		},
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "lists all workstations",
			Action:    commands.ListWorkstations,
		},
		{
			Name:      "delete",
			ShortName: "d",
			Usage:     "deletes a workstation",
			Action:    commands.DeleteWorkstation,
			Args:      "<name>",
		},
		{
			Name:      "attach",
			ShortName: "a",
			Usage:     "attach to a workstation",
			Action:    commands.AttachWorkstation,
			Args:      "<name>",
		},
	}

	app.Run(os.Args)
}
