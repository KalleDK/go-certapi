package certserver

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var CertName = map[certapi.CertType]string{
	certapi.Key:         "certificate.key",
	certapi.Cert:        "certificate.crt",
	certapi.CertChain:   "certificate-chain.crt",
	certapi.PKCS12:      "certificate.pfx",
	certapi.PKCS12Chain: "certificate-chain.pfx",
}

var CertMIME = map[certapi.CertType]string{
	certapi.Key:         "application/x-pem-file",
	certapi.Cert:        "application/x-pem-file",
	certapi.CertChain:   "application/x-pem-file",
	certapi.PKCS12:      "application/x-pkcs12",
	certapi.PKCS12Chain: "application/x-pkcs12",
}

type CertFileInfo struct {
	ModTime time.Time
	Size    int64
}

type CertFile struct {
	CertFileInfo
	Data []byte
}

type CertBackend interface {
	GetCertFile(domain string, t certapi.CertType, key certapi.APIKey) (CertFile, error)
	GetCertInfo(domain string, key certapi.APIKey) (certapi.CertInfo, error)
}

type CertMgr struct {
	backend CertBackend
	engine  *gin.Engine
}

func (c *CertMgr) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.engine.ServeHTTP(w, req)
}

func NewCertHandler(id uuid.UUID, backend CertBackend) http.Handler {
	return NewCertMgr(id, backend)
}

func NewCertMgr(id uuid.UUID, backend CertBackend) *CertMgr {
	mgr := &CertMgr{
		backend: backend,
		engine:  gin.Default(),
	}

	mgr.engine.GET("/favicon.ico", serveFavicon)
	mgr.engine.GET("/ping", servePing(id))
	mgr.engine.GET("/cert/:domain", serveCertInfo(backend))
	mgr.engine.GET("/cert/:domain/:certtype", serveCerts(backend))

	return mgr
}

func parseAPIKey(c *gin.Context) (key certapi.APIKey) {
	keystr := c.GetHeader("Authorization")
	if len(keystr) < len("Bearer ") {
		return certapi.APIKey{}
	}

	keystr = keystr[len("Bearer "):]
	if err := key.UnmarshalText([]byte(keystr)); err != nil {
		return certapi.APIKey{}
	}

	return
}

func serveError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func serveFile(c *gin.Context, filename string, modtime time.Time, contenttype string, data []byte) {
	header := c.Writer.Header()
	header.Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	header.Set("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, contenttype, data)
}

func serveCertInfo(backend CertBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")
		key := parseAPIKey(c)

		certinfo, err := backend.GetCertInfo(domain, key)
		if err != nil {
			serveError(c, err)
			return
		}

		c.JSON(http.StatusOK, certinfo)

	}
}

func serveCerts(backend CertBackend) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")
		certtype := certapi.CertType(c.Param("certtype"))
		key := parseAPIKey(c)

		filename, ok := CertName[certtype]
		if !ok {
			serveError(c, errors.New("missing certname"))
			return
		}

		contenttype, ok := CertMIME[certtype]
		if !ok {
			serveError(c, errors.New("missing certtype"))
			return
		}

		certfile, err := backend.GetCertFile(domain, certtype, key)
		if err != nil {
			serveError(c, err)
			return
		}

		serveFile(c, filename, certfile.ModTime, contenttype, certfile.Data)
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
