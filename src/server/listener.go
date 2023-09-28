package server

import (
	"errors"
	"net"
)

type Listener struct {
	accept chan net.Conn
	net.Listener
}

func newListener(l net.Listener) *Listener {
	return &Listener{make(chan net.Conn), l}
}

func (l *Listener) Accept() (net.Conn, error) {
	if l.accept == nil {
		return nil, errors.New("Listener closed")
	}
	return <-l.accept, nil
}

func (l *Listener) Close() error {
	close(l.accept)
	l.accept = nil
	return nil
}
