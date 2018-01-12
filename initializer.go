package main

import (
	. "fmt"

	. "github.com/pkorotkov/diskloner/internal"
)

type ErrProgressDirectory string

func (e ErrProgressDirectory) Error() string {
	return string(e)
}

func InitializeApp() (err error) {
	if err = CreatePathFor(FSEntity.Directory, AppPath.ProgressDirectory); err != nil {
		err = ErrProgressDirectory(Sprintf("failed to create progress directory: %s", err))
		return
	}
	return
}
