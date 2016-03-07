package main

/*
#include <linux/fs.h>
#include <stdlib.h>
#include <stdint.h>
#include <sys/ioctl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#if defined(__linux__)
    #if defined(_IOR) && !defined(BLKGETSIZE64)
        #define BLKGETSIZE64 _IOR(0x12, 114, size_t)
    #endif
#endif

uint64_t
get_disk_capacity_in_bytes(int fd) {
    uint64_t fs;
    if (ioctl(fd, BLKGETSIZE64, &fs) < 0) {
        return 0;
    }
    return fs;
}

uint32_t
get_disk_logical_sector_size(int fd) {
    uint32_t lss;
    if (ioctl(fd, BLKSSZGET, &lss) < 0) {
        return 0;
    }
    return lss;
}

uint32_t
get_disk_physical_sector_size(int fd) {
    uint32_t pss;
    if (ioctl(fd, BLKBSZGET, &pss) < 0) {
        return 0;
    }
    return pss;
}

int
is_file_block_device(int fd, int *ecode) {
    struct stat sb;
    *ecode = 0;
    if (fstat(fd, &sb) != 0) {
        *ecode = 1;
        return 0;
    }
    return S_ISBLK(sb.st_mode);
}
*/
import "C"

import (
	"bytes"
	. "fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func isProcessConnectedToTerminal() bool {
	return C.isatty(1) == 1
}

func isRoot() bool {
	return C.geteuid() == 0
}

func executeShellCommand(c string) ([]byte, error) {
	return exec.Command("sh", "-c", c).Output()
}

func getRHSValue(bs []byte) string {
	return string(bytes.TrimSpace(bytes.Split(bs, []byte{61})[1]))
}

func getDiskProfile(disk *os.File) (sn, dt string, pss, lss int32, c int64) {
	var (
		err error
		out []byte
	)
	sn, dt = "???", "???"
	out, err = executeShellCommand(Sprintf("udevadm info --query=property %s", disk.Name()))
	if err != nil {
		return
	}
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		switch {
		case bytes.Contains(line, []byte("ID_SERIAL_SHORT=")):
			if lsn := getRHSValue(line); len(lsn) != 0 {
				sn = lsn
			}
		case bytes.Contains(line, []byte("DEVTYPE=")):
			if ldt := getRHSValue(line); len(ldt) != 0 {
				dt = ldt
			}
		}
	}
	fd := C.int(disk.Fd())
	pss = int32(C.get_disk_physical_sector_size(fd))
	lss = int32(C.get_disk_logical_sector_size(fd))
	c = int64(C.get_disk_capacity_in_bytes(fd))
	return
}

func isFileBlockDevice(f *os.File) (bool, error) {
	var ecode C.int
	r := C.is_file_block_device(C.int(f.Fd()), &ecode)
	if ecode != 0 {
		return false, Errorf("failed to check file type")
	}
	return r != 0, nil
}

func createParentDirectories(path string) error {
	d := filepath.Dir(path)
	if err := os.MkdirAll(d, os.ModeDir); err != nil {
		return err
	}
	// Set drwxr-xr-x permission mode.
	return os.Chmod(d, os.FileMode(0755))
}
