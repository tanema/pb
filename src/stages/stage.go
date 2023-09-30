package stages

import (
	_ "embed"

	"github.com/tanema/pb/src/artifacts"
	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/stages/lisp"
	"github.com/tanema/pb/src/stages/start"
	"github.com/tanema/pb/src/stages/wait"
	"github.com/tanema/pb/src/term"
)

type (
	Stage func(*term.Input, *pstore.DB) error
)

var (
	//go:embed default/cowsay.tmpl
	cow string
	//go:embed default/meow.tmpl
	meow string
	//go:embed default/milk.tmpl
	milk     string
	handlers = map[string]Stage{
		"start":       start.Run,
		"waitforinfo": wait.Run,
		"lisp":        lisp.Run,
	}
)

// Run will find the current stage and run it
func Run(in *term.Input, db *pstore.DB) error {
	artifacts.Setup(db)
	db.Set("hint", "are you trying to cheat by looking at the data?")
	if in.HasFlags("artifacts") {
		artifacts.Print(db)
	} else if in.HasFlags("reset") {
		for _, key := range db.Keys() {
			if key != "artifacts" {
				if err := db.Del(key); err != nil {
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
		if db.Get("stage") == "" {
			db.Set("stage", "start")
		}
		return handlers[db.Get("stage")](in, db)
	}
	return nil
}
