package util

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tanema/pb/src/term"
)

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

func Hex(in string, data ...any) string {
	return hex.EncodeToString([]byte(fmt.Sprintf(in, data...)))
}
