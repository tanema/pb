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

type MerryStage struct {
	in    *term.Input
	usage string
	hints []string
}

func New(in *term.Input) *MerryStage {
	return &MerryStage{
		in: in,
		usage: `This stage will require some tricks. The puzzlebox is in disguise. It may talk
to you differently depending on how you speak to it.`,
		hints: []string{
			`ew it smells like ouroboros in here!`,
			`https://media.giphy.com/media/TGKPVy5nvJQLxlxliH/giphy.gif`,
			`have you tried stdin?`,
			`how can you chain the puzzle box to itself?`,
		},
	}
}

func (stage *MerryStage) Title() string              { return "Merry-Go-Round" }
func (stage *MerryStage) Man() string                { return stage.usage }
func (stage *MerryStage) Help() string               { return stage.usage }
func (stage *MerryStage) Hints() []string            { return stage.hints }
func (stage *MerryStage) Options() map[string]string { return nil }

func (stage *MerryStage) Run() error {
	if !stage.in.DB.Key("current_app_name") {
		stage.in.DB.Set("current_app_name", os.Args[0])
	}

	if stage.in.DB.Get("current_app_name") != os.Args[0] {
		fmt.Println("you have done it! I have transformed! You have now completed the puzzle box.")
		util.SetStage(stage.in, "next")
		return nil
	} else if !stage.in.None() && len(stage.in.Stdin) == 0 {
		return term.Errorf(`not like that, speak to me like we are on {{"Love is Blind"|magenta}}`, nil)
	} else if key, err := crypto.LoadKey(stage.in.DB); err != nil {
		return err
	} else if len(stage.in.Stdin) > 0 {
		return stage.consume(key)
	} else {
		return stage.puke(key)
	}
}

func (stage *MerryStage) puke(key *crypto.EncryptionKey) error {
	cipher, err := key.Encrypt([]byte("rename me and you will release me!"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Print(util.Base64(string(cipher)))
	return nil
}

func (stage *MerryStage) consume(key *crypto.EncryptionKey) error {
	cipherText, err := util.DecodeBase64(string(stage.in.Stdin))
	if err != nil {
		return errors.New("this is not base64!")
	}
	text, err := key.Decrypt(cipherText)
	if err != nil {
		return errors.New("failed to decrypt the message! Are you sure you sent me the correct stuff?")
	}
	return errors.New(string(text))
}
