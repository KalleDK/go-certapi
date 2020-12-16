package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type APIKey [sha256.Size]byte

func (k APIKey) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%x", k)), nil
}

func (k *APIKey) UnmarshalText(text []byte) (err error) {
	if n, err := hex.Decode(k[:], text); err != nil || n != sha256.Size {
		return errors.New("invalid api key length")
	}
	return nil
}

type Settings struct {
	CertHome string
	Key      APIKey
}

func main() {
	sum := sha256.Sum256([]byte("hello world"))
	str := fmt.Sprintf("%x", sum)
	fmt.Println(str)

	s := Settings{
		CertHome: "Flaf",
		Key:      sum,
	}

	b, _ := json.Marshal(s)
	fmt.Println(string(b))
	fmt.Println(s.Key)

	var s2 Settings
	json.Unmarshal(b, &s2)
	fmt.Println(s)
	fmt.Println(s2)
}
