package stages

import (
	_ "embed"
	"fmt"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/term"
)

type (
	stage        interface{ run() }
	stageFactory = func(*term.Input, *pstore.DB) stage
)

var (
	//go:embed data/cowsay.tmpl
	cow string
	//go:embed data/meow.tmpl
	meow     string
	handlers = map[string]stageFactory{
		"start":       newStartStage,
		"waitforinfo": newWaitForInfo,
	}
)

func addArtifact(db *pstore.DB, artf string) {
	db.Set("artifacts", db.Get("artifacts")+artf+"\n")
}

// Run will find the current stage and run it
func Run(in *term.Input, db *pstore.DB) {
	if in.HasFlags("artifacts") {
		fmt.Println(db.Get("artifacts"))
		return
	} else if in.HasFlags("reset") {
		db.Drop()
	} else if in.HasFlags("moo") {
		term.Println(cow, nil)
		return
	} else if in.HasFlags("meow") {
		term.Println(meow, nil)
		return
	}
	if stage, ok := handlers[db.Get("stage")]; ok {
		stage(in, db).run()
	} else {
		handlers["start"](in, db).run()
	}
}
