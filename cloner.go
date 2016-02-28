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
get_disk_size_in_bytes(int fd) {
	uint64_t fs;
	if (0 > ioctl(fd, BLKGETSIZE64, &fs)) {
		return 0;
	}
	return fs;
}
*/
import "C"

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const logicalSectorSize = 512

var (
	userRWFileMode = os.FileMode(0600)
	allRWFileMode  = os.FileMode(0666)
	allRWXFileMode = os.FileMode(0777)
)

var (
	md5HasherPool    = sync.Pool{New: func() interface{} { return md5.New() }}
	sha1HasherPool   = sync.Pool{New: func() interface{} { return sha1.New() }}
	sha256HasherPool = sync.Pool{New: func() interface{} { return sha256.New() }}
	sha512HasherPool = sync.Pool{New: func() interface{} { return sha512.New() }}
)

type DiskCloningReport struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	BenchPath    string    `json:"bench_path"`
	Type         string    `json:"type"`
	SerialNumber string    `json:"serial_number"`
	FullCapacity int64     `json:"full_capacity"`
	MD5Hash      string    `json:"md5_hash"`
	SHA1Hash     string    `json:"sha1_hash"`
	SHA256Hash   string    `json:"sha256_hash"`
	SHA512Hash   string    `json:"sha512_hash"`
}

type countWriter struct {
	progress       chan float64
	total, portion int64
	count          int
}

func (cw *countWriter) Close() error {
	close(cw.progress)
	return nil
}

// Write implements the io.Writer interface.
// It always completes with no error.
func (cw *countWriter) Write(p []byte) (int, error) {
	n := len(p)
	cw.portion += int64(n)
	if p := float64(cw.portion) / float64(cw.total); p >= float64(cw.count)*0.001 {
		cw.count++
		cw.progress <- 100.0 * p
	}
	return n, nil
}

func CloneDisk(diskPath, imagePath string) (dcr *DiskCloningReport, err error) {
	ts := time.Now()
	var disk *os.File
	if disk, err = os.OpenFile(diskPath, unix.O_RDONLY|unix.O_NONBLOCK, allRWXFileMode); err != nil {
		return
	}
	defer disk.Close()
	serialNumber, devType := getDiskSNAndType(diskPath)
	capacity := int64(C.get_disk_size_in_bytes(C.int(disk.Fd())))
	md5h := md5HasherPool.Get().(hash.Hash)
	sha1h := sha1HasherPool.Get().(hash.Hash)
	sha256h := sha256HasherPool.Get().(hash.Hash)
	sha512h := sha512HasherPool.Get().(hash.Hash)
	// Execute final assignment at the next to last deferring.
	defer func() {
		if err == nil {
			dcr = &DiskCloningReport{
				ts,
				time.Now(),
				diskPath,
				devType,
				serialNumber,
				capacity,
				fmt.Sprintf("%x", md5h.Sum(nil)),
				fmt.Sprintf("%x", sha1h.Sum(nil)),
				fmt.Sprintf("%x", sha256h.Sum(nil)),
				fmt.Sprintf("%x", sha512h.Sum(nil)),
			}
		}
		md5h.Reset()
		md5HasherPool.Put(md5h)
		sha1h.Reset()
		sha1HasherPool.Put(sha1h)
		sha256h.Reset()
		sha256HasherPool.Put(sha256h)
		sha512h.Reset()
		sha512HasherPool.Put(sha512h)
	}()
	// Create parent directories if some don't exist.
	d := filepath.Dir(imagePath)
	if err = os.MkdirAll(d, os.ModeDir); err != nil {
		return
	}
	os.Chmod(d, allRWXFileMode)
	var image *os.File
	if image, err = os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, userRWFileMode); err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = image.Close()
		} else {
			image.Close()
		}
	}()
	progress := make(chan float64)
	quit := make(chan struct{})
	go func() {
		// Create progress file.
		iifp := imagePath + ".progress"
		iif, err := os.Create(iifp)
		if err != nil {
			log.Println("failed to create progress file:", err)
			return
		}
		defer func() {
			iif.Close()
			// Try removing progress file.
			os.Remove(iifp)
			quit <- struct{}{}
		}()
		for p := range progress {
			// TODO: Set operation time out. Just in case.
			iif.WriteString(fmt.Sprintf("\r%.2f%%", p))
		}
	}()
	// TODO: Use sync.Pool.
	buf := make([]byte, logicalSectorSize)
	cw := &countWriter{progress, capacity, 0, 1}
	rr := newRescueSectorReader(disk, logicalSectorSize)
	defer rr.Close()
	// TODO: Remove this testing line someday.
	// _, err = io.Copy(image, io.TeeReader(disk, io.MultiWriter(md5h, sha1h, sha256h, sha512h, cw)))
	_, err = io.CopyBuffer(image, io.TeeReader(rr, io.MultiWriter(md5h, sha1h, sha256h, sha512h, cw)), buf)
	// Close count writer first right before return.
	cw.Close()
	// Wait to remove the progress file.
	<-quit
	return
}
