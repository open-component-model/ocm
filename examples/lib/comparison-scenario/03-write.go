package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-component-model/ocm/examples/lib/helper"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const KEYFILE = "/tmp/comparison.pub"

func Write(cfg *helper.Config) error {
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
	err = WriteComponentVersion(cfg, cv)

	pubkey := signingattr.Get(ctx).GetPublicKey("acme.org")

	fmt.Printf("persisting public key to %s", KEYFILE)
	file, err := os.OpenFile(KEYFILE, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrapf(err, "cannot persist public key")
	}
	defer file.Close()
	err = rsa.WriteKeyData(pubkey, file)
	if err != nil {
		return errors.Wrapf(err, "cannot persist public key")
	}
	return nil
}

func RegisterPublicKeyFromFile(ctx ocm.Context) error {
	data, err := os.ReadFile(KEYFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot read key file %s", KEYFILE)
	}
	key, err := rsa.ParsePublicKey(data)
	if err != nil {
		return errors.Wrapf(err, "invalid %s", KEYFILE)
	}
	signingattr.Get(ctx).RegisterPublicKey("acme.org", key)
	return nil
}

func RegisterCredentials(ctx ocm.Context, cfg *helper.Config) error {
	credctx := ctx.CredentialsContext()

	// register credentials for given OCI registry in context.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	fmt.Printf("consumer id for repository %s: %s\n", cfg.Repository, id)
	prefix := id[ociidentity.ID_PATHPREFIX]
	i := strings.LastIndex(prefix, "/")
	if i > 0 {
		id[ociidentity.ID_PATHPREFIX] = prefix[:i]
	}
	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)
	credctx.SetCredentialsForConsumer(id, creds)
	return nil
}

// WriteComponentVersion writes a component version to an
// OCM repository.
func WriteComponentVersion(cfg *helper.Config, cv ocm.ComponentVersionAccess) error {
	fmt.Printf("*** writing component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	ctx := cv.GetContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return err
	}

	// now get the access to the repository
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot get repository access for %s", cfg.Repository)
	}
	defer repo.Close()

	err = repo.AddComponentVersion(cv, true)
	if err != nil {
		return errors.Wrapf(err, "cannot add version")
	}
	return nil
}
