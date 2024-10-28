package ppi

import (
	"reflect"

	"github.com/mandelsoft/goutils/generics"
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

var TransferHandlerQuestions = map[string]reflect.Type{
	Q_UPDATE_VERSION:    generics.TypeOf[ComponentVersionQuestion](),
	Q_ENFORCE_TRANSPORT: generics.TypeOf[ComponentVersionQuestion](),
	Q_OVERWRITE_VERSION: generics.TypeOf[ComponentVersionQuestion](),
	Q_TRANSFER_VERSION:  generics.TypeOf[ComponentReferenceQuestion](),
	Q_TRANSFER_RESOURCE: generics.TypeOf[ArtifactQuestion](),
	Q_TRANSFER_SOURCE:   generics.TypeOf[ArtifactQuestion](),
}
