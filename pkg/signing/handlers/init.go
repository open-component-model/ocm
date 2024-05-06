package handlers

import (
	_ "github.com/sigstore/cosign/v2/pkg/providers/all"

	_ "github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	_ "github.com/open-component-model/ocm/pkg/signing/handlers/rsa-pss"
	_ "github.com/open-component-model/ocm/pkg/signing/handlers/rsa-pss-signingservice"
	_ "github.com/open-component-model/ocm/pkg/signing/handlers/rsa-signingservice"
	_ "github.com/open-component-model/ocm/pkg/signing/handlers/sigstore"
)
