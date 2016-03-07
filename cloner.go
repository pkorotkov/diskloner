package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	. "fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"./internal/definitions"

	"golang.org/x/sys/unix"
)

type DiskCloner struct {
	deviceType         string
	serialNumber       string
	physicalSectorSize int32
	logicalSectorSize  int32
	capacity           int64
	disk               *os.File
	imageWriters       []*imageWriter
	progressFile       *os.File
}

func NewDiskCloner(diskPath string) (cloner *DiskCloner, err error) {
	// Open disk to read.
	var disk *os.File
	if disk, err = os.OpenFile(diskPath, unix.O_RDONLY|unix.O_NONBLOCK, os.FileMode(0400)); err != nil {
		return
	}
	var ok bool
	if ok, err = isFileBlockDevice(disk); err != nil {
		return
	}
	if !ok {
		err = Errorf("given path does not point to block device")
		return
	}
	sn, dt, pss, lss, c := getDiskProfile(disk)
	// Create progress file.
	var pf *os.File
	pfp := filepath.Join(definitions.AppPath.Progress, Sprintf("%d", os.Getpid()))
	if pf, err = os.OpenFile(pfp, os.O_WRONLY|os.O_CREATE|os.O_SYNC, os.FileMode(0644)); err != nil {
		return
	}
	cloner = &DiskCloner{dt, sn, pss, lss, c, disk, nil, pf}
	return
}

func (dc *DiskCloner) Close() error {
	for _, iw := range dc.imageWriters {
		iw.Close()
	}
	pfp := dc.progressFile.Name()
	dc.progressFile.Close()
	os.Remove(pfp)
	return dc.disk.Close()
}

func (dc *DiskCloner) SetImages(ips []string) error {
	for _, ip := range ips {
		iw, err := newImageWriter(ip, dc.capacity)
		if err != nil {
			dc.imageWriters = nil
			return Errorf("failed to allocate image %s: %s", ip, err)
		}
		dc.imageWriters = append(dc.imageWriters, iw)
	}
	return nil
}

func (dc *DiskCloner) clone(progress chan float64) {
	lss := int(dc.logicalSectorSize)
	sector := make([]byte, lss)
	zeroSector := make([]byte, lss, lss)
	md5h, sha1h, sha256h, sha512h := md5.New(), sha1.New(), sha256.New(), sha512.New()
	hw := io.MultiWriter(md5h, sha1h, sha256h, sha512h)
	ts := time.Now()
	count, portion := 1.0, 0
	var unreadSectors []int64
	for {
		n, err := dc.disk.Read(sector)
		cp, _ := dc.disk.Seek(0, 1)
		if err != nil {
			if err != io.EOF {
				log.Warning("detected unreadable sector at offset %d", cp)
				unreadSectors = append(unreadSectors, cp)
				sector, n = zeroSector, lss
				// Jump to the next sector.
				dc.disk.Seek(int64(lss), 1)
			} else {
				// This is the check of final call with (0, io.EOF) result.
				if n == 0 {
					break
				}
			}
		}
		hw.Write(sector)
		for _, iw := range dc.imageWriters {
			if iw.Aborted() {
				continue
			}
			_, err = iw.Write(sector)
			if err != nil {
				log.Error("failed to write in %s: %s", iw.file.Name(), err)
				p := iw.file.Name()
				// Stop writing to this writer - abort it.
				if err = iw.Abort(); err != nil {
					log.Warning("failed to fully abort image writer %s", p)
				}
			}
		}
		// Report progress in form XXX.YY%.
		portion += lss
		if p := float64(portion) / float64(dc.capacity); p >= count*0.001 {
			count++
			progress <- 100.0 * p
		}
	}
	close(progress)
	cr := &cloningReport{
		StartTime: ts,
		EndTime:   time.Now(),
		DiskProfile: diskProfile{
			dc.deviceType,
			dc.serialNumber,
			dc.physicalSectorSize,
			dc.logicalSectorSize,
			dc.capacity,
		},
		Hashes: hashes{
			Sprintf("%x", md5h.Sum(nil)),
			Sprintf("%x", sha1h.Sum(nil)),
			Sprintf("%x", sha256h.Sum(nil)),
			Sprintf("%x", sha512h.Sum(nil)),
		},
		UnreadLogicalSectors: unreadSectors,
	}
	bs, _ := json.MarshalIndent(cr, "", "    ")
	for _, iw := range dc.imageWriters {
		// Create info file.
		iif, err := os.Create(iw.file.Name() + ".info")
		if err != nil {
			log.Error("failed to create info file %s", err)
			continue
		}
		iif.Write(bs)
		iif.Close()
	}
}

func (dc *DiskCloner) Clone() {
	done := make(chan struct{})
	progress := make(chan float64)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		dc.clone(progress)
	}()
	for p := range progress {
		dc.progressFile.WriteString(Sprintf("\r%.2f%%", p))
	}
	<-done
}
