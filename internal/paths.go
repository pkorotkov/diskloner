package internal

import "runtime"

var AppPath struct {
	ProgressDirectory string
	UUIDFile          string
}

func init() {
	switch runtime.GOOS {
	case "linux":
		AppPath.ProgressDirectory = "/var/run/diskloner/progress"
		AppPath.UUIDFile = "/proc/sys/kernel/random/uuid"
	}
}
