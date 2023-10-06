package util

import (
	"errors"
	"fmt"
	"os"

	"github.com/tanema/pb/src/term"
)

func SetStage(in *term.Input, stage string) {
	in.DB.Set("stage", stage)
	in.DB.Set("hints", "0")
	os.Exit(0)
}

func DisplayHint(in *term.Input, hints []string) error {
	hint, _ := term.Sprintf(hints[in.Hints], nil)
	in.Hints = (in.Hints + 1) % len(hints)
	in.DB.Set("hints", fmt.Sprintf("%v", in.Hints))
	return errors.New(hint)
}
