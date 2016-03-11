package internal

import "runtime"

var AppPath struct {
	ProgressFile string
	UUIDFile     string
}

func init() {
	switch runtime.GOOS {
	case "linux":
		AppPath.ProgressFile = "/var/run/diskloner/progress"
		AppPath.UUIDFile = "/proc/sys/kernel/random/uuid"
	}
}
