# Tiego

Tiego is a CLI to manage and workstations running on [Diego](https://github.com/cloudfoundry-incubator/diego-release). It allows you to pick your container flavor with [Docker](http://docker.com) images.

## Setup

Just go get it: `go get github.com/luan/tiego`

## Usage

*Optionally* set a TEAPOT url:

```bash
export TEAPOT=http://teapot.192.168.11.11
```

Run commands:

```
NAME:
   tiego - manages tiego workstations and shell sessions

USAGE:
   tiego [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   create, c    creates a workstation
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --teapot, -t 'http://127.0.0.1:8080' address of the Teapot to use [$TEAPOT]
   --help, -h                           show help
   --version, -v                        print the version
```
