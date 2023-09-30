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
		fmt.Println("There was a problem. I cannot work like this.")
		if in.HasFlags("V") {
			fmt.Println(err)
		}
		os.Exit(1)
	} else if err := stages.Run(in, db); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
