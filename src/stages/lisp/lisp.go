package lisp

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/chzyer/readline"

	"github.com/tanema/pb/src/lisp"
	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/usage.tmpl
	usage string
	//go:embed data/manpage.man
	manPage string
)

func Run(in *term.Input, db *pstore.DB) error {
	if err := util.InstallManpage(db, manPage); err != nil {
		return err
	} else if in.HasOpt("help", "h") {
		return util.ErrorFmt(usage, in.Env.User)
	}
	return repl()
}

func repl() error {
	fmt.Fprintln(os.Stderr, `This is a terrible implementation of ANSI Common Lisp with little
to no functionality.

It is free software, provided as is, with absolutely no warranty,
and no guarantees. Good luck, god speed.`)

	rl, err := readline.New("> ")
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	twice := 0
	for {
		text, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt && twice < 1 {
				buf.Reset()
				fmt.Fprintln(os.Stderr, "Press ctrl-c twice to exit.")
				twice++
			} else if err == readline.ErrInterrupt && twice >= 1 {
				break
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}
		twice = 0
		buf.WriteString(text + " ")
		val, err := lisp.Eval(buf.String())
		if err == lisp.ErrorUnderflow {
			rl.SetPrompt("...> ")
		} else {
			rl.SetPrompt("> ")
			buf.Reset()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				fmt.Fprintln(os.Stdout, val)
			}
		}
	}
	return nil
}
