package commands

import (
	"os"
	"regexp"

	"github.com/luan/teapot"
	"github.com/luan/tiego/say"
	"github.com/tmtk75/cli"
)

func ListWorkstations(c *cli.Context) {
	teapotAddr := c.GlobalString("teapot")
	client := teapot.NewClient(teapotAddr)

	workstations, err := client.ListWorkstations()
	if err != nil {
		say.Print(0, say.Bold(say.Red("FAILED: ")))
		var errorMessage string
		if len(err.Error()) > 0 {
			fieldRegexp, _ := regexp.Compile("([^:]*: )([^,]*)(,?)")
			errorMessage = fieldRegexp.ReplaceAllString(err.Error(), "$1"+say.Cyan("$2")+"$3")
		} else {
			errorMessage = "Could not talk to Teapot, did you set the " + say.Cyan("TEAPOT") + " url correctly?"
		}
		say.Println(0, errorMessage)
		os.Exit(1)
	}

	if len(workstations) == 0 {
		say.Println(0, "No workstations found.")
		os.Exit(0)
	}

	say.Println(0, say.Bold("%-20s %-40s %-10s", "Name", "Docker Image", "State"))
	for _, workstation := range workstations {
		var state string
		switch workstation.State {
		case "RUNNING":
			state = say.Green("%s", workstation.State)
		case "CLAIMED":
			state = say.Yellow("%s", workstation.State)
		case "STOPPED":
			state = say.Gray("%s", workstation.State)
		case "CRASHED":
			state = say.Red("%s", workstation.State)
		case "UNCLAIMED":
			state = say.LightGray("%s", workstation.State)
		}
		say.Println(0, "%-20s %-40s %-10s", workstation.Name, workstation.DockerImage, state)
	}
}
