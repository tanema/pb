package wait

import (
	"bufio"
	"errors"
	"log"
	"net"
	"time"
)

type listener struct {
	accept chan net.Conn
	net.Listener
}

func newListener(l net.Listener) *listener {
	return &listener{
		make(chan net.Conn),
		l,
	}
}

func (l *listener) Accept() (net.Conn, error) {
	if l.accept == nil {
		return nil, errors.New("Listener closed")
	}
	return <-l.accept, nil
}

func (l *listener) Close() error {
	close(l.accept)
	l.accept = nil
	return nil
}

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

func MuxListener(l net.Listener) (ssh net.Listener, other net.Listener) {
	sshListener, otherListener := newListener(l), newListener(l)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error accepting conn:", err)
				continue
			}
			conn.SetReadDeadline(time.Now().Add(time.Second * 10))
			bconn := bufferedConn{conn, bufio.NewReaderSize(conn, 3)}
			p, err := bconn.Peek(3)
			conn.SetReadDeadline(time.Time{})
			if err != nil {
				log.Println("Error peeking into conn:", err)
				continue
			}
			prefix := string(p)
			selectedListener := otherListener
			if prefix == "SSH" {
				selectedListener = sshListener
			}
			if selectedListener.accept != nil {
				selectedListener.accept <- bconn
			}
		}
	}()
	return sshListener, otherListener
}
