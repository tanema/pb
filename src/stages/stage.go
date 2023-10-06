package stages

import (
	_ "embed"

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
	Stage func(*term.Input) error
)

var (
	//go:embed default/cowsay.tmpl
	cow string
	//go:embed default/meow.tmpl
	meow string
	//go:embed default/milk.tmpl
	milk     string
	handlers = map[string]Stage{
		"start":        start.Run,
		"waitforinfo":  wait.Run,
		"lisp":         lisp.Run,
		"merrygoround": merry.Run,
		"next":         next.Run,
	}
)

// Run will find the current stage and run it
func Run(in *term.Input) error {
	artifacts.Setup(in.DB)
	in.DB.Set("hint", "are you trying to cheat by looking at the data?")
	if in.HasFlags("artifacts") {
		artifacts.Print(in.DB)
	} else if in.HasFlags("reset") {
		for _, key := range in.DB.Keys() {
			if key != "artifacts" {
				if err := in.DB.Del(key); err != nil {
					return err
				}
			}
		}
	} else if in.HasOpt("moo", "cow") {
		term.Println(cow, nil)
	} else if in.HasOpt("meow", "cat", "kitty") {
		term.Println(meow, nil)
	} else if in.HasOpt("milk", "cheese") {
		term.Println(milk, nil)
	} else {
		if in.DB.Get("stage") == "" {
			util.SetStage(in, "start")
		}
		return handlers[in.DB.Get("stage")](in)
	}
	return nil
}
