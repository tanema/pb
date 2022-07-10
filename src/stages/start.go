package stages

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/term"
)

var (
	//go:embed data/pb.1
	manPage string
	//go:embed data/start/usage.tmpl
	startUsage string
)

type startStage struct {
	in *term.Input
	db *pstore.DB
}

func newStartStage(in *term.Input, db *pstore.DB) stage {
	return &startStage{in: in, db: db}
}

func (stg *startStage) run() {
	stg.installArtifacts()
	if stg.in.None() || stg.in.HasFlags("reset") {
		term.Println(startUsage, stg.in.Env.User)
	} else if stg.in.HasFlags("h", "help") {
		term.Println(`Did you think it would be that {{"EASY"|bold|white|Red}}?`, stg.in.Env.User)
	} else if stg.in.HasArgs("help") {
		term.Println("Oh very clever {{.|bold|cyan}}! Trying the command was a good idea.", stg.in.Env.User)
	} else if stg.in.HasFlags("not", "easy") || stg.in.HasArgs("not", "easy") {
		term.Println(`What? Are you just typing in anything I say in {{"bold"|bold}}?`, nil)
	} else if stg.in.HasFlags("bold") || stg.in.HasArgs("bold") {
		term.Println(`{{"OH COME ON!"|bold}}`, nil)
	} else if stg.in.HasArgs(stg.in.Env.User) {
		if check := stg.db.Get("candyman"); len(check) >= 2 {
			term.Println(`Oh good job {{.}}, you have arrived. Try the candy.`, stg.in.Env.User)
		} else {
			stg.db.Set("candyman", check+"1")
			term.Println(`What is this? Candyman?`, nil)
		}
	} else if stg.in.HasArgs("candy") && stg.in.HasArgs("eat") {
		stg.db.Set("stage", "waitforinfo")
		term.Println(`{{"Congrats!" | cyan}} you did it, you are now onto the second stage.`, nil)
	} else if stg.in.HasArgs("candy") || stg.in.HasFlags("candy") {
		term.Println("What do you want to do to the candy?", nil)
	}
}

func (stg *startStage) installArtifacts() {
	if stg.db.Get("stage") != "start" {
		addArtifact(stg.db, filepath.Join(os.Getenv("HOME"), ".config", "pb"))
		stg.db.Set("hint", "are you trying to cheat by looking at the data?")
		stg.db.Set("stage", "start")
	}
	if stg.db.Get("manpage") != "" {
		return
	}
	file, err := os.Create("/usr/local/share/man/man1/pb.1")
	addArtifact(stg.db, file.Name())
	if err == nil {
		defer file.Close()
		defer stg.db.Set("manpage", "yes")
		file.Write([]byte(manPage))
	}
}
