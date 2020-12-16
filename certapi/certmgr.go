package certapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/ini.v1"
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
	ID       uuid.UUID
	CertHome string
	Key      APIKey
}

type CertInfo struct {
	StartDate     time.Time
	NextRenewTime time.Time
	Serial        string
}

func (c *CertInfo) FromIni(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	cfg, err := ini.Load(b)
	if err != nil {
		return err
	}
	section := cfg.Section("")

	{
		key, err := section.GetKey("Le_CertCreateTime")
		if err != nil {
			return err
		}
		v, err := key.Int64()
		if err != nil {
			return err
		}
		c.StartDate = time.Unix(v, 0)
	}

	{
		key, err := section.GetKey("Le_NextRenewTime")
		if err != nil {
			return err
		}
		v, err := key.Int64()
		if err != nil {
			return err
		}
		c.NextRenewTime = time.Unix(v, 0)
	}

	{
		key, err := section.GetKey("Le_LinkCert")
		if err != nil {
			return err
		}
		v := key.String()
		if err != nil {
			return err
		}
		idx := strings.LastIndex(v, "/")
		c.Serial = v[idx+1:]
	}

	return nil
}

type CertHome struct {
	Path string
	Key  APIKey
}

func (ch CertHome) KeyFile(domain string, keystr string) (string, error) {
	var key APIKey
	if err := key.UnmarshalText([]byte(keystr)); err != nil {
		return "", errors.New("invalid api key 1")
	}

	if bytes.Compare(key[:], ch.Key[:]) != 0 {
		return "", errors.New("invalid api key 2")
	}

	p := filepath.Join(ch.Path, domain, domain+".key")
	f, err := os.Stat(p)
	if err != nil || f.IsDir() {
		return "", err
	}
	return p, nil
}

func (ch CertHome) Cert(domain string) (string, error) {
	p := filepath.Join(ch.Path, domain, domain+".cer")
	f, err := os.Stat(p)
	if err != nil || f.IsDir() {
		return "", err
	}
	return p, nil
}

func (ch CertHome) Full(domain string) (string, error) {
	p := filepath.Join(ch.Path, domain, "fullchain.cer")
	f, err := os.Stat(p)
	if err != nil || f.IsDir() {
		return "", err
	}
	return p, nil
}

func (ch CertHome) Info(domain string) (info CertInfo, err error) {
	p := filepath.Join(ch.Path, domain, domain+".conf")
	fp, err := os.Open(p)
	if err != nil {
		return info, errors.New("domain not found")
	}
	if err := info.FromIni(fp); err != nil {
		return info, err
	}

	return
}
