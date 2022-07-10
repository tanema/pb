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
	db, err := pstore.New("pb", ".data")
	if err != nil {
		fmt.Println("There was a problem dipping my toe into your system. I cannot work like this.")
		os.Exit(1)
	}
	stages.Run(in, db)
}
