package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/examples/lib/helper"
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

	fmt.Printf("persisting public key to %s", KEYFILE)
	pubkey := signingattr.Get(ctx).GetPublicKey("acme.org")
	file, err := os.OpenFile(KEYFILE, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
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
	prefix := id[identity.ID_PATHPREFIX]
	i := strings.LastIndex(prefix, "/")
	if i > 0 {
		id[identity.ID_PATHPREFIX] = prefix[:i]
	}
	creds := identity.SimpleCredentials(cfg.Username, cfg.Password)
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
