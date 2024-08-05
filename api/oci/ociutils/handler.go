package ociutils

import (
	"sync"

	"ocm.software/ocm/api/oci/cpi"
	common "ocm.software/ocm/api/utils/misc"
)

type InfoHandler interface {
	Description(pr common.Printer, m cpi.ManifestAccess, config []byte)
	Info(m cpi.ManifestAccess, config []byte) interface{}
}

var (
	lock     sync.Mutex
	handlers = map[string]InfoHandler{}
)

func RegisterInfoHandler(mime string, h InfoHandler) {
	lock.Lock()
	defer lock.Unlock()
	handlers[mime] = h
}

func getHandler(mime string) InfoHandler {
	lock.Lock()
	defer lock.Unlock()
	return handlers[mime]
}
