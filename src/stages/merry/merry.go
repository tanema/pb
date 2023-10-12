package merry

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/tanema/pb/src/crypto"
	"github.com/tanema/pb/src/term"
	"github.com/tanema/pb/src/util"
)

var (
	//go:embed data/usage.tmpl
	usage string
	//go:embed data/manpage.man
	manPage string
	//go:embed data/key.pem
	key   []byte
	hints = []string{
		`ew it smells like ouroboros in here!`,
		`https://media.giphy.com/media/TGKPVy5nvJQLxlxliH/giphy.gif`,
		`have you tried stdin?`,
		`how can you chain the puzzle box to itself?`,
	}
)

func Run(in *term.Input) error {
	if err := util.InstallManpage(in.DB, manPage); err != nil {
		return err
	} else if in.HasOpt("help", "h") {
		return util.ErrorFmt(usage, in.Env.User)
	} else if in.HasOpt("hint") {
		return util.DisplayHint(in, hints)
	} else if !in.None() && len(in.Stdin) == 0 {
		return util.ErrorFmt(`not like that, speak to me like we are on {{"Love is Blind"|magenta}}`, nil)
	}

	key, err := crypto.LoadKey()
	if err != nil {
		return err
	}

	if !in.DB.Key("current_app_name") {
		in.DB.Set("current_app_name", os.Args[0])
	} else if in.DB.Get("current_app_name") != os.Args[0] {
		fmt.Println("you have done it! I have transformed! You have now completed the puzzle box.")
		util.SetStage(in, "next")
	}

	if len(in.Stdin) > 0 {
		return consume(in, key)
	} else {
		return puke(in, key)
	}
}

func puke(in *term.Input, key *crypto.EncryptionKey) error {
	cipher, err := key.Encrypt([]byte("rename me and you will release me!"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Print(util.Base64(string(cipher)))
	return nil
}

func consume(in *term.Input, key *crypto.EncryptionKey) error {
	cipherText, err := util.DecodeBase64(string(in.Stdin))
	if err != nil {
		return errors.New("this is not base64!")
	}
	text, err := key.Decrypt(cipherText)
	if err != nil {
		return errors.New("failed to decrypt the message! Are you sure you sent me the correct stuff?")
	}
	return errors.New(string(text))
}
