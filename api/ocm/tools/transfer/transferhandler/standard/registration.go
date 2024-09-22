package standard

import (
	"fmt"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils/registrations"
)

func init() {
	transferhandler.RegisterHandlerRegistrationHandler("ocm/standard", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ transferhandler.ByNameCreationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) ByName(ctx ocm.Context, path string, olist ...transferhandler.TransferOption) (bool, transferhandler.TransferHandler, error) {
	if path != "" {
		return true, nil, fmt.Errorf("invalid standard handler %q", path)
	}

	h, err := New(olist...)
	return true, h, err
}

func (r *RegistrationHandler) GetHandlers(target *transferhandler.Target) registrations.HandlerInfos {
	return registrations.NewLeafHandlerInfo("standard transfer handler", `
The <code>standard</code> transfer handler works on the standard transfer options.`,
	)
}
