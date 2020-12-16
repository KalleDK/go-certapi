package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

type CertInfo struct {
	StartDate time.Time
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
	key, err := cfg.Section("").GetKey("Le_CertCreateTime")
	if err != nil {
		return err
	}
	v, err := key.Int64()
	if err != nil {
		return err
	}
	c.StartDate = time.Unix(v, 0)
	return nil
}

func main() {
	var certinfo CertInfo
	fp, err := os.Open("demo.ini")
	if err != nil {
		os.Exit(1)
	}
	defer fp.Close()
	certinfo.FromIni(fp)

	fmt.Println(certinfo.StartDate)

	router := gin.Default()
	router.Run()
}
