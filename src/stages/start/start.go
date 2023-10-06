package start

import (
	_ "embed"
	"errors"

	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/manpage.man
	manPage string
	//go:embed data/usage.tmpl
	startUsage string
	hints      = []string{
		`have you tried looking at the help text with {{"pb --help"|cyan}}?`,
		`did you know pb has a {{"manpage" | magenta}}?`,
		`try writing more commands like {{"pb example" | cyan}}`,
		`{{"https://www.imdb.com/title/tt0103919/" | cyan | underline}}`,
	}
)

func Run(in *term.Input) error {
	if err := util.InstallManpage(in.DB, manPage); err != nil {
		return err
	} else if in.None() {
		return util.ErrorFmt(startUsage, in.Env.User)
	} else if in.HasFlags("h", "help") {
		return util.ErrorFmt(`Did you think it would be that {{"EASY"|bold|white|Red}}?`, in.Env.User)
	} else if in.HasArgs("help") {
		return util.ErrorFmt("Oh very clever {{.|bold|cyan}}! Trying the command was a good idea.", in.Env.User)
	} else if in.HasOpt("hint", "hints") {
		return util.DisplayHint(in, hints)
	} else if in.HasOpt("not", "easy") {
		return util.ErrorFmt(`What? Are you just typing in anything I say in {{"bold"|bold}}?`, nil)
	} else if in.HasArgs("example") {
		return util.ErrorFmt(`ah so I see you take {{"hints"|bold}}`, nil)
	} else if in.HasOpt("bold") {
		return util.ErrorFmt(`OH COME ON {{.|bold|cyan}}!`, in.Env.User)
	} else if in.HasOpt(in.Env.User) {
		check := in.DB.Get("candyman")
		if len(check) < 2 {
			term.Println(`What is this? {{"Candyman?"|bold}}`, nil)
		} else if len(check) >= 2 && len(check) < 5 {
			term.Println(`You could have summoned {{"bloody mary"|red}} by now.`, in.Env.User)
		} else if len(check) >= 5 {
			return util.ErrorFmt(`Oh good job {{.}}, you have arrived. Try to {{"swarm"|bold}} the candy.`, in.Env.User)
		}
		in.DB.Set("candyman", check+"1")
		return nil
	} else if in.HasOpt("candyman") {
		return errors.New("You went too far, you were on the right track")
	} else if in.HasOpt("candy") && in.HasOpt("swarm") {
		term.Println(`{{"Congrats!" | cyan}} you did it, you are now onto the second stage.`, nil)
		util.SetStage(in, "waitforinfo")
		return nil
	} else if in.HasOpt("candy") {
		return errors.New("What do you want to do to the candy?")
	}
	return errors.New("no idea what you are trying to do")
}
