package main

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"io"
	"os"
	"unsafe"
)

type rescueSectorReader struct {
	cfd        *C.FILE
	sectorSize int32
	zeroSector []byte
}

func newRescueSectorReader(file *os.File, ss int32) *rescueSectorReader {
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
		return int(ss), nil
	}
	return int(ss), nil
}
