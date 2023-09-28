package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanema/pb/src/pstore"
)

func get(db *pstore.DB) []string {
	raw := db.Get("artifacts")
	if raw == "" {
		return []string{}
	}
	return strings.Split(db.Get("artifacts"), ";")
}

func set(db *pstore.DB, artifacts []string) {
	db.Set("artifacts", strings.Join(artifacts, ";"))
}

// Add will add a new artifact path to the config
func Add(db *pstore.DB, artf string) {
	if artf == "" {
		return
	}
	artifacts := get(db)
	for _, path := range artifacts {
		if path == artf {
			return
		}
	}
	set(db, append(artifacts, artf))
}

// Remove will remove an artifact path from the config
func Remove(db *pstore.DB, artf string) {
	artifacts := get(db)
	for i, path := range artifacts {
		if path == artf {
			set(db, append(artifacts[:i], artifacts[i+1:]...))
			return
		}
	}
}

// Print will output a list of all of the artifacts
func Print(db *pstore.DB) {
	fmt.Print(strings.Join(get(db), "\n"))
}

// Setup will ensure that the config path is in the config
func Setup(db *pstore.DB) {
	Add(db, filepath.Join(os.Getenv("HOME"), ".config", "pb"))
}
