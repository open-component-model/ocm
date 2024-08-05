package internal

import (
	"ocm.software/ocm/api/credentials"
)

type UploadTargetSpecInfo struct {
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}
