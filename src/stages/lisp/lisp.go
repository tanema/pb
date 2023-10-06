package lisp

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/exp/slices"

	"github.com/tanema/pb/src/lisp"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/usage.tmpl
	usage string
	//go:embed data/manpage.man
	manPage    string
	order      = []string{"blue", "green", "yellow", "red"}
	pinNumber  = "4921"
	pinNumbers = strings.Split(pinNumber, "")
	touched    = 0
	hints      = []string{
		`it's lisp, don't think too hard but think with prefixes`,
		`check out the {{"(help)" | cyan}} output`,
		`how could you combine {{"touch" | cyan}} calls into a single line of code?`,
		`what does {{"touch" | cyan}} output? Is it the same every time?`,
		`{{"(print (str 4) (str (+ 1 1)))" | cyan}}`,
	}
)

func Run(in *term.Input) error {
	puzzleEnv := lisp.NewEnv(map[string]any{
		"help":   help,
		"look":   look,
		"touch":  touch,
		"unlock": unlock(in),
	})

	if err := util.InstallManpage(in.DB, manPage); err != nil {
		return err
	} else if in.HasOpt("help", "h") {
		return util.ErrorFmt(usage, in.Env.User)
	} else if in.HasOpt("hint") {
		return util.DisplayHint(in, hints)

	} else if in.HasPipe {
		return evalSrc(puzzleEnv, string(in.Stdin))
	} else if len(in.Args) > 0 {
		file, err := os.Open(in.Args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		src, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		return evalSrc(puzzleEnv, string(src))
	}
	return repl(puzzleEnv)
}

func evalSrc(env map[string]any, src string) error {
	_, err := lisp.EvalSrc(env, src)
	return err
}

func repl(env map[string]any) error {
	term.Println(`This is a terrible implementation of {{"ANSI Common Lisp"|bold}} with little
to no functionality.

For more information, you can use the {{"(help)"|cyan}} function and see documentation on
defined functions with {{"(doc [fname])"|cyan}}.

It is free software, provided as is, with absolutely no warranty,
and no guarantees. Good luck, god speed.`, nil)

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
				rl.SetPrompt("> ")
				term.Println(`Press {{"ctrl-c"|cyan}} twice to exit.`, nil)
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
		val, err := lisp.EvalSrc(env, buf.String())
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
		if touched > 0 {
			term.Println(`you hear a loud {{"ka-thunk"| red}}! something fell back into place.`, nil)
			touched = 0
		}
	}
	return nil
}

func help(env map[string]any, args []any) (any, error) {
	return term.Sprintf(`This is a limited implementation of lisp. You are able to explore more
functionality a few ways.

{{"env"|bold}}:    Use the env function to see all of the defined symbols within
        the current environment. This is good for finding what functionalities
        that you have access to. usage: {{"(env)"|cyan}}

{{"doc"|bold}}:    Use the doc function to see per-function documentation, to see
        usage and what they do. usage: {{"(doc funcName)"|cyan}}

{{"unlock"|bold}}: This is your target. This is the function that you need to unlock the
        next stage of the puzzle box. usage: {{"(unlock pinNumber)"|cyan}}

Some other funcs you might want to look at are {{"look"|bold}} and {{"touch"|bold}}`, nil)
}

func look(env map[string]any, args []any) (any, error) {
	if lisp.IsDocCall(env, args) {
		return `look will allow you to look around the puzzle environment.

Usage: (look "direction")`, nil
	} else if len(args) == 0 {
		return `you're in a dark room with 4 light sources on each side of you.`, nil
	} else if len(args) >= 1 {
		val, err := lisp.EvalForm(env, args[0])
		if err != nil {
			return nil, err
		}
		dir, ok := val.(string)
		if !ok {
			return nil, errors.New("I was expecting a string, cannot look in a direction that doesnt make sense")
		}
		switch dir {
		case "north", "front", "forward":
			return "to the north, you see a glowing red button", nil
		case "south", "back", "behind":
			return "to the north, you see a glowing blue button", nil
		case "east", "right":
			return "to the north, you see a glowing yellow button, there is a small message beside the button that says 'look up'", nil
		case "west", "left":
			return "to the north, you see a glowing green button", nil
		case "up", "above":
			return "you see a message scrawled the ceiling saying 'you need to touch the buttons in order and all at once!'", nil
		case "down", "below":
			return "you see wet dirty floor, however you see a message scratched into the dirt saying 'if you don`t touch them all at once, it will reset!'", nil
		default:
			return nil, errors.New("I don't know that direction")
		}
	}
	return nil, nil
}

func touch(env map[string]any, args []any) (any, error) {
	if lisp.IsDocCall(env, args) {
		return `touch will allow you to touch an item around you.`, nil
	} else if len(args) == 0 {
		return `you reach your hand out, touching at nothing`, nil
	} else if len(args) >= 1 {
		val, err := lisp.EvalForm(env, args[0])
		if err != nil {
			return nil, err
		}

		color, ok := val.(string)
		if !ok {
			return nil, errors.New("I was expecting a string, cannot touch something that doesnt make sense")
		}
		color = strings.Split(color, " ")[0]
		if slices.Contains(order, color) && order[touched] == color {
			touched++
			return pinNumbers[touched-1], nil
		} else if slices.Contains(order, color) {
			return "x", nil
		} else {
			return nil, errors.New("I don't know which button you are talking about")
		}
	}
	return nil, nil
}

func unlock(in *term.Input) func(map[string]any, []any) (any, error) {
	return func(env map[string]any, args []any) (any, error) {
		if lisp.IsDocCall(env, args) {
			return `unlock will unlock the next stage of the puzzle.`, nil
		} else if len(args) == 0 {
			return `cannot unlock without a pin code`, nil
		} else if len(args) == 1 {
			val, err := lisp.EvalForm(env, args[0])
			if err != nil {
				return nil, err
			}
			pin, ok := val.(string)
			if !ok {
				pinNm, ok := val.(float64)
				if !ok {
					return nil, errors.New("that pin code is indecipherable")
				}
				pin = fmt.Sprintf("%v", pinNm)
			}
			if pin == pinNumber && touched == 4 {
				term.Println(`{{"congrats"|bold}}, you have unlocked the next stage!`, nil)
				util.SetStage(in, "merrygoround")
			} else if pin == pinNumber && touched != 4 {
				return nil, errors.New("the pin does nothing without the buttons in place")
			} else if pin != pinNumber {
				str, _ := term.Sprintf("{{. | red}} is incorrect", pin)
				return nil, errors.New(str)
			}
		}
		return nil, nil
	}
}
