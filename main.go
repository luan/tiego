package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/codegangsta/cli"
)

var domain = "tiego"
var routeRoot string
var client receptor.Client

func createWokstation(c *cli.Context) {
	name := c.Args().First()
	dockerImage := c.Args().Get(1)
	processGuid := fmt.Sprintf("%s-%s", domain, name)

	route := fmt.Sprintf("%s-%s.%s", name, domain, routeRoot)
	err := client.CreateDesiredLRP(receptor.DesiredLRPCreateRequest{
		ProcessGuid: processGuid,
		Domain:      domain,
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  dockerImage,
		// Setup: &models.SerialAction{
		// 	Actions: []models.Action{
		// 		&models.DownloadAction{
		// 			From:     "http://onsi-public.s3.amazonaws.com/riker.tar.gz",
		// 			To:       "/tmp",
		// 			CacheKey: "riker",
		// 		},
		// 		&models.DownloadAction{
		// 			From:     "http://onsi-public.s3.amazonaws.com/crusher.tar.gz",
		// 			To:       "/tmp",
		// 			CacheKey: "crusher",
		// 		},
		// 	},
		// },
		Action: &models.RunAction{
			Path: "echo",
		},
		// Monitor: &models.RunAction{
		// 	Path:      "/tmp/crusher",
		// 	Args:      []string{"--port-check=8080"},
		// 	LogSource: "CRUSHER",
		// },
		DiskMB:    128,
		MemoryMB:  64,
		Ports:     []uint32{8080},
		Routes:    []string{route},
		LogGuid:   processGuid,
		LogSource: "TIEGO",
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	receptorAddr := os.Getenv("RECEPTOR")
	if receptorAddr == "" {
		panic("No RECEPTOR set")
	}

	client = receptor.NewClient(receptorAddr)
	routeRoot = strings.Split(receptorAddr, "receptor.")[1]

	app := cli.NewApp()
	app.Name = "create"
	app.Usage = "creates a diego workstation"

	app.Commands = []cli.Command{
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "creates a workstation",
			Action:    createWokstation,
		},
	}
	app.Run(os.Args)
}
