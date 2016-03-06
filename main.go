package main

import (
	"os"

	. "github.com/docopt/docopt-go"
	"github.com/pkorotkov/logstick"
	"github.com/pkorotkov/safe"
)

var version = "0.1"

var usage = `diskloner is a console app for cloning disks and partitions.

Usage:
  diskloner status
  diskloner clone <disk-path> <image-path>...
  diskloner -h | --help
  diskloner -v | --version

Options:
  -h --help     Show this message.
  -v --version  Show version.`

var log = logstick.NewLogger()

func init() {
	log.AddWriter(os.Stderr)
}

func main() {
	defer safe.CatchExit()
	defer log.Close()
	if err := InitializeApp(); err != nil {
		log.Error("failed to initialize application: %s", err)
		safe.Exit(1)
	}
	args, _ := Parse(usage, nil, true, version, false)
	switch {
	case args["status"]:
		log.Info("fetching status...")
	case args["clone"]:
		// Don't allow go further if app isn't run under root.
		if !isRoot() {
			log.Error("application requires root privileges")
			safe.Exit(2)
		}
		dc, err := NewDiskCloner(args["<disk-path>"].(string))
		if err != nil {
			log.Error("failed to create cloner: %s", err)
			safe.Exit(3)
		}
		defer dc.Close()
		if err = dc.SetImages(args["<image-path>"].([]string)); err != nil {
			log.Error("failed to set images: %s", err)
			safe.Exit(4)
		}
		dc.Clone()
	default:
		log.Error("invalid set of arguments")
		safe.Exit(5)
	}
	safe.Exit(0)
}
