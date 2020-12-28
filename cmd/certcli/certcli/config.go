package certcli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/KalleDK/go-certapi/certapi"
)

type DomainStore struct {
	Path   string
	Config *Config
}

func (s DomainStore) Init() error {
	{
		f, err := os.Open(s.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Readdirnames(1)
		if err == nil {
			return errors.New("dir is not empty")
		}
		if err != io.EOF {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(s.Path, "certs"), 755); err != nil {
		return err
	}
	if err := s.SaveConfig(Config{Certs: map[string]CertInfo{}}); err != nil {
		return err
	}

	return nil
}

func (s DomainStore) Get(domain string) (cinfo CertInfo, cstore CertStore, err error) {
	if s.Config == nil {
		s.Config = &Config{}
		if err = s.LoadConfig(s.Config); err != nil {
			return
		}
	}
	return s.Config.Certs[domain], CertStore{filepath.Join(s.Path, "certs", domain)}, nil
}

func (s DomainStore) LoadConfig(config *Config) error {
	b, err := ioutil.ReadFile(filepath.Join(s.Path, "domains.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, config)
}

func (s DomainStore) SaveConfig(conf Config) error {
	b := bytes.Buffer{}
	dec := json.NewEncoder(&b)
	dec.SetIndent("", "  ")
	if err := dec.Encode(conf); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(s.Path, "domains.json"), b.Bytes(), 666)
}

func (s DomainStore) Add(domain string, info CertInfo) (cstore CertStore, err error) {
	if s.Config == nil {
		s.Config = &Config{}
		if err = s.LoadConfig(s.Config); err != nil {
			return
		}
	}
	_, ok := s.Config.Certs[domain]
	if ok {
		return cstore, errors.New("domain already exists")
	}
	s.Config.Certs[domain] = info

	if err = os.Mkdir(filepath.Join(s.Path, "certs", domain), 755); err != nil {
		return
	}

	if err = s.SaveConfig(*s.Config); err != nil {
		return
	}

	return CertStore{filepath.Join(s.Path, "certs", domain)}, nil
}

func (s DomainStore) Remove(domain string) error {
	var config Config
	if err := s.LoadConfig(&config); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(s.Path, "certs", domain)); err != nil {
		return err
	}
	delete(config.Certs, domain)
	if err := s.SaveConfig(config); err != nil {
		return err
	}
	return nil
}

type CertStore struct {
	Path string
}

func (s CertStore) Env() []string {
	return []string{
		"CERT_STATE=" + filepath.Join(s.Path, "state"),
		"CERT_CERT=" + filepath.Join(s.Path, "server.cer"),
		"CERT_KEY=" + filepath.Join(s.Path, "server.key"),
		"CERT_FULL=" + filepath.Join(s.Path, "fullchain.cer"),
	}
}

func (s CertStore) Remove() error { return os.RemoveAll(s.Path) }

func (s CertStore) LoadState(state *certapi.CertInfo) error {
	statepath := filepath.Join(s.Path, "state")
	fp, err := os.Open(statepath)
	if err != nil {
		return nil
	}
	defer fp.Close()
	dec := json.NewDecoder(fp)
	return dec.Decode(state)
}

func (s CertStore) SaveState(state certapi.CertInfo) error {
	statepath := filepath.Join(s.Path, "state")
	fp, err := os.Create(statepath)
	if err != nil {
		return err
	}
	defer fp.Close()
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(state)
}

func (s CertStore) SaveCertificate(data []byte) error {
	return ioutil.WriteFile(filepath.Join(s.Path, "server.cer"), data, 666)
}

func (s CertStore) SaveKey(data []byte) error {
	return ioutil.WriteFile(filepath.Join(s.Path, "server.key"), data, 666)
}

func (s CertStore) SaveFullchain(data []byte) error {
	return ioutil.WriteFile(filepath.Join(s.Path, "fullchain.cer"), data, 666)
}

type CertInfo struct {
	Server    string
	ReloadCmd string   `json:",omitempty"`
	Args      []string `json:",omitempty"`
}

type Config struct {
	Certs map[string]CertInfo
}

func MakeBearer(pass string) string {
	sha := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sha[:])
}
