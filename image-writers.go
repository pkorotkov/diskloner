package main

import (
	. "fmt"
	"os"
)

type imageWriters struct {
	fileMode os.FileMode
	writers  []*imageWriter
}

func newImageWriters(fileMode os.FileMode) *imageWriters {
	return &imageWriters{fileMode: fileMode}
}

func (iws *imageWriters) Close() error {
	for _, iw := range iws.writers {
		iw.Close()
	}
	return nil
}

func (iws *imageWriters) AbortAll() {
	var err error
	for _, iw := range iws.writers {
		if err = iw.Abort(); err != nil {
			log.Error("failed to abort image: %s", err)
		}
	}
	return
}

func (iws *imageWriters) AddImageWriter(ip string, c int64) (err error) {
	iw, e := newImageWriter(ip, c, iws.fileMode)
	if e != nil {
		err = Errorf("failed to allocate image file %s: %s", ip, e)
		return
	}
	iws.writers = append(iws.writers, iw)
	return
}

func (iws *imageWriters) Write(bs []byte) (n int, err error) {
	for _, iw := range iws.writers {
		n, err = iw.Write(bs)
		if err != nil {
			iws.AbortAll()
			return
		}
	}
	return
}

func (iws *imageWriters) DumpReports(r []byte) {
	for _, iw := range iws.writers {
		// Create report file.
		iif, err := os.Create(iw.file.Name() + ".report")
		if err != nil {
			log.Error("failed to create report file: %s", err)
			continue
		}
		iif.Write(r)
		iif.Close()
	}
}
