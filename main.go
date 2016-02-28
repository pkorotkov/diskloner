package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

var (
	dp = flag.String("dp", "", "path to device (e.g. /dev/sda)")
	ip = flag.String("ip", "", "path to DD image file")
)

func init() {
	log.SetPrefix("diskloner: ")
	// Clear flags.
	log.SetFlags(0)
}

func main() {
	flag.Parse()
	if len(*dp) == 0 || len(*ip) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	cr, err := CloneDisk(*dp, *ip)
	if err != nil {
		log.Println("failed to perform cloning:", err)
		os.Exit(2)
	}
	bs, err := json.MarshalIndent(cr, "", "    ")
	if err != nil {
		log.Println("failed to marshal cloning report:", err)
		os.Exit(3)
	}
	// Create info file.
	iif, err := os.Create(*ip + ".info")
	if err != nil {
		os.Exit(4)
	}
	defer iif.Close()
	iif.Write(bs)
}
