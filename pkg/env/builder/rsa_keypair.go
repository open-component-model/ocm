// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

// TODO: switch to context local setting.
func (b *Builder) RSAKeyPair(name string) {
	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	b.failOn(err)
	reg := signingattr.Get(b.OCMContext())
	reg.RegisterPublicKey(name, pub)
	reg.RegisterPrivateKey(name, priv)
}
