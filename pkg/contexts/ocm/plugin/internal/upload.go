package internal

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

type UploadTargetSpecInfo struct {
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}
