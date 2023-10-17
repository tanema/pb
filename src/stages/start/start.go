package start

import (
	"errors"

	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

type StartStage struct {
	in                *term.Input
	title, man, usage string
	hints             []string
	options           map[string]string
}

func New(in *term.Input) *StartStage {
	return &StartStage{
		in:    in,
		title: "Let's go to the movies.",
		man:   "Ah so you know unix! Very clever. I wonder what you will find here. This may or may not change.",
		usage: "Your job, is to be a detective and figure out how to open me. There will be several stages to get through and solve, and eventually I will get sick of you and tell you that you completed it. I will not make it easy though. There may be a way that you can find more help on how to do this.",
		hints: []string{
			`have you tried looking at the help text with {{"pb --help"|cyan}}?`,
			`did you know pb has a {{"manpage" | magenta}}?`,
			`try writing more commands like {{"pb example" | cyan}}`,
			`{{"https://www.imdb.com/title/tt0103919/" | cyan | underline}}`,
		},
		options: map[string]string{
			"--candy": "Every one needs a little sweetness in their life",
		},
	}
}

func (stage *StartStage) Title() string              { return stage.title }
func (stage *StartStage) Man() string                { return stage.man }
func (stage *StartStage) Help() string               { return term.Sprintf(stage.usage, stage.in.Env.User) }
func (stage *StartStage) Hints() []string            { return stage.hints }
func (stage *StartStage) Options() map[string]string { return stage.options }

func (stage *StartStage) Run() error {
	if stage.in.None() {
		return util.ErrorShowUsage
	} else if stage.in.HasArgs("help") {
		return term.Errorf(`Oh very clever! Trying the command was a good idea. but it will {{"not"|red}} be that {{"easy"|bold}}`, nil)
	} else if stage.in.HasOpt("not", "easy") {
		return term.Errorf(`What? Are you just typing in anything I say in {{"bold"|bold}} {{.|bold}}?`, stage.in.Env.User)
	} else if stage.in.HasArgs("example") {
		return term.Errorf(`ah so I see you take {{"hints"|bold}} {{.|bold}}`, stage.in.Env.User)
	} else if stage.in.HasOpt("bold") {
		return term.Errorf(`OH COME ON {{.|bold|cyan}}!`, stage.in.Env.User)
	} else if stage.in.HasOpt(stage.in.Env.User) {
		check := stage.in.DB.Get("candyman")
		if len(check) == 0 {
			term.Println(`What is this? {{"Candyman?"|bold}}`, nil)
		} else if len(check) == 1 {
			term.Println(`Yes great you can say your own name twice.`, nil)
		} else if len(check) == 2 {
			term.Println(`This might be doing something? Do you think?`, nil)
		} else if len(check) == 3 {
			term.Println(`You could have summoned {{"bloody mary"|red}} by now.`, stage.in.Env.User)
		} else if len(check) == 4 {
			term.Println(`{{"bloody mary"|red}} is behind you!`, stage.in.Env.User)
		} else if len(check) >= 5 {
			return term.Errorf(`Oh good job {{.}}, you have arrived. Try to {{"--swarm"|bold}} the candy.`, stage.in.Env.User)
		}
		stage.in.DB.Set("candyman", check+"1")
		return nil
	} else if stage.in.HasOpt("candyman") {
		return errors.New("You went too far, you were on the right track")
	} else if stage.in.HasOpt("candy") && stage.in.HasOpt("swarm") {
		term.Println(`{{"Congrats!" | cyan}} you did it, you are now onto the second stage.`, nil)
		util.SetStage(stage.in, "waitforinfo")
		return nil
	} else if stage.in.HasOpt("candy") {
		return errors.New("What do you want to do to the candy?")
	}
	return errors.New("no idea what you are trying to do")
}
