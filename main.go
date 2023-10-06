package main

import (
	"fmt"
	"os"

	"github.com/tanema/pb/src/stages"
	"github.com/tanema/pb/src/term"
)

func main() {
	if in, err := term.ParseInput(); err != nil {
		fmt.Println("There was a problem. I cannot work like this.")
		if in.HasFlags("V") {
			fmt.Println(err)
		}
		os.Exit(1)
	} else if err := stages.Run(in); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
