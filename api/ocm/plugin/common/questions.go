package common

import (
	"ocm.software/ocm/api/ocm/plugin/internal"
)

const (
	Q_UPDATE_VERSION    = internal.Q_UPDATE_VERSION
	Q_OVERWRITE_VERSION = internal.Q_OVERWRITE_VERSION
	Q_ENFORCE_TRANSPORT = internal.Q_ENFORCE_TRANSPORT
	Q_TRANSFER_VERSION  = internal.Q_TRANSFER_VERSION
	Q_TRANSFER_RESOURCE = internal.Q_TRANSFER_RESOURCE
	Q_TRANSFER_SOURCE   = internal.Q_TRANSFER_SOURCE
)

var TransferHandlerQuestions = map[string]string{
	Q_UPDATE_VERSION:    "update non-signature relevant parts",
	Q_ENFORCE_TRANSPORT: "enforce transport as component version does not exist in target ",
	Q_OVERWRITE_VERSION: "update signature relevant parts",
	Q_TRANSFER_VERSION:  "decide on updating a component version",
	Q_TRANSFER_RESOURCE: "value transport for resources",
	Q_TRANSFER_SOURCE:   "value transport for sources",
}
