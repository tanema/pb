package wait

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
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

	return srv.ListenAndServe("127.0.0.1:2023")
}

func sigInfo() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.Signal(29))
	for {
		<-c
		fmt.Println("That was clever! This is a shortcut!")
		fmt.Println(passwdMsg)
	}
}

func handleSSH(db *pstore.DB) func(sshTerm io.Writer, cmd string) error {
	return func(sshTerm io.Writer, cmd string) error {
		cmdParts := strings.Split(strings.TrimSpace(cmd), " ")
		switch cmdParts[0] {
		case "exit":
			return errors.New("Goodbye")
		case "ls", "list", "dir", "ll", "la":
			rnd, _ := term.Sprintf(fileList, fakefiles)
			sshTerm.Write([]byte(rnd))
		case "login":
			if len(cmdParts) == 1 {
				fmt.Fprintln(sshTerm, "Usage: login [password]")
			} else if cmdParts[1] == password {
				db.Set("stage", "lisp")
				fmt.Fprintln(sshTerm, "You have been authenticated. You are now on logged into stage 3")
				os.Exit(0)
			} else if cmdParts[1] == passHex {
				fmt.Fprint(sshTerm, "such a curse to be so close, you could say that this password is hexed")
			} else {
				fmt.Fprintln(sshTerm, "Incorrect password. This incident will be reported to the authorities.")
			}
		case "su", "sudo":
			fmt.Fprint(sshTerm, "We are confident, aren't we?\n")
		case "cat":
			if len(cmdParts) == 1 {
				fmt.Fprint(sshTerm, "huh?\n")
			} else if len(cmdParts) >= 1 && strings.ToLower(cmdParts[1]) == "readme.md" {
				fmt.Fprintf(sshTerm, "The password is %v\n", passHex)
			} else {
				fmt.Fprint(sshTerm, "meow?\n")
			}
		case "rm":
			fmt.Fprint(sshTerm, "What exactly are you trying to acheive?\n")
		case "help":
			fmt.Fprint(sshTerm, "I think you need a secret word. Try looking around\n")
		case "look":
			fmt.Fprint(sshTerm, "This is not monkey island, this is a computer. Have you tried looking for INFO\n")
		case "hello", "hi":
			fmt.Fprint(sshTerm, "Yes, hello again.\n")
		case "2.14.98":
			fmt.Fprint(sshTerm, "nope that is just a random number.\n")
		case "secret":
			fmt.Fprint(sshTerm, "not like that.\n")
		default:
			fmt.Fprint(sshTerm, "I have no idea what you are trying to say.\n")
		}
		return nil
	}
}
