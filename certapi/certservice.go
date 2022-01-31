package certapi

import "time"

type ItemInfo struct {
	ModTime time.Time
	Size    int64
}

type Item struct {
	ItemInfo
	Data []byte
}

type CertService interface {
	GetItem(domain string, itemtype string, apikey APIKey) (item Item, err error)
	GetItemInfo(domain string, itemtype string, apikey APIKey) (info ItemInfo, err error)
}
