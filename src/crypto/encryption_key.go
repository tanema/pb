package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tanema/pb/src/pstore"
	"github.com/tanema/pb/src/util"
)

type (
	// EncryptionKey contains server key information
	EncryptionKey struct {
		KID    string `json:"kid"`
		Enc    string `json:"enc"`
		EncKey string `json:"encKey"`
		RawKey []byte `json:"-"`
	}
	// Ciphertext contains encrypted message
	ciphertext struct {
		KID  string `json:"kid"`
		Enc  string `json:"enc"`
		Cty  string `json:"cty"`
		Iv   string `json:"iv"`
		Data string `json:"data"`
	}
)

func LoadKey(db *pstore.DB) (*EncryptionKey, error) {
	if result, err := readKey(db); err == nil {
		return result, nil
	} else {
		rawKey := randomNBytes(32)
		key := &EncryptionKey{
			KID:    "master-key",
			Enc:    "A256GCM",
			EncKey: base64.RawURLEncoding.EncodeToString(rawKey),
			RawKey: rawKey,
		}
		return key, writeKey(db, key)
	}
}

func readKey(db *pstore.DB) (*EncryptionKey, error) {
	key := &EncryptionKey{}
	if data, err := util.DecodeBase64(db.Get("skeleton")); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, &key); err != nil {
		return nil, err
	} else if encKey, err := base64.RawURLEncoding.DecodeString(key.EncKey); err != nil {
		return nil, err
	} else {
		key.RawKey = encKey
	}
	return key, nil
}

func writeKey(db *pstore.DB, key *EncryptionKey) error {
	buf, err := json.Marshal(key)
	if err != nil {
		return err
	}
	return db.Set("skeleton", util.Base64(string(buf)))
}

func randomNBytes(size int) []byte {
	out := make([]byte, size)
	io.ReadFull(rand.Reader, out[:])
	return out
}

// Decrypt decrypts a given ciphertext byte array using the web crypto key
func (key *EncryptionKey) Decrypt(message []byte) ([]byte, error) {
	msg := &ciphertext{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, err
	} else if msg.KID != key.KID {
		return nil, fmt.Errorf("attempt to decrypt message with KID %v using different KID %v", msg.KID, key.KID)
	} else if msg.Enc != "A256GCM" {
		return nil, fmt.Errorf("attempt to decrypt message with unknown enc: %+q", msg.Enc)
	} else if msg.Cty != "jwk+json" {
		return nil, fmt.Errorf("attempt to decrypt message with unknown cty: %+q", msg.Cty)
	} else if ciphertext, err := base64.RawURLEncoding.DecodeString(msg.Data); err != nil {
		return nil, err
	} else if block, err := aes.NewCipher(key.RawKey); err != nil {
		return nil, err
	} else if iv, err := base64.RawURLEncoding.DecodeString(msg.Iv); err != nil {
		return nil, err
	} else if len(iv) != 12 {
		return nil, fmt.Errorf("invalid iv length (%d) in the message, expected 12", len(iv))
	} else if aead, err := cipher.NewGCM(block); err != nil {
		return nil, err
	} else if plaintext, err := aead.Open(nil, iv, ciphertext, nil); err != nil {
		return nil, err
	} else {
		return plaintext, nil
	}
}

// Encrypt encrypts a given plaintext byte array
func (key *EncryptionKey) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key.RawKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := randomNBytes(aead.NonceSize())
	return json.Marshal(ciphertext{
		KID:  key.KID,
		Enc:  "A256GCM",
		Cty:  "jwk+json",
		Iv:   base64.RawURLEncoding.EncodeToString(iv),
		Data: base64.RawURLEncoding.EncodeToString(aead.Seal(nil, iv, plaintext, nil)),
	})
}
