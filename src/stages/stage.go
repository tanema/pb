package stages

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/tanema/pb/src/artifacts"
	"github.com/tanema/pb/src/stages/lisp"
	"github.com/tanema/pb/src/stages/merry"
	"github.com/tanema/pb/src/stages/next"
	"github.com/tanema/pb/src/stages/start"
	"github.com/tanema/pb/src/stages/wait"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

type (
	Stage interface {
		Run() error
		Title() string
		Man() string
		Help() string
		Options() map[string]string
		Hints() []string
	}
)

var (
	//go:embed default/manpage.man.tmpl
	manPage string
	//go:embed default/usage.tmpl
	usage string
	//go:embed default/cowsay.tmpl
	cow string
	//go:embed default/meow.tmpl
	meow string
	//go:embed default/milk.tmpl
	milk string
)

// Run will find the current stage and run it
func Run(in *term.Input) error {
	handlers := map[string]Stage{
		"start":        start.New(in),
		"waitforinfo":  wait.New(in),
		"lisp":         lisp.New(in),
		"merrygoround": merry.New(in),
		"next":         next.New(in),
	}

	artifacts.Setup(in.DB)
	in.DB.Set("hint", "are you trying to cheat by looking at the data?")
	if in.HasFlags("artifacts") {
		artifacts.Print(in.DB)
		return nil
	} else if in.HasFlags("reset") {
		for _, key := range in.DB.Keys() {
			if key != "artifacts" {
				if err := in.DB.Del(key); err != nil {
					return err
				}
			}
		}
		return nil
	} else if in.HasOpt("moo", "cow") {
		return term.Println(cow, nil)
	} else if in.HasOpt("meow", "cat", "kitty") {
		return term.Println(meow, nil)
	} else if in.HasOpt("milk", "cheese") {
		return term.Println(milk, nil)
	}

	if in.DB.Get("stage") == "" {
		util.SetStage(in, "start")
	}
	currentStage := handlers[in.DB.Get("stage")]

	if err := installManpage(in, currentStage); err != nil {
		return err
	} else if in.HasFlags("help", "h") {
		return printUsage(in, currentStage)
	} else if in.HasFlags("hint") {
		return printHint(in, currentStage)
	} else if err := currentStage.Run(); err == util.ErrorShowUsage {
		return printUsage(in, currentStage)
	} else if err != nil {
		return err
	}
	return nil
}

func installManpage(in *term.Input, stage Stage) error {
	file, err := os.Create("/usr/local/share/man/man1/pb.1")
	if err != nil {
		return err
	} else if _, err = file.Write([]byte(term.Sprintf(manPage, stage))); err != nil {
		return err
	}
	artifacts.Add(in.DB, file.Name())
	return file.Close()
}

func printUsage(in *term.Input, stage Stage) error {
	return term.Println(usage, stage)
}

func printHint(in *term.Input, stage Stage) error {
	hints := stage.Hints()
	in.Hints = (in.Hints + 1) % len(hints)
	in.DB.Set("hints", fmt.Sprintf("%v", in.Hints))
	return errors.New(term.Sprintf(hints[in.Hints], nil))
}
