package certapi

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const keySize = sha256.Size
const encodedSize = keySize * 2

type APIKey [keySize]byte

func (k APIKey) MarshalText() (text []byte, err error) {
	text = make([]byte, encodedSize)
	hex.Encode(text, k[:])
	return
}

func (k *APIKey) UnmarshalText(text []byte) (err error) {
	if n, err := hex.Decode(k[:], text); err != nil || n != keySize {
		return errors.New("invalid API key length")
	}
	return
}

func (k APIKey) String() string {
	return hex.EncodeToString(k[:])
}
