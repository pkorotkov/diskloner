package main

/*
#include <stdlib.h>
#include <stdint.h>
#include <sys/ioctl.h>
#include <linux/fs.h>

#if defined(__linux__) && defined(_IOR) && !defined(BLKGETSIZE64)
#define BLKGETSIZE64 _IOR(0x12, 114, size_t)
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
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// func ucharArrayToGoString(data *C.uchar) string {
// 	return C.GoString((*C.char)(unsafe.Pointer(data)))
// }

// func cintArrayToGoIntSlice(ca *C.int, l C.size_t) []int {
// 	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(ca)), Len: int(l), Cap: int(l)}
// 	cs := *(*[]C.int)(unsafe.Pointer(&hdr))
// 	gs := make([]int, len(cs))
// 	for i, v := range cs {
// 		gs[i] = int(v)
// 	}
// 	return gs
// }

// func ioctl(fd, cmd, ptr uintptr) error {
// 	_, _, e := unix.Syscall(unix.SYS_IOCTL, fd, cmd, ptr)
// 	if e != 0 {
// 		return e
// 	}
// 	return nil
// }

func getDiskProfile(disk *os.File) (int32, int32, int64) {
	c := int64(C.get_disk_capacity_in_bytes(C.int(disk.Fd())))
	pss := int32(C.get_disk_physical_sector_size(C.int(disk.Fd())))
	lss := int32(C.get_disk_logical_sector_size(C.int(disk.Fd())))
	return pss, lss, c
}

func getRHSValue(bs []byte) string {
	return string(bytes.TrimSpace(bytes.Split(bs, []byte{61})[1]))
}

func getDiskSNAndType(dp string) (sn, dt string) {
	var (
		err error
		out []byte
	)
	sn, dt = "???", "???"
	out, err = exec.Command("sh", "-c", fmt.Sprintf("udevadm info --query=property %s", dp)).Output()
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
	return
}
