package certapi

import (
	"time"
)

const (
	Info        string = "info"
	Key         string = "key"
	Cert        string = "certificate"
	CertChain   string = "certificate-chain"
	PKCS12      string = "pkcs12"
	PKCS12Chain string = "pkcs12-chain"
)

const (
	MimeTypeInfo        string = "application/json"
	MimeTypeKey         string = "application/x-pem-file"
	MimeTypeCert        string = "application/x-pem-file"
	MimeTypeCertChain   string = "application/x-pem-file"
	MimeTypePKCS12      string = "application/x-pkcs12"
	MimeTypePKCS12Chain string = "application/x-pkcs12"
)

const (
	FilenameInfo        string = "certificate.json"
	FilenameKey         string = "certificate.key"
	FilenameCert        string = "certificate.crt"
	FilenameCertChain   string = "certificate-chain.crt"
	FilenamePKCS12      string = "certificate.pfx"
	FilenamePKCS12Chain string = "certificate-chain.pfx"
)

type CertInfo struct {
	StartDate     time.Time
	NextRenewTime time.Time
	Serial        string
}
