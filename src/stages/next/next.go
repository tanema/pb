package next

import (
	_ "embed"

	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/usage.tmpl
	usage string
	//go:embed data/manpage.man
	manPage string
)

func Run(in *term.Input) error {
	if err := util.InstallManpage(in.DB, manPage); err != nil {
		return err
	}
	return util.ErrorFmt(usage, in.Env.User)
}
