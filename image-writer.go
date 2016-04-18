package main

import (
	"os"
	"sync"

	. "github.com/pkorotkov/diskloner/internal"
)

type imageWriter struct {
	abortOnce sync.Once
	aborted   bool
	file      *os.File
}

func newImageWriter(ip string, c int64, m os.FileMode) (iw *imageWriter, err error) {
	if err = CreatePathFor(FSEntity.File, ip); err != nil {
		return
	}
	var file *os.File
	if file, err = os.OpenFile(ip, os.O_WRONLY|os.O_CREATE, m); err != nil {
		return
	}
	if err = file.Truncate(c); err != nil {
		_ = abort(file)
		return
	}
	iw = &imageWriter{file: file}
	return
}

func (iw *imageWriter) Close() error {
	if iw.aborted {
		return nil
	}
	return iw.file.Close()
}

func abort(file *os.File) error {
	p := file.Name()
	file.Close()
	return os.Remove(p)
}

func (iw *imageWriter) Abort() (err error) {
	iw.abortOnce.Do(func() {
		iw.aborted = true
		err = abort(iw.file)
		return
	})
	return
}

func (iw *imageWriter) Write(bs []byte) (int, error) {
	return iw.file.Write(bs)
}
