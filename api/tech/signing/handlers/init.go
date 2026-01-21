package handlers

import (
	_ "github.com/sigstore/cosign/v3/pkg/providers/all"
	_ "ocm.software/ocm/api/tech/signing/handlers/rsa"
	_ "ocm.software/ocm/api/tech/signing/handlers/rsa-pss"
	_ "ocm.software/ocm/api/tech/signing/handlers/rsa-pss-signingservice"
	_ "ocm.software/ocm/api/tech/signing/handlers/rsa-signingservice"
	_ "ocm.software/ocm/api/tech/signing/handlers/sigstore"
)
