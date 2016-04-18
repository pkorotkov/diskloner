package main

import (
	. "fmt"

	. "github.com/pkorotkov/diskloner/internal"
)

func InitializeApp() (err error) {
	if err = CreatePathFor(FSEntity.Directory, AppPath.ProgressDirectory); err != nil {
		err = Errorf("failed to create progress directory: %s", err)
		return
	}
	return
}
