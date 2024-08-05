package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/examples/lib/helper"
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
