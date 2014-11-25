package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/codegangsta/cli"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/pivotal-cf-experimental/veritas/say"
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

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func resize(ws *websocket.Conn) {
	rows, cols, _ := pty.Getsize(os.Stdin)
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	enc.Encode(Winsize{Height: uint16(rows), Width: uint16(cols), x: 0, y: 0})
	ws.WriteMessage(websocket.BinaryMessage, buffer.Bytes())
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

	resized := make(chan os.Signal, 10)
	signal.Notify(resized, syscall.SIGWINCH)

	go func() {
		for {
			<-resized
			resize(ws)
		}
	}()
	resize(ws)

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
			Path:       "/tmp/tea",
			LogSource:  "TEA",
			Privileged: true,
		},
		Monitor: &models.RunAction{
			Path:      "/tmp/crusher",
			Args:      []string{"--port-check=8080"},
			LogSource: "CRUSHER",
		},
		DiskMB:    6000,
		MemoryMB:  256,
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

func contains(s []receptor.ActualLRPResponse, e string) int {
	for i, a := range s {
		if a.ProcessGuid == e {
			return i
		}
	}
	return -1
}

func namify(guid string) string {
	return strings.Replace(guid, "tiego-", "", 1)
}

func listWorkstations(c *cli.Context) {
	desiredLRPs, _ := client.DesiredLRPsByDomain(domain)
	actualLRPs, _ := client.ActualLRPsByDomain(domain)
	lrps := []receptor.ActualLRPResponse{}

	say.Println(0, say.Cyan("\nWorkstations:\n"))
	say.Println(1, "%-30s %-40s %-30s", "Name", "Docker Image", "State")
	say.Println(1, "----------------------------------------------------------------------------------------------------")

	for _, lrp := range actualLRPs {
		lrps = append(lrps, lrp)
	}

	for _, lrp := range desiredLRPs {
		state := "STOPPED"
		if i := contains(lrps, lrp.ProcessGuid); i >= 0 {
			state = lrps[i].State
		}
		name := namify(lrp.ProcessGuid)
		say.Println(1, "%-30s %-40s %-30s", name, strings.Replace(lrp.RootFSPath, "#", ":", 1), say.Yellow(state))
	}
}

func deleteWorstation(c *cli.Context) {
	name := c.Args().First()
	processGuid := fmt.Sprintf("%s-%s", domain, name)
	err := client.DeleteDesiredLRP(processGuid)
	if err != nil {
		say.Println(0, say.Red(strings.Replace(err.Error(), "LRP", "Workstation", 1)))
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
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "list available workstation",
			Action:    listWorkstations,
		},
		{
			Name:      "destroy",
			ShortName: "d",
			Usage:     "destroy a workstation",
			Action:    deleteWorstation,
		},
	}
	app.Run(os.Args)
}
