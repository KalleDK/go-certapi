package certapi

import (
	"time"
)

type CertType string

const (
	Key         CertType = "key"
	Cert        CertType = "certificate"
	CertChain   CertType = "certificate-chain"
	PKCS12      CertType = "pkcs12"
	PKCS12Chain CertType = "pkcs12-chain"
)

type CertInfo struct {
	StartDate     time.Time
	NextRenewTime time.Time
	Serial        string
}
