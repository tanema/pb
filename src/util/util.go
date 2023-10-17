package util

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/tanema/pb/src/term"
)

var ErrorShowUsage = errors.New("showUsage")

func Base64(in string, data ...any) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(in, data...)))
}

func DecodeBase64(in string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(in)
}

func Hex(in string, data ...any) string {
	return hex.EncodeToString([]byte(fmt.Sprintf(in, data...)))
}

func OnSignal(fn func(os.Signal), sig ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	for {
		fn(<-c)
	}
}

func SetStage(in *term.Input, stage string) {
	in.DB.Set("stage", stage)
	in.DB.Set("hints", "0")
	os.Exit(0)
}
