package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"
	"sync"

	. "./internal"
)

var progressLineUpdatePool = sync.Pool{
	New: func() interface{} {
		return &ProgressLineUpdate{}
	},
}

func listenForMessage(l *net.UnixListener, messages chan Message) {
	gob.Register(&CloningMessage{})
	gob.Register(&InquiringMessage{})
	gob.Register(&CompletedMessage{})
	gob.Register(&AbortedMessage{})
	conn, err := l.AcceptUnix()
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

func monitorStatus(quit chan os.Signal) {
	// TODO: Make this for multiple monitors.
	l, err := net.ListenUnix("unix", &net.UnixAddr{AppPath.ProgressFile, "unix"})
	if err != nil {
		log.Error("failed to establish monitoring connection: %s", err)
		return
	}
	defer os.Remove(AppPath.ProgressFile)
	messages := make(chan Message)
	sessions := make(map[string]int)
	p := NewProgress()
	defer p.Close()
	for {
		go listenForMessage(l, messages)
		select {
		case message := <-messages:
			if message != nil {
				_, ok := sessions[message.UUID()]
				if !ok {
					sid := Sprintf("%s...%s", message.UUID()[0:6], message.UUID()[32:36])
					sessions[message.UUID()] = p.AddLine(NewProgressLine(sid))
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
				p.UpdateLine(l)
				progressLineUpdatePool.Put(l)
			}
		case <-quit:
			return
		}
	}
	return
}
