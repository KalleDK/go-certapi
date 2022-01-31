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

var certMIME = map[string]string{
	certapi.Key:         certapi.MimeTypeKey,
	certapi.Cert:        certapi.MimeTypeCert,
	certapi.CertChain:   certapi.MimeTypeCertChain,
	certapi.PKCS12:      certapi.MimeTypePKCS12,
	certapi.PKCS12Chain: certapi.MimeTypePKCS12Chain,
}

var certName = map[string]string{
	certapi.Key:         certapi.FilenameKey,
	certapi.Cert:        certapi.FilenameCert,
	certapi.CertChain:   certapi.FilenameCertChain,
	certapi.PKCS12:      certapi.FilenamePKCS12,
	certapi.PKCS12Chain: certapi.FilenamePKCS12Chain,
}

type CertMgr struct {
	backend certapi.CertService
	engine  *gin.Engine
}

func (c *CertMgr) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.engine.ServeHTTP(w, req)
}

func NewCertHandler(id uuid.UUID, backend certapi.CertService) http.Handler {
	return NewCertMgr(id, backend)
}

func NewCertMgr(id uuid.UUID, backend certapi.CertService) *CertMgr {
	mgr := &CertMgr{
		backend: backend,
		engine:  gin.Default(),
	}

	mgr.engine.GET("/favicon.ico", serveFavicon)
	mgr.engine.GET("/ping", servePing(id))
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

func serveCerts(backend certapi.CertService) func(c *gin.Context) {
	return func(c *gin.Context) {
		domain := c.Param("domain")
		certtype := c.Param("certtype")
		key := parseAPIKey(c)

		filename, ok := certName[certtype]
		if !ok {
			serveError(c, errors.New("missing certname"))
			return
		}

		contenttype, ok := certMIME[certtype]
		if !ok {
			serveError(c, errors.New("missing certtype"))
			return
		}

		certfile, err := backend.GetItem(domain, certtype, key)
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
