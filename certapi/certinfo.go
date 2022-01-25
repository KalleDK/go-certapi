package certapi

import (
	"io"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type CertType string

const (
	CertFile      CertType = "certificate"
	KeyFile       CertType = "key"
	FullchainFile CertType = "fullchain"
)

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

type CertInfo struct {
	StartDate     time.Time
	NextRenewTime time.Time
	Serial        string
}

func (c *CertInfo) FromIni(r io.Reader) (err error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	cfg, err := ini.Load(b)
	if err != nil {
		return err
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

	return nil
}
