package start

import (
	_ "embed"
	"errors"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/manpage.man
	manPage string
	//go:embed data/usage.tmpl
	startUsage string
)

func Run(in *term.Input, db *pstore.DB) error {
	if err := util.InstallManpage(db, manPage); err != nil {
		return err
	} else if in.None() {
		return util.ErrorFmt(startUsage, in.Env.User)
	} else if in.HasFlags("h", "help") {
		return util.ErrorFmt(`Did you think it would be that {{"EASY"|bold|white|Red}}?`, in.Env.User)
	} else if in.HasArgs("help") {
		return util.ErrorFmt("Oh very clever {{.|bold|cyan}}! Trying the command was a good idea.", in.Env.User)
	} else if in.HasOpt("not", "easy") {
		return util.ErrorFmt(`What? Are you just typing in anything I say in {{"bold"|bold}}?`, nil)
	} else if in.HasOpt("bold") {
		return util.ErrorFmt(`OH COME ON {{.|bold|cyan}}!`, in.Env.User)
	} else if in.HasOpt(in.Env.User) {
		check := db.Get("candyman")
		if len(check) >= 2 {
			return util.ErrorFmt(`Oh good job {{.}}, you have arrived. Try to {{"eat"|bold}} the candy.`, in.Env.User)
		}
		db.Set("candyman", check+"1")
		return util.ErrorFmt(`What is this? {{"Candyman?"|bold}}`, nil)
	} else if in.HasOpt("candyman") {
		return errors.New("You went too far, you were on the right track")
	} else if in.HasOpt("candy") && in.HasOpt("eat") {
		db.Set("stage", "waitforinfo")
		return util.ErrorFmt(`{{"Congrats!" | cyan}} you did it, you are now onto the second stage.`, nil)
	} else if in.HasOpt("candy") {
		return errors.New("What do you want to do to the candy?")
	}
	return errors.New("no idea what you are trying to do")
}
