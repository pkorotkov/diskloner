package main

import (
	"os"
	"os/signal"

	. "github.com/docopt/docopt-go"
	"github.com/pkorotkov/logstick"
	"github.com/pkorotkov/safe"
	"golang.org/x/sys/unix"
)

var version = "0.1"

var usage = `diskloner is a console app for cloning disks and partitions.

Usage:
  diskloner status
  diskloner clone [-n <hfn> | --name <hfn>] <disk-path> <image-path>...
  diskloner -h | --help
  diskloner -v | --version

Options:
  -n <hfn>, --name <hfn>  Human friendly name of cloning session [default: UNNAMED].
  -h, --help              Show this message.
  -v, --version           Show version.`

var (
	log = logstick.NewLogger()
)

func init() {
	_ = log.AddWriter(os.Stderr)
}

func main() {
	defer safe.CatchExit()
	defer log.Close()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT)
	signal.Notify(quit, unix.SIGTERM)
	args, _ := Parse(usage, nil, true, version, false)
	switch {
	case args["status"]:
		monitorStatus(quit)
	case args["clone"]:
		// Don't allow go further if app isn't run under root.
		if !isRoot() {
			log.Error("application requires root privileges")
			safe.Exit(2)
		}
		cs, err := NewCloningSession(args["--name"].(string), args["<disk-path>"].(string), args["<image-path>"].([]string))
		if err != nil {
			log.Error("failed to create cloner: %s", err)
			safe.Exit(3)
		}
		defer cs.Close()
		cs.Clone(quit)
	default:
		log.Error("invalid set of arguments")
		safe.Exit(4)
	}
	safe.Exit(0)
}
