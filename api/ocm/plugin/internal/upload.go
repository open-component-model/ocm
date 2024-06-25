package internal

import (
	"github.com/open-component-model/ocm/api/credentials"
)

type UploadTargetSpecInfo struct {
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}
