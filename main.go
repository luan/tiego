package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/codegangsta/cli"
	"github.com/gorilla/websocket"
	"github.com/pkg/term"
)

var domain = "tiego"
var routeRoot string
var client receptor.Client

func readLoop(c *websocket.Conn, w io.Writer, done chan bool) {
	for {
		_, m, err := c.ReadMessage()
		if err != nil {
			done <- true
			return
		}

		w.Write(m)
	}
}

func writeLoop(c *websocket.Conn, r io.Reader, done chan bool) {
	br := bufio.NewReader(r)
	for {
		x, size, err := br.ReadRune()
		if size <= 0 || err != nil {
			done <- true
			return
		}

		p := make([]byte, size)
		utf8.EncodeRune(p, x)

		err = c.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			done <- true
			return
		}
	}
}

func attachWorkstation(c *cli.Context) {
	name := c.Args().First()

	route := fmt.Sprintf("ws://%s-%s.%s:80/shell", name, domain, routeRoot)
	u, _ := url.Parse(route)

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		panic(err)
	}

	ws, _, err := websocket.NewClient(conn, u, http.Header{"Origin": {route}}, 1024, 1024)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	var in io.Reader

	term, err := term.Open(os.Stdin.Name())
	if err == nil {
		err = term.SetRaw()
		if err != nil {
			log.Fatalln("failed to set raw:", term)
		}

		defer term.Restore()

		in = term
	} else {
		in = os.Stdin
	}

	done := make(chan bool)
	go readLoop(ws, os.Stdout, done)
	go writeLoop(ws, in, done)
	<-done
}

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
		Setup: &models.SerialAction{
			Actions: []models.Action{
				&models.DownloadAction{
					From:     "https://dl.dropboxusercontent.com/u/33868236/tea.tar.gz",
					To:       "/tmp",
					CacheKey: "tea",
				},
				&models.DownloadAction{
					From:     "http://onsi-public.s3.amazonaws.com/crusher.tar.gz",
					To:       "/tmp",
					CacheKey: "crusher",
				},
			},
		},
		Action: &models.RunAction{
			Path:      "/tmp/tea",
			LogSource: "TEA",
		},
		Monitor: &models.RunAction{
			Path:      "/tmp/crusher",
			Args:      []string{"--port-check=8080"},
			LogSource: "CRUSHER",
		},
		DiskMB:    128,
		MemoryMB:  64,
		Ports:     []uint32{8080},
		Routes:    []string{route},
		LogGuid:   processGuid,
		LogSource: "TIEGO",
	})
	fmt.Println(route)
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
		{
			Name:      "attach",
			ShortName: "a",
			Usage:     "attach to a workstation",
			Action:    attachWorkstation,
		},
	}
	app.Run(os.Args)
}
