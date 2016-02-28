package main

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"unsafe"
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
			if lsn := string(bytes.TrimSpace(bytes.Split(line, []byte{61})[1])); len(lsn) != 0 {
				sn = lsn
			}
		case bytes.Contains(line, []byte("DEVTYPE=")):
			if ldt := string(bytes.TrimSpace(bytes.Split(line, []byte{61})[1])); len(ldt) != 0 {
				dt = ldt
			}
		}
	}
	return
}

type rescueSectorReader struct {
	cfd        *C.FILE
	sectorSize int
	zeroSector []byte
}

func newRescueSectorReader(file *os.File, ss int) *rescueSectorReader {
	mode := C.CString("rb")
	defer C.free(unsafe.Pointer(mode))
	return &rescueSectorReader{C.fdopen((C.int)(file.Fd()), mode), ss, make([]byte, ss, ss)}
}

func (rsr *rescueSectorReader) Close() error {
	C.fclose(rsr.cfd)
	return nil
}

// TODO: Rewrite in pure Go.
func (rsr *rescueSectorReader) Read(b []byte) (int, error) {
	ss := rsr.sectorSize
	r := C.fread(unsafe.Pointer(&b[0]), (C.size_t)(ss), 1, rsr.cfd)
	if ir := int(r); ir != 1 {
		if C.feof(rsr.cfd) != 0 {
			return ir, io.EOF
		}
		b = rsr.zeroSector
		C.fseek(rsr.cfd, (C.long)(ss), C.SEEK_CUR)
		return ss, nil
	}
	return ss, nil
}
