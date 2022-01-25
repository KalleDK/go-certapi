package certapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CertBackend interface {
	GetCert(t CertType, domain string, key APIKey) (cert []byte, err error)
	HasAccess(t CertType, domain string, key APIKey) bool
}

type CertMgr struct {
	backend CertBackend
	engine  *gin.Engine
}

func NewCertMgr(id uuid.UUID) (c CertMgr) {
	c.engine = gin.Default()

	c.engine.GET("/favicon.ico", serveFavicon)
	c.engine.GET("/ping", servePing(id))

	c.engine.GET("/cert/:domain/:certtype", serveCerts(c.backend))

	return
}

func serveKey(backend CertBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.GetString("domain")
		apikey, _ := c.Get("apikey")

		keyfile, err := ch.KeyFile(domain, keystr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.File(keyfile)
		}

	}
}

func parseAPIKey(c *gin.Context) (key APIKey) {
	keystr := c.GetHeader("Authorization")
	if len(keystr) < len("Bearer ") {
		return APIKey{}
	}

	keystr = keystr[len("Bearer "):]
	if err := key.UnmarshalText([]byte(keystr)); err != nil {
		return APIKey{}
	}

	return
}

func serveCerts(backend CertBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")
		certtype := CertType(c.Param("certtype"))
		key := parseAPIKey(c)

		cert, err := backend.GetCert(certtype, domain, key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {

			c.Writer.Header().Set("Content-Type", "application/x-pem-file")
			c.Writer.WriteHeader(http.StatusOK)
			c.Writer.Write(cert)
		}
	}
}

func servePing(id uuid.UUID) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("pong: %s", id),
		})
	}
}

func serveFavicon(c *gin.Context) {
	const favicon = `<svg
xmlns="http://www.w3.org/2000/svg"
viewBox="0 0 16 16">

<text x="0" y="14">🔒</text>
</svg>`

	c.Header("Content-Type", "image/svg+xml")
	c.Writer.WriteString(favicon)
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

	if !bytes.Equal(key[:], ch.Key[:]) {
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
