package util

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/tanema/pb/src/term"
)

func WriteFmt(out io.Writer, template string, data any) {
	rnd, _ := term.Sprintf(template, data)
	out.Write([]byte(rnd))
}

func ErrorFmt(tmpl string, data interface{}) error {
	str, err := term.Sprintf(tmpl, data)
	if err != nil {
		return err
	}
	return errors.New(str)
}

func Base64(in string, data ...any) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(in, data...)))
}

func DecodeBase64(in string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(in)
}

func Hex(in string, data ...any) string {
	return hex.EncodeToString([]byte(fmt.Sprintf(in, data...)))
}
