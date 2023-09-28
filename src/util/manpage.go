package util

import (
	"os"

	"github.com/tanema/pb/src/artifacts"
	"github.com/tanema/pb/src/pstore"
)

func InstallManpage(db *pstore.DB, manPage string) error {
	file, err := os.Create("/usr/local/share/man/man1/pb.1")
	if err != nil {
		return err
	} else if _, err = file.Write([]byte(manPage)); err != nil {
		return err
	}
	artifacts.Add(db, file.Name())
	return file.Close()
}
