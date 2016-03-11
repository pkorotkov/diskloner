package main

import (
	"encoding/gob"
	"net"
	"os"

	. "./internal"
)

type message struct {
	sessionName string
	body        []byte
}

func monitorStatus(quit chan os.Signal) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{AppPath.Progress, "unix"})
	if err != nil {
		log.Error("failed to establish monitoring connection: %s", err)
		return
	}
	defer os.Remove(AppPath.Progress)
	messages := make(chan *ProgressMessage)
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
				// TODO: Change output format.
				log.Info("message %v", m)
			}
		case <-quit:
			return
		}
	}
	return
}
