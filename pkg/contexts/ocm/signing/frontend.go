// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

func SignComponentVersion(printer common.Printer, cv ocm.ComponentVersionAccess, name string, privkey interface{}, optlist ...Option) (*metav1.DigestSpec, error) {
	var opts Options

	opts.Eval(
		SignatureName(name),
		PrivateKey(name, privkey),
		Update(),
		Recursive(),
		VerifyDigests(),
	)
	opts.Eval(optlist...)

	if opts.VerifySignature {
		return nil, errors.Newf("impossible verification option set for signing")
	}
	if opts.Signer == nil {
		opts.Signer = signing.DefaultHandlerRegistry().GetSigner(rsa.Algorithm)
	}
	err := opts.Complete(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "inconsistent options for signing")
	}
	return Apply(printer, nil, cv, &opts)
}

func VerifyComponentVersion(printer common.Printer, cv ocm.ComponentVersionAccess, name string, pubkey interface{}, optlist ...Option) (*metav1.DigestSpec, error) {
	var opts Options
	if len(cv.GetDescriptor().Signatures) == 1 && name == "" {
		name = cv.GetDescriptor().Signatures[0].Name
	}

	opts.Eval(
		VerifyDigests(),
		VerifySignature(name),
		Recursive(),
	)
	if name != "" && pubkey != nil {
		PublicKey(name, pubkey).ApplySigningOption(&opts)
	}
	opts.Eval(optlist...)

	if opts.SignatureName() != "" && (opts.Keys == nil || opts.Keys.GetPublicKey(opts.SignatureName()) == nil) {
		PublicKey(name, pubkey).ApplySigningOption(&opts)
	}

	if opts.Signer != nil {
		return nil, errors.Newf("impossible signer option set for verification")
	}
	err := opts.Complete(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "inconsistent options for verification")
	}
	return Apply(printer, nil, cv, &opts)
}
