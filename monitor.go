package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"
	"sync"

	. "./internal"
)

var lineWithSIDAndStateUpdatePool = sync.Pool{
	New: func() interface{} {
		return &LineWithSIDAndStateUpdate{}
	},
}

func monitorStatus(quit chan os.Signal) {
	// TODO: Make this for multiple monitors.
	l, err := net.ListenUnix("unix", &net.UnixAddr{AppPath.ProgressFile, "unix"})
	if err != nil {
		log.Error("failed to establish monitoring connection: %s", err)
		return
	}
	defer os.Remove(AppPath.ProgressFile)
	messages := make(chan *CloningMessage)
	sessions := make(map[string]int)
	p := NewProgress()
	defer p.Close()
	for {
		go func(m chan *CloningMessage) {
			var mp CloningMessage
			conn, err := l.AcceptUnix()
			if err != nil {
				log.Error("got accept error: %s", err)
				m <- nil
				return
			}
			defer conn.Close()
			err = gob.NewDecoder(conn).Decode(&mp)
			if err != nil {
				log.Error("failed to read progress message: %s", err)
				m <- nil
				return
			}
			m <- &mp
			return
		}(messages)
		select {
		case m := <-messages:
			if m != nil {
				_, ok := sessions[m.UUID]
				if !ok {
					sid := Sprintf("%s...%s", m.UUID[0:6], m.UUID[32:36])
					sessions[m.UUID] = p.AddLine(NewLineWithSIDAndState(sid, "TODO", m.TotalBytes))
				}
				l := lineWithSIDAndStateUpdatePool.Get().(*LineWithSIDAndStateUpdate)
				l.Id, l.State, l.Current = sessions[m.UUID], "TODO", m.CopiedBytes
				p.UpdateLine(l)
				lineWithSIDAndStateUpdatePool.Put(l)
			}
		case <-quit:
			return
		}
	}
	return
}
