package main

import (
	"encoding/gob"
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
					sessions[m.UUID] = uip.AddBar(10000).AppendCompleted()
					bar = sessions[m.UUID]
					bar.PrependFunc(func(bar *uip.Bar) string {
						return m.UUID[:9] + "..."
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
