package server

import (
	"bufio"
	"net"
)

type bufferedConn struct {
	net.Conn
	r *bufio.Reader
}

func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b bufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}
