package term

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/sethvargo/go-envconfig"
)

// Input captures terminal input
type Input struct {
	IsTTY   bool
	HasPipe bool
	Flags   map[string]any
	Args    []string
	Stdin   string
	Env     struct {
		User   string `env:"USER,default=Timmy"`
		Home   string `env:"HOME"`
		Shell  string `env:"SHELL"`
		Editor string `env:"EDITOR"`
		Lang   string `env:"LANG"`
	}
}

// ParseInput will parse flags and positional arguments as well as read from
// stdin to fully collect all inputs
func ParseInput() *Input {
	stat, _ := os.Stdin.Stat()
	in := &Input{
		IsTTY:   isatty.IsTerminal(os.Stdout.Fd()),
		HasPipe: (stat.Mode() & os.ModeCharDevice) == 0,
		Flags:   map[string]any{},
	}
	in.readPipe()
	in.parseArgs()
	envconfig.Process(context.Background(), &in.Env)
	return in
}

// None will return true if the cli was passed no arguments
func (in *Input) None() bool {
	return len(in.Flags) == 0 && len(in.Args) == 0 && !in.HasPipe
}

// HasOpt will check if the cli was provided a flag OR arg that matches
func (in *Input) HasOpt(args ...string) bool {
	return in.HasArgs(args...) || in.HasFlags(args...)
}

// HasFlags will return true if one of the flags was used
func (in *Input) HasFlags(flags ...string) bool {
	for _, flag := range flags {
		if _, ok := in.Flags[flag]; ok {
			return true
		}
	}
	return false
}

// HasArgs will return true if one of the flags was used
func (in *Input) HasArgs(args ...string) bool {
	for _, toBeFound := range args {
		for _, arg := range in.Args {
			if toBeFound == arg {
				return true
			}
		}
	}
	return false
}

func (in *Input) readPipe() {
	if !in.HasPipe {
		return
	}
	var buf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		buf.Write(scanner.Bytes())
	}
	in.Stdin = buf.String()
}

func (in *Input) parseArgs() {
	args := os.Args[1:]
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			arg = strings.TrimPrefix(arg, "--")
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				in.Flags[strings.ToLower(parts[0])] = parts[1]
			} else {
				in.Flags[strings.ToLower(arg)] = true
			}
		} else if strings.HasPrefix(arg, "-") {
			arg = strings.TrimPrefix(arg, "-")
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				in.Flags[strings.ToLower(parts[0])] = parts[1]
			} else {
				for _, v := range strings.Split(arg, "") {
					in.Flags[v] = true
				}
			}
		} else {
			in.Args = append(in.Args, strings.ToLower(arg))
		}
	}
}
