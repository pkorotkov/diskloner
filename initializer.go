package main

import (
	. "fmt"

	. "./internal"
)

func InitializeApp() (err error) {
	if err = CreatePathFor(FSEntity.Directory, AppPath.ProgressDirectory); err != nil {
		err = Errorf("failed to create progress directory: %s", err)
		return
	}
	return
}
