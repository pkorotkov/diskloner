package main

import (
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
	messages := make(chan []byte)
	for {
		go func(m chan []byte) {
			mb := make([]byte, 16)
			conn, err := l.AcceptUnix()
			if err != nil {
				log.Error("got accept error: %s", err)
				m <- nil
				return
			}
			defer conn.Close()
			_, err = conn.Read(mb)
			if err != nil {
				log.Error("failed to read message: %s", err)
				m <- nil
				return
			}
			m <- mb
			return
		}(messages)
		select {
		case m := <-messages:
			if m != nil {
				log.Info("message %s", m)
			}
		case <-quit:
			return
		}
	}
	return
}
