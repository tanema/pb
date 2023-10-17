package next

import (
	_ "embed"

	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

type NextStage struct {
	in    *term.Input
	usage string
	hints []string
}

func New(in *term.Input) *NextStage {
	return &NextStage{
		in:    in,
		usage: "That is it for now! This is will sit here until more stages are added!",
		hints: []string{"nothing here because there is nothing to do! You're done!"},
	}
}

func (stage *NextStage) Title() string              { return "Next Up" }
func (stage *NextStage) Man() string                { return stage.usage }
func (stage *NextStage) Help() string               { return stage.usage }
func (stage *NextStage) Hints() []string            { return stage.hints }
func (stage *NextStage) Options() map[string]string { return nil }
func (stage *NextStage) Run() error                 { return util.ErrorShowUsage }
