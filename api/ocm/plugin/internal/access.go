package internal

import (
	"ocm.software/ocm/api/credentials"
)

type AccessSpecInfo struct {
	Short      string                       `json:"short"`
	MediaType  string                       `json:"mediaType"`
	Hint       string                       `json:"hint"`
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}
