package main

import (
	"os"
	"os/signal"

	. "./internal"

	. "github.com/docopt/docopt-go"
	"github.com/pkorotkov/logstick"
	"github.com/pkorotkov/safe"
	"golang.org/x/sys/unix"
)

var version = "0.2.0"

var usage = `diskloner is a console app for cloning disks and partitions.

Usage:
  diskloner monitor
  diskloner clone [-n <hfn> | --name <hfn>] <disk-path> <image-path>...
  diskloner inquire <disk-path> <info-path>...
  diskloner -h | --help
  diskloner -v | --version

Options:
  -n <hfn>, --name <hfn>  Human friendly name of cloning session [default: -].
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
	err := InitializeApp()
	if err != nil {
		log.Error("failed to initialize app: %s", err)
		safe.Exit(1)
	}
	switch {
	case args["monitor"]:
		MonitorStatus(quit)
	case args["clone"]:
		// Don't allow go further if app isn't run under root.
		if !IsRoot() {
			log.Error("application requires root privileges")
			safe.Exit(2)
		}
		var cs *CloningSession
		cs, err = NewCloningSession(args["--name"].(string), args["<disk-path>"].(string), args["<image-path>"].([]string))
		if err != nil {
			log.Error("failed to create cloner: %s", err)
			safe.Exit(3)
		}
		defer cs.Close()
		err = cs.Clone(quit)
		if err != nil {
			log.Error("failed to perform cloning: %s", err)
			safe.Exit(4)
		}
	case args["inquire"]:
		// TODO: Implement it.
		// Reserved error exit code is 5.
	default:
		log.Error("invalid set of arguments")
		safe.Exit(6)
	}
	safe.Exit(0)
}
