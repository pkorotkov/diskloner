package main

import (
	"os"
	"os/signal"

	. "github.com/pkorotkov/diskloner/internal"

	. "github.com/docopt/docopt-go"
	"github.com/hexbox/logstick"
	"github.com/pkorotkov/safe"
	"golang.org/x/sys/unix"
)

var version = "v1.0.1"

var usage = `diskloner is a console app for cloning disks and partitions.

Usage:
  diskloner monitor
  diskloner clone [-n <session-name> | --name <session-name>] <disk-path> <image-path>...
  diskloner inquire <disk-path> <info-path>...
  diskloner -h | --help
  diskloner -v | --version

Options:
  -n <session-name>, --name <session-name>  Human friendly name of cloning session [default: -].
  -h, --help                                Show this message.
  -v, --version                             Show version.

Commands:
  monitor  Start monitoring local cloning processes.
  clone    Start a named cloning process (session).
  inquire  To be implemented.`

var (
	log = logstick.NewLogger()
)

func init() {
	_ = log.AddWriter(os.Stderr)
}

func main() {
	defer safe.CatchExit()
	defer log.Close()
	var err error
	// Quit channel is controlled by OS signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT)
	signal.Notify(quit, unix.SIGTERM)
	args, _ := Parse(usage, nil, true, version, false)
	// Do all neccessary initialization.
	// For now it's not allowed to customize its setup.
	if err = InitializeApp(); err != nil {
		if _, ok := err.(ErrProgressDirectory); ok {
			log.Error("app has not enough privileges: %s", err)
		} else {
			log.Error("failed to initialize app: %s", err)
		}
		safe.Exit(1)
	}
	// Go through commands.
	switch {
	case args["monitor"]:
		MonitorStatus(quit)
	case args["clone"]:
		// Don't allow go further if app isn't run under root.
		if !IsRoot() {
			log.Error("app requires root privileges")
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
