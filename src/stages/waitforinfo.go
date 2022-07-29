package stages

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/term"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	//go:embed data/wait/usage.tmpl
	waitUsage string
	//go:embed data/wait/key.pem
	key []byte
)

type waitForInfoStage struct {
	in *term.Input
	db *pstore.DB
}

func newWaitForInfo(in *term.Input, db *pstore.DB) stage {
	return &waitForInfoStage{in: in, db: db}
}

func (stg *waitForInfoStage) run() {
	if stg.in.None() {
		term.Println(waitUsage, stg.in.Env.User)
	} else if stg.in.HasArgs("listen") {
		stg.sshListen()
	} else if stg.in.HasArgs("speak") {
		for {
			fmt.Println("cnVubmluZyBvbiBwb3J0IDIwMjIK")
			time.Sleep(time.Second)
		}
	}
}

func (stg *waitForInfoStage) sshListen() {
	config := &ssh.ServerConfig{NoClientAuth: true}
	pk, _ := ssh.ParsePrivateKey(key)
	config.AddHostKey(pk)
	listener, _ := net.Listen("tcp", "127.0.0.1:2022")

	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Oh that is nice")
		w.Write([]byte("hello friend"))
	}))

	for {
		nConn, _ := listener.Accept()
		_, chans, reqs, _ := ssh.NewServerConn(nConn, config)
		fmt.Println("oh that tickles!")
		go ssh.DiscardRequests(reqs)
		for newChannel := range chans {
			channel, requests, _ := newChannel.Accept()
			go func(in <-chan *ssh.Request) {
				for req := range in {
					req.Reply(true, nil)
				}
			}(requests)
			term := terminal.NewTerminal(channel, "> ")
			go func() {
				defer channel.Close()
				for {
					line, err := term.ReadLine()
					if err != nil {
						break
					}
					if strings.TrimSpace(line) == "exit" {
						channel.Close()
					}
					fmt.Println("read:", line)
					term.Write([]byte("Oh yeah I guess so\n"))
				}
			}()
		}
	}
}
