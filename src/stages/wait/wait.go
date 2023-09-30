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

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/server"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

const port = "2023"

type fileItem struct {
	Owner string
	Name  string
	Day   int
	Month string
	Time  string
	Size  string
}

var (
	//go:embed data/usage.tmpl
	waitUsage string
	//go:embed data/manpage.man
	manPage string
	//go:embed data/file_list.tmpl
	fileList  string
	portMsg   = util.Base64("My port is %v, call me!", port)
	password  = "hackerman"
	passHex   = util.Hex(password)
	passwdMsg = fmt.Sprintf("the password is: %s", passHex)
	fakefiles = []fileItem{
		{Owner: "timanema", Name: "readme.md", Day: 23, Month: "Sep", Time: "20:13", Size: "50mb"},
	}
)

func Run(in *term.Input, db *pstore.DB) error {
	if err := util.InstallManpage(db, manPage); err != nil {
		return err
	} else if in.None() {
		return util.ErrorFmt(waitUsage, in.Env.User)
	} else if in.HasArgs("listen") {
		return listen(db)
	} else if in.HasArgs("speak") {
		fmt.Print("I dont feel so good, I think I might puuu:")
		return util.ErrorFmt("{{.|bold|green}}", portMsg)
	}
	return errors.New("no idea what you are trying to do")
}

func listen(db *pstore.DB) error {
	srv := server.New()
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Oh that is nice, it's one way to connect with me. But sssshhh don't tell anyone")
		w.Write([]byte("Hello friend! I am afraid I prefer different communication styles."))
	})
	srv.HandleSSH("> ", `============================================
*          Puzzle Box OS 2.14.98           *
============================================
To authenitcate run the login command.

`, handleSSH(db))

	go util.OnSignal(func(sig os.Signal) {
		fmt.Println("That was clever! This is a shortcut!")
		fmt.Println(passwdMsg)
	}, syscall.Signal(29))

	return srv.ListenAndServe("127.0.0.1:2023")
}

func handleSSH(db *pstore.DB) func(sshTerm io.Writer, cmd string) error {
	return func(sshTerm io.Writer, cmd string) error {
		cmdParts := strings.Split(strings.TrimSpace(cmd), " ")
		switch cmdParts[0] {
		case "exit":
			return errors.New("Goodbye")
		case "ls", "list", "dir", "ll", "la":
			util.WriteFmt(sshTerm, fileList, fakefiles)
		case "login":
			if len(cmdParts) == 1 {
				fmt.Fprintln(sshTerm, "Usage: login [password]")
			} else if cmdParts[1] == password {
				db.Set("stage", "lisp")
				fmt.Fprintln(sshTerm, "You have been authenticated. You are now on logged into stage 3")
				os.Exit(0)
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
}
