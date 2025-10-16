package builder

import (
	"github.com/mandelsoft/filepath/pkg/filepath"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils"
)

// TODO: switch to context local setting.
func (b *Builder) RSAKeyPair(name ...string) {
	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	b.failOn(err)
	reg := signingattr.Get(b.OCMContext())
	for _, n := range name {
		reg.RegisterPublicKey(n, pub)
		reg.RegisterPrivateKey(n, priv)
	}
}

func (b *Builder) ReadRSAKeyPair(name, path string) {
	reg := signingattr.Get(b.OCMContext())
	pubfound := false
	path, _ = utils.ResolvePath(path)
	if ok, _ := b.Exists(filepath.Join(path, "rsa.pub")); ok {
		pubbytes, err := b.ReadFile(filepath.Join(path, "rsa.pub"))
		b.failOn(err)
		pub, err := signutils.ParsePublicKey(pubbytes)
		b.failOn(err)
		reg.RegisterPublicKey(name, pub)
		pubfound = true
	}
	if ok, _ := b.Exists(filepath.Join(path, "rsa.priv")); ok {
		privbytes, err := b.ReadFile(filepath.Join(path, "rsa.priv"))
		b.failOn(err)
		priv, err := signutils.ParsePrivateKey(privbytes)
		b.failOn(err)
		reg.RegisterPrivateKey(name, priv)
		if !pubfound {
			pub, _, err := rsa.GetPublicKey(priv)
			b.failOn(err)
			reg.RegisterPublicKey(name, pub)
		}
	}
}
