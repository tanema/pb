package wait

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/tanema/pb/src/server"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

type (
	WaitStage struct {
		in         *term.Input
		man, usage string
		hints      []string
		options    map[string]string
	}
	fileItem struct {
		Owner string
		Name  string
		Day   int
		Month string
		Time  string
		Size  string
	}
)

const (
	port     = "2023"
	password = "hackerman"
)

var (
	portMsg   = util.Base64("My port is %v, call me!", port)
	passHex   = util.Hex(password)
	passwdMsg = fmt.Sprintf("the password is: %s", passHex)
	fileList  = `{{range . -}}
-rw-r--r--  1 {{.Owner | cyan}}  staff   {{.Size}} {{.Day}} {{.Month}} {{.Time}} {{.Name}}
{{end}}`
	fakefiles = []fileItem{
		{Owner: "timanema", Name: "readme.md", Day: 23, Month: "Sep", Time: "20:13", Size: "5mb"},
		{Owner: "root", Name: "note.txt", Day: 21, Month: "Sep", Time: "10:05", Size: "1mb"},
	}
)

func New(in *term.Input) *WaitStage {
	return &WaitStage{
		in: in,
		usage: `Good job! You made it to stage 2! Now what?,

In this stage we will communicate in many ways.

I will {{"listen"|bold}} to you, and if you want, I can {{"speak"|bold}} as well!`,
		man: "So you think you are clever now because you got to the second step right?",
		hints: []string{
			`{{"base64 -d" | cyan}} will be your friend.`,
			`you might need to use {{"ssh" | magenta}}.`,
			`do you know linux tools like {{"ls" | cyan}} and {{"cat" | cyan}}?`,
			`do you know what {{"SIGINFO" | yellow}} is?`,
		},
		options: map[string]string{
			"--listen": "Let me listen to what you have to say.",
			"--speak":  "You listen to what I have to say",
		},
	}
}

func (stage *WaitStage) Title() string              { return "A Conversation" }
func (stage *WaitStage) Man() string                { return stage.man }
func (stage *WaitStage) Help() string               { return stage.usage }
func (stage *WaitStage) Hints() []string            { return stage.hints }
func (stage *WaitStage) Options() map[string]string { return stage.options }

func (stage *WaitStage) Run() error {
	if stage.in.None() {
		return util.ErrorShowUsage
	} else if stage.in.HasOpt("listen") {
		return stage.listen()
	} else if stage.in.HasOpt("speak") {
		fmt.Print("I dont feel so good, I think I might puuu:")
		return term.Errorf("{{.|bold|green}}", portMsg)
	}
	return errors.New("no idea what you are trying to do")
}

func (stage *WaitStage) listen() error {
	srv := server.New()
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Oh that is nice, it's one way to connect with me. But sssshhh don't tell anyone")
		w.Write([]byte("Hello friend! I am afraid I prefer different communication styles."))
	})
	srv.HandleSSH("> ", `============================================
*          Puzzle Box OS 2.14.98           *
============================================
To authenitcate run the login command.

`, stage.handleSSH)

	go util.OnSignal(func(sig os.Signal) {
		fmt.Println("That was clever! This is a shortcut!")
		fmt.Println(passwdMsg)
	}, syscall.Signal(29))

	return srv.ListenAndServe("127.0.0.1:2023")
}

func (stage *WaitStage) handleSSH(sshTerm io.Writer, cmd string) error {
	cmdParts := strings.Split(strings.TrimSpace(cmd), " ")
	switch cmdParts[0] {
	case "exit":
		return errors.New("Goodbye")
	case "ls", "list", "dir", "ll", "la":
		term.Fprint(sshTerm, fileList, fakefiles)
	case "login":
		if len(cmdParts) == 1 {
			fmt.Fprintln(sshTerm, "Usage: login [password]")
		} else if cmdParts[1] == password {
			term.Println(`You have been {{"authenticated"|cyan}}. You are now on logged into {{"stage 3"|red}}`, nil)
			util.SetStage(stage.in, "lisp")
		} else if cmdParts[1] == passHex {
			fmt.Fprintln(sshTerm, "such a curse to be so close, you could say that this password is hexed")
		} else {
			fmt.Fprintln(sshTerm, "Incorrect password. This incident will be reported to the authorities.")
		}
	case "su", "sudo":
		fmt.Fprintln(sshTerm, "We are confident, aren't we?")
	case "cat":
		if len(cmdParts) == 1 {
			fmt.Fprintln(sshTerm, "huh?")
		} else if len(cmdParts) >= 1 && strings.ToLower(cmdParts[1]) == "readme.md" {
			fmt.Fprintf(sshTerm, "The password is %v\n", passHex)
		} else if len(cmdParts) >= 1 && strings.ToLower(cmdParts[1]) == "note.txt" {
			fmt.Fprintln(sshTerm, "not this file, the other.")
		} else if len(cmdParts) >= 1 {
			fmt.Fprintln(sshTerm, "you do not have permission to read that.")
		} else {
			fmt.Fprintln(sshTerm, "meow?")
		}
	case "rm":
		fmt.Fprintln(sshTerm, "What exactly are you trying to acheive?")
	case "help":
		fmt.Fprintln(sshTerm, "Try looking around.")
	case "look":
		fmt.Fprintln(sshTerm, "This is not monkey island, this is a computer. Have you tried looking for INFO")
	case "hello", "hi":
		fmt.Fprintln(sshTerm, "Yes, hello again.")
	case "2.14.98":
		fmt.Fprintln(sshTerm, "nope that is just a random number.")
	case "secret":
		fmt.Fprintln(sshTerm, "not like that.")
	default:
		fmt.Fprintf(sshTerm, "unknown command %v", cmdParts[0])
	}
	return nil
}
