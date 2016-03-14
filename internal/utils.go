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

func CreateDirectoriesFor(fse fsEntity, path string) (err error) {
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

// GetUUID generates a random UUID according to RFC 4122.
func GetUUID() string {
	uuid, _ := ioutil.ReadFile(AppPath.UUIDFile)
	return string(bytes.TrimSpace(uuid))
}