package main

import (
	"encoding/gob"
	. "fmt"
	"net"
	"os"

	. "./internal"

	uip "github.com/gosuri/uiprogress"
)

func monitorStatus(quit chan os.Signal) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{AppPath.ProgressFile, "unix"})
	if err != nil {
		log.Error("failed to establish monitoring connection: %s", err)
		return
	}
	defer os.Remove(AppPath.ProgressFile)
	messages := make(chan *ProgressMessage)
	sessions := make(map[string]*uip.Bar)
	uip.Start()
	for {
		go func(m chan *ProgressMessage) {
			var mp ProgressMessage
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
				bar, ok := sessions[m.UUID]
				if !ok {
					sessions[m.UUID] = uip.AddBar(10000)
					bar = sessions[m.UUID]
					bar.AppendCompleted()
					bar.PrependFunc(func(b *uip.Bar) string {
						return Sprintf("%s (%s...%s)", m.Name, m.UUID[0:6], m.UUID[32:36])
					})
				}
				bar.Set(m.Count)
				// TODO: Use sync.Pool.
				// TODO: Put the progress message object back into pool.
			}
		case <-quit:
			return
		}
	}
	return
}
