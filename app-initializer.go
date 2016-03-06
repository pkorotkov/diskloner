package main

import (
	. "fmt"
	"os"

	"./internal/definitions"
)

func InitializeApp() (err error) {
	if _, err = os.Stat(definitions.AppPath.Progress); os.IsNotExist(err) {
		// Create progress directory if it does not exist.
		if err = createParentDirectories(definitions.AppPath.Progress); err != nil {
			err = Errorf("failed to create progress directory: %s", err)
			return
		}
	}
	return nil
}
