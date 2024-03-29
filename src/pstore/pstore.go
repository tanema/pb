package pstore

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// DB is the store for data in a simple file
type DB struct {
	filename string
	data     map[string]string
	mx       sync.Mutex
}

// New will create a new file store in the user config path
func New(appName, filename string) (*DB, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", appName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	db := &DB{
		filename: filepath.Join(configDir, filename),
		data:     map[string]string{},
	}
	return db, db.read()
}

func (db *DB) read() error {
	if _, err := os.Stat(db.filename); os.IsNotExist(err) {
		return nil
	} else if file, err := os.Open(db.filename); err != nil {
		return err
	} else if byteData, err := io.ReadAll(file); err != nil {
	} else if rawData, err := base64.StdEncoding.DecodeString(string(byteData)); err != nil {
		return err
	} else if string(rawData) == "" {
		return nil
	} else if err := json.Unmarshal(rawData, &db.data); err != nil {
		return err
	} else if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func (db *DB) commit() error {
	if file, err := os.OpenFile(db.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
	} else if rawData, err := json.Marshal(db.data); err != nil {
		return err
	} else if _, err := file.WriteString(base64.StdEncoding.EncodeToString(rawData)); err != nil {
		return err
	} else if err := file.Close(); err != nil {
		return err
	}
	return nil
}

// Get will return the value for the key. If no value, an empty string will be
// returned
func (db *DB) Get(key string) string {
	db.mx.Lock()
	defer db.mx.Unlock()
	return db.data[key]
}

// Key will return true if the key exists in the store
func (db *DB) Key(key string) bool {
	db.mx.Lock()
	defer db.mx.Unlock()
	_, ok := db.data[key]
	return ok
}

// Keys will return all of the keys in the store
func (db *DB) Keys() []string {
	db.mx.Lock()
	defer db.mx.Unlock()
	keys := []string{}
	for key := range db.data {
		keys = append(keys, key)
	}
	return keys
}

// Set will set the value with the key in the store
func (db *DB) Set(key, val string) error {
	db.mx.Lock()
	defer db.mx.Unlock()
	db.data[key] = val
	return db.commit()
}

// Del will remove the key/val from the store
func (db *DB) Del(key string) error {
	db.mx.Lock()
	defer db.mx.Unlock()
	delete(db.data, key)
	return db.commit()
}

// Drop will wipe the entire store
func (db *DB) Drop() error {
	db.mx.Lock()
	defer db.mx.Unlock()
	db.data = map[string]string{}
	return db.commit()
}

// Dump will return all the data in the store
func (db *DB) Dump() map[string]string {
	db.mx.Lock()
	defer db.mx.Unlock()
	return db.data
}
