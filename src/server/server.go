package server

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//go:embed data/key.pem
var key []byte

type Server struct {
	mux      *http.ServeMux
	host     string
	listener net.Listener
	http     net.Listener
	closed   chan error

	ssh        net.Listener
	sshPrompt  string
	sshBanner  string
	sshHandler func(io.Writer, string) error
}

func New() *Server {
	return &Server{
		mux:    http.NewServeMux(),
		closed: make(chan error),
	}
}

func (server *Server) ListenAndServe(host string) error {
	var err error
	server.listener, err = net.Listen("tcp", host)
	if err != nil {
		return err
	}

	server.host = host
	server.http = newListener(server.listener)
	server.ssh = newListener(server.listener)

	go server.atc()
	go http.Serve(server.http, server.mux)
	go server.serveSSH()

	return <-server.closed
}

func (server *Server) Handle(pattern string, handler http.Handler) {
	server.mux.Handle(pattern, handler)
}

func (server *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	server.mux.HandleFunc(pattern, handler)
}

func (server *Server) HandleSSH(prompt, banner string, handler func(io.Writer, string) error) {
	server.sshPrompt = prompt
	server.sshBanner = banner
	server.sshHandler = handler
}

func (server *Server) atc() {
	for {
		conn, err := server.listener.Accept()
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
		selectedListener := server.http.(*Listener)
		if prefix == "SSH" {
			selectedListener = server.ssh.(*Listener)
		}
		if selectedListener.accept != nil {
			selectedListener.accept <- bconn
		}
	}
}

func (server *Server) serveSSH() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	pk, _ := ssh.ParsePrivateKey(key)
	config.AddHostKey(pk)

	for {
		nConn, err := server.ssh.Accept()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			fmt.Println("new conn error", err)
			continue
		}
		go ssh.DiscardRequests(reqs)
		go server.handleSSHConn(conn, chans)
	}

}

func (server *Server) handleSSHConn(conn *ssh.ServerConn, chans <-chan ssh.NewChannel) {
	defer conn.Close()
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		ch, _, err := newChan.Accept()
		if err != nil {
			continue
		}
		defer ch.Close()

		sshTerm := terminal.NewTerminal(ch, server.prompt())
		fmt.Fprint(sshTerm, server.sshBanner)
		for {
			sshTerm.SetPrompt(server.prompt())
			line, err := sshTerm.ReadLine()
			if err != nil {
				continue
			}
			if err := server.sshHandler(sshTerm, line); err != nil {
				return
			}
		}
	}
}

func (server *Server) prompt() string {
	if server.sshPrompt == "" {
		return "> "
	}
	return server.sshPrompt
}

func (server *Server) Close() error {
	if err := server.ssh.Close(); err != nil {
		return err
	} else if err := server.http.Close(); err != nil {
		return err
	} else if err := server.listener.Close(); err != nil {
		return err
	}
	close(server.closed)
	return nil
}
