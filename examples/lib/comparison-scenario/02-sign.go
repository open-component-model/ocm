// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const SIGNATURE_NAME = "acme.org"

func Sign(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	cv, err := CreateComponentVersion(ctx)
	if err != nil {
		return err
	}
	defer cv.Close()
	err = SignComponentVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "signing failed")
	}
	return nil
}

// SignComponentVersion creates a key pair, registered it for
// further use and signs the component version.
func SignComponentVersion(cv ocm.ComponentVersionAccess) error {
	fmt.Printf("*** signing component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	privkey, pubkey, err := rsa.CreateKeyPair()
	if err != nil {
		return errors.Wrapf(err, "cannot create RSA key pair")
	}

	signinfo := signingattr.Get(cv.GetContext())
	signinfo.RegisterPublicKey(SIGNATURE_NAME, pubkey)
	signinfo.RegisterPrivateKey(SIGNATURE_NAME, privkey)

	_, err = signing.SignComponentVersion(cv, SIGNATURE_NAME)
	if err != nil {
		return errors.Wrapf(err, "signing failed")
	}
	PrintSignatures(cv)
	return err
}
