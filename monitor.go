package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "./internal"
)

var progressLineUpdatePool = sync.Pool{
	New: func() interface{} {
		return &ProgressLineUpdate{}
	},
}

func listenForMessage(ul *net.UnixListener, messages chan Message) {
	// Register all types of messages.
	gob.Register(&CloningMessage{})
	gob.Register(&InquiringMessage{})
	gob.Register(&CompletedMessage{})
	gob.Register(&AbortedMessage{})
	conn, err := ul.AcceptUnix()
	if err != nil {
		log.Error("got accept error: %s", err)
		messages <- nil
		return
	}
	defer conn.Close()
	var mp Message
	err = gob.NewDecoder(conn).Decode(&mp)
	if err != nil {
		log.Error("failed to read progress message: %s", err)
		messages <- nil
		return
	}
	messages <- mp
	return
}

func MonitorStatus(quit chan os.Signal) {
	// Create a UNIX domain socket server to wait connections from working cloners/inquirers.
	file := filepath.Join(AppPath.ProgressDirectory, Sprintf("%d", time.Now().UnixNano()))
	ul, err := net.ListenUnix("unix", &net.UnixAddr{file, "unix"})
	if err != nil {
		log.Error("failed to establish monitoring connection: %s", err)
		return
	}
	defer os.Remove(file)
	messages := make(chan Message)
	sessions := make(map[string]int)
	progress := NewProgress()
	defer progress.Close()
	for {
		go listenForMessage(ul, messages)
		select {
		case message := <-messages:
			if message != nil {
				_, ok := sessions[message.UUID()]
				if !ok {
					sid := Sprintf("%s...%s", message.UUID()[0:6], message.UUID()[32:36])
					sessions[message.UUID()] = progress.AddLine(NewProgressLine(sid))
				}
				l := progressLineUpdatePool.Get().(*ProgressLineUpdate)
				l.Id = sessions[message.UUID()]
				switch m := message.(type) {
				case *CloningMessage:
					l.State = "Cloning"
					l.Current = m.CopiedBytes
					l.Total = m.TotalBytes
				case *CompletedMessage:
					l.State = "Completed"
				case *AbortedMessage:
					l.State = "Aborted"
				}
				progress.UpdateLine(l)
				progressLineUpdatePool.Put(l)
			}
		case <-quit:
			return
		}
	}
	return
}
