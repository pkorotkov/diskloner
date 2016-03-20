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

var (
	diskFileMode  = os.FileMode(0400)
	imageFileMode = os.FileMode(0600)
)

type CloningSession struct {
	name         string
	uuid         string
	diskProfile  diskProfile
	disk         *os.File
	imageWriters *imageWriters
}

func NewCloningSession(name, diskPath string, imagePaths []string) (cs *CloningSession, err error) {
	// Open disk to read.
	var disk *os.File
	if disk, err = os.OpenFile(diskPath, unix.O_RDONLY|unix.O_NONBLOCK, diskFileMode); err != nil {
		return
	}
	var ok bool
	if ok, err = IsFileBlockDevice(disk); err != nil {
		return
	}
	if !ok {
		err = Errorf("given path does not point to block device")
		return
	}
	// TODO: Rename to GetDiskInfo.
	dt, ptt, m, sn, pss, lss, c := GetDiskInfo(disk)
	iws := newImageWriters(imageFileMode)
	for _, ip := range imagePaths {
		if err = iws.AddImageWriter(ip, c); err != nil {
			return
		}
	}
	// TODO: Rework it to creating folder and move to the global app init function.
	if err = CreateDirectoriesFor(FSEntity.File, AppPath.ProgressFile); err != nil {
		err = Errorf("failed to create directory for progress file: %s", err)
		return
	}
	cs = &CloningSession{name, GetUUID(), diskProfile{dt, ptt, m, sn, pss, lss, c}, disk, iws}
	return
}

func (cs *CloningSession) Close() error {
	cs.imageWriters.Close()
	return cs.disk.Close()
}

func (cs *CloningSession) copySectors(progress chan Message, reports chan *CloningReport, quit chan os.Signal) {
	defer close(progress)
	md5h, sha1h, sha256h, sha512h := md5.New(), sha1.New(), sha256.New(), sha512.New()
	hw := io.MultiWriter(md5h, sha1h, sha256h, sha512h, cs.imageWriters)
	var (
		unreadSectors []int64
		count         = 1.0
		portion       = int64(0)
		sector        = make([]byte, cs.diskProfile.LogicalSectorSize)
		zeroSector    = make([]byte, cs.diskProfile.LogicalSectorSize, cs.diskProfile.LogicalSectorSize)
		ts            = time.Now()
	)
	for {
		select {
		case <-quit:
			reports <- nil
			progress <- &AbortedMessage{cs.uuid}
			return
		default:
			n, err := cs.disk.Read(sector)
			cp, _ := cs.disk.Seek(0, 1)
			if err != nil {
				if err != io.EOF {
					log.Warning("detected unreadable sector at offset %d", cp)
					unreadSectors = append(unreadSectors, cp)
					sector, n = zeroSector, cs.diskProfile.LogicalSectorSize
					// Jump to the next sector.
					cs.disk.Seek(int64(cs.diskProfile.LogicalSectorSize), 1)
				} else {
					// This is the check of final call with (0, io.EOF) result.
					if n == 0 {
						reports <- &CloningReport{
							Name:        cs.name,
							UUID:        cs.uuid,
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
						progress <- &CompletedMessage{cs.uuid}
						return
					}
				}
			}
			// Write sector to underlying writers.
			// It discontinues in case of any writing error
			// with destroying all image files.
			if _, err = hw.Write(sector); err != nil {
				reports <- nil
				progress <- &AbortedMessage{cs.uuid}
				return
			}
			// Report progress.
			portion += int64(cs.diskProfile.LogicalSectorSize)
			if p := float64(portion) / float64(cs.diskProfile.Capacity); p >= count*0.0001 {
				count++
				progress <- &CloningMessage{
					cs.uuid,
					int64(portion),
					cs.diskProfile.Capacity,
				}
			}
		}
	}
}

func (cs *CloningSession) clone(progress chan Message, quit chan os.Signal) error {
	reports := make(chan *CloningReport)
	go cs.copySectors(progress, reports, quit)
	// Wait for copySectors signals to reports (value - when completed successfully, nil - otherwise).
	if r := <-reports; r != nil {
		rbs, _ := json.MarshalIndent(r, "", "    ")
		cs.imageWriters.DumpReports(rbs)
	}
	return nil
}

func (cs *CloningSession) Clone(quit chan os.Signal) error {
	var (
		done     = make(chan error)
		progress = make(chan Message)
	)
	go func() {
		done <- cs.clone(progress, quit)
	}()
	address := &net.UnixAddr{AppPath.ProgressFile, "unix"}
	gob.Register(&CloningMessage{})
	gob.Register(&InquiringMessage{})
	gob.Register(&CompletedMessage{})
	gob.Register(&AbortedMessage{})
	for pm := range progress {
		conn, err := net.DialUnix("unix", nil, address)
		if err != nil {
			// TODO: Clean it up somehow.
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				log.Warning("failed to establish monitoring connection: %s", err)
			}
			continue
		}
		err = gob.NewEncoder(conn).Encode(&pm)
		if err != nil {
			log.Error("progress message not sent: %s", err)
		}
		conn.Close()
	}
	// Wait until clone function signals its completetion.
	return <-done
}
