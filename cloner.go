package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/gob"
	"encoding/json"
	. "fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	. "./internal"

	"golang.org/x/sys/unix"
)

type CloningSession struct {
	name         string
	diskProfile  diskProfile
	disk         *os.File
	imageWriters []*imageWriter
}

func NewCloningSession(name, diskPath string, imagePaths []string) (cs *CloningSession, err error) {
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
	dt, sn, pss, lss, c := getDiskProfile(disk)
	var iws []*imageWriter
	for _, ip := range imagePaths {
		iw, e := newImageWriter(ip, c)
		if e != nil {
			err = Errorf("failed to allocate image file %s: %s", ip, e)
			return
		}
		iws = append(iws, iw)
	}
	cs = &CloningSession{name, diskProfile{dt, sn, pss, lss, c}, disk, iws}
	return
}

func (cs *CloningSession) Close() error {
	for _, iw := range cs.imageWriters {
		iw.Close()
	}
	return cs.disk.Close()
}

func (cs *CloningSession) readSectors(progress chan *ProgressMessage, reports chan *cloningReport, quit chan os.Signal) {
	defer close(progress)
	lss := int(cs.diskProfile.LogicalSectorSize)
	sector := make([]byte, lss)
	zeroSector := make([]byte, lss, lss)
	count, portion := 1.0, 0
	md5h, sha1h, sha256h, sha512h := md5.New(), sha1.New(), sha256.New(), sha512.New()
	hw := io.MultiWriter(md5h, sha1h, sha256h, sha512h)
	var unreadSectors []int64
	ts := time.Now()
	for {
		select {
		case <-quit:
			reports <- nil
			return
		default:
			n, err := cs.disk.Read(sector)
			cp, _ := cs.disk.Seek(0, 1)
			if err != nil {
				if err != io.EOF {
					log.Warning("detected unreadable sector at offset %d", cp)
					unreadSectors = append(unreadSectors, cp)
					sector, n = zeroSector, lss
					// Jump to the next sector.
					cs.disk.Seek(int64(lss), 1)
				} else {
					// This is the check of final call with (0, io.EOF) result.
					if n == 0 {
						reports <- &cloningReport{
							SessionName: cs.name,
							StartTime:   ts,
							EndTime:     time.Now(),
							DiskProfile: cs.diskProfile,
							Hashes: hashes{
								Sprintf("%x", md5h.Sum(nil)),
								Sprintf("%x", sha1h.Sum(nil)),
								Sprintf("%x", sha256h.Sum(nil)),
								Sprintf("%x", sha512h.Sum(nil)),
							},
							UnreadLogicalSectors: unreadSectors,
						}
						return
					}
				}
			}
			hw.Write(sector)
			for _, iw := range cs.imageWriters {
				if iw.Aborted() {
					continue
				}
				_, err = iw.Write(sector)
				if err != nil {
					log.Error("failed to write in %s: %s", iw.file.Name(), err)
					fp := iw.file.Name()
					// Stop writing to this writer - abort it.
					if err = iw.Abort(); err != nil {
						log.Warning("failed to fully abort image writer %s", fp)
					}
				}
			}
			// Report progress in form XXX.YYY%.
			portion += lss
			if p := float64(portion) / float64(cs.diskProfile.Capacity); p >= count*0.0001 {
				count++
				progress <- &ProgressMessage{cs.name, 100.0 * p}
			}
		}
	}
}

func (cs *CloningSession) clone(progress chan *ProgressMessage, quit chan os.Signal) {
	reports := make(chan *cloningReport)
	go cs.readSectors(progress, reports, quit)
	// Wait for readSectors signals to reports (value - when completed successfully, nil - otherwise).
	r := <-reports
	if r != nil {
		bs, _ := json.MarshalIndent(r, "", "    ")
		for _, iw := range cs.imageWriters {
			if iw.Aborted() {
				continue
			}
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
}

func (cs *CloningSession) Clone(quit chan os.Signal) {
	var (
		done     = make(chan struct{})
		progress = make(chan *ProgressMessage)
	)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		cs.clone(progress, quit)
	}()
	address := &net.UnixAddr{AppPath.Progress, "unix"}
	for pm := range progress {
		conn, err := net.DialUnix("unix", nil, address)
		if err != nil {
			// TODO: Clean it up somehow.
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				log.Warning("failed to establish monitoring connection: %s", err)
			}
			continue
		}
		err = gob.NewEncoder(conn).Encode(pm)
		if err != nil {
			log.Error("progress message not sent: %s", err)
		}
		conn.Close()
	}
	// Wait until clone function signals its completetion.
	<-done
	return
}
