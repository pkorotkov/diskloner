package internal

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
)

type fsEntity int

var FSEntity struct {
	File, Directory fsEntity
} = struct {
	File, Directory fsEntity
}{1, 2}

func CreatePathFor(fse fsEntity, path string) (err error) {
	var d string
	switch fse {
	case FSEntity.File:
		d = filepath.Dir(path)
	case FSEntity.Directory:
		d = path
	}
	// TODO: Lift this parameter to function arguments.
	if err = os.MkdirAll(d, os.FileMode(0755)); err != nil {
		return
	}
	return
}

func GetUnixDomainSocketsInDirectory(dp string) (ss []string, err error) {
	var fis []os.FileInfo
	if fis, err = ioutil.ReadDir(dp); err != nil {
		return
	}
	for _, fi := range fis {
		if fi.Mode()&os.ModeSocket != 0 {
			ss = append(ss, filepath.Join(dp, fi.Name()))
		}
	}
	return
}

// GetUUID generates a random UUID according to RFC 4122.
func GetUUID() string {
	uuid, _ := ioutil.ReadFile(AppPath.UUIDFile)
	return string(bytes.TrimSpace(uuid))
}
