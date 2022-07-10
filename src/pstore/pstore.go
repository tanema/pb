package pstore

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/juju/fslock"
)

// DB is the store for data in a simple file
type DB struct {
	filename string
	flock    *fslock.Lock
	mx       sync.Mutex
}

// New will create a new file store in the user config path
func New(appName, filename string) (*DB, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	filename = filepath.Join(configDir, filename)
	if info, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		file.Close()
	} else if err != nil {
		return nil, err
	} else if info.IsDir() {
		return nil, fmt.Errorf("cannot create db out of a directory")
	}
	return &DB{filename: filename, flock: fslock.New(filename)}, nil
}

func (db *DB) begin() map[string]string {
	data := map[string]string{}
	file, _ := os.Open(db.filename)
	defer file.Close()
	gob.NewDecoder(file).Decode(&data)
	return data
}

func (db *DB) commit(data map[string]string) {
	file, _ := os.OpenFile(db.filename, os.O_WRONLY, 0755)
	defer file.Close()
	gob.NewEncoder(file).Encode(data)
}

func (db *DB) lock() {
	db.mx.Lock()
	db.flock.Lock()
}

func (db *DB) unlock() {
	db.mx.Unlock()
	db.flock.Unlock()
}

// Get will return the value for the key. If no value, an empty string will be
// returned
func (db *DB) Get(key string) string {
	db.lock()
	defer db.unlock()
	data := db.begin()
	return data[key]
}

// Key will return true if the key exists in the store
func (db *DB) Key(key string) bool {
	db.lock()
	defer db.unlock()
	data := db.begin()
	_, ok := data[key]
	return ok
}

// Keys will return all of the keys in the store
func (db *DB) Keys() []string {
	db.lock()
	defer db.unlock()
	data := db.begin()
	keys := []string{}
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// Set will set the value with the key in the store
func (db *DB) Set(key, val string) {
	db.lock()
	defer db.unlock()
	data := db.begin()
	data[key] = val
	db.commit(data)
}

// Del will remove the key/val from the store
func (db *DB) Del(key string) {
	db.lock()
	defer db.unlock()
	data := db.begin()
	delete(data, key)
	db.commit(data)
}

// Drop will wipe the entire store
func (db *DB) Drop() {
	db.lock()
	defer db.unlock()
	db.commit(map[string]string{})
}

// Dump will return all the data in the store
func (db *DB) Dump() map[string]string {
	db.lock()
	defer db.unlock()
	return db.begin()
}
