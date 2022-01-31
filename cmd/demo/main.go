package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certapi/certapi/certserver"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
)

type AcmeBackend struct {
	fs     fs.FS
	authdb map[certapi.APIKey]string
}

func (b *AcmeBackend) realname(domain string, ctype string) (string, error) {
	switch ctype {
	case certapi.Cert:
		return fmt.Sprintf("%s.cer", domain), nil
	case certapi.Key:
		return fmt.Sprintf("%s.key", domain), nil
	}
	return "", errors.New("file not found")
}

func parseIniTime(section *ini.Section, name string) (t time.Time, err error) {
	key, err := section.GetKey(name)
	if err != nil {
		return
	}

	v, err := key.Int64()
	if err != nil {
		return
	}

	return time.Unix(v, 0), nil
}

func parseIniSerial(section *ini.Section, name string) (s string, err error) {
	key, err := section.GetKey(name)
	if err != nil {
		return
	}

	v := key.String()
	if err != nil {
		return
	}

	idx := strings.LastIndex(v, "/")
	return v[idx+1:], nil
}

func fromIni(b []byte) (c certapi.CertInfo, err error) {

	cfg, err := ini.Load(b)
	if err != nil {
		return
	}

	section := cfg.Section("")

	c.StartDate, err = parseIniTime(section, "Le_CertCreateTime")
	if err != nil {
		return
	}

	c.NextRenewTime, err = parseIniTime(section, "Le_NextRenewTime")
	if err != nil {
		return
	}

	c.Serial, err = parseIniSerial(section, "Le_LinkCert")
	if err != nil {
		return
	}

	return c, nil
}

func (b *AcmeBackend) GetItemInfo(domain string, itemtype string, apikey certapi.APIKey) (certinfo certapi.ItemInfo, err error) {
	/*
		apidomain, ok := b.authdb[apikey]
		if !(ok && (domain == apidomain)) {
			err = errors.New("invalid API key")
			return
		}

		sub, err := fs.Sub(b.fs, domain)
		if err != nil {
			return
		}

		data, err := fs.ReadFile(sub, fmt.Sprintf("%s.conf", domain))
		if err != nil {
			return
		}
	*/
	return
	//return fromIni(data)

}

func (b *AcmeBackend) GetItem(domain string, itemtype string, apikey certapi.APIKey) (cert certapi.Item, err error) {
	apidomain, ok := b.authdb[apikey]
	if !(ok && (domain == apidomain)) {
		err = errors.New("invalid API key")
		return
	}

	sub, err := fs.Sub(b.fs, domain)
	if err != nil {
		return
	}

	name, err := b.realname(domain, itemtype)
	if err != nil {
		return
	}

	stat, err := fs.Stat(sub, name)
	if err != nil {
		return
	}
	cert.ModTime = stat.ModTime()
	cert.Size = stat.Size()

	cert.Data, err = fs.ReadFile(sub, name)
	if err != nil {
		return
	}

	return cert, nil
}

func main() {
	k := certapi.APIKey{}
	k[0] = 245
	fmt.Println(k)
	c := certserver.NewCertMgr(uuid.UUID{}, &AcmeBackend{
		authdb: map[certapi.APIKey]string{
			k: "example.com",
		},
		fs: os.DirFS("../../data"),
	})
	http.ListenAndServe(":8000", c)
}
