package main

import (
	"os"
)

type imageWriter struct {
	aborted bool
	file    *os.File
}

func newImageWriter(ip string, c int64) (*imageWriter, error) {
	var err error
	if err = createParentDirectories(ip); err != nil {
		return nil, err
	}
	var file *os.File
	if file, err = os.OpenFile(ip, os.O_WRONLY|os.O_CREATE, os.FileMode(0600)); err != nil {
		return nil, err
	}
	if err = file.Truncate(c); err != nil {
		_ = abort(file)
		return nil, err
	}
	return &imageWriter{false, file}, nil
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

func (iw *imageWriter) Abort() error {
	iw.aborted = true
	return abort(iw.file)
}

func (iw *imageWriter) Aborted() bool {
	return iw.aborted
}

func (iw *imageWriter) Write(bs []byte) (int, error) {
	// if iw.aborted {
	// 	return 0, imageWriterAbortedError(Sprintf("image writer %s is aborted", iw.path))
	// }
	return iw.file.Write(bs)
}
