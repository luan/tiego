package commands

import (
	"os"
	"regexp"

	"github.com/luan/teapot"
	"github.com/luan/tiego/say"
	"github.com/tmtk75/cli"
)

func DeleteWokstation(c *cli.Context) {
	teapotAddr := c.GlobalString("teapot")
	client := teapot.NewClient(teapotAddr)
	name, _ := c.ArgFor("name")

	err := client.DeleteWorkstation(name)
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
	say.Println(0, say.Bold(say.Green("OK")))
}
