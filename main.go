package main

import (
	"fmt"
	"os"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/stages"
	"github.com/tanema/pb/src/term"
)

func main() {
	in := term.ParseInput()
	if db, err := pstore.New("pb", ".data"); err != nil {
		fmt.Println("There was a problem dipping my toe into your system. I cannot work like this.")
		os.Exit(1)
	} else if err := stages.Run(in, db); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
