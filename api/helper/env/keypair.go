package env

import (
	"github.com/mandelsoft/filepath/pkg/filepath"

	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils"
)

func (e *Environment) RSAKeyPair(name ...string) {
	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	e.failOn(err)
	reg := signingattr.Get(e.OCMContext())
	for _, n := range name {
		reg.RegisterPublicKey(n, pub)
		reg.RegisterPrivateKey(n, priv)
	}
}

func (e *Environment) ReadRSAKeyPair(name, path string) {
	reg := signingattr.Get(e.OCMContext())
	pubfound := false
	path, _ = utils.ResolvePath(path)
	if ok, _ := e.Exists(filepath.Join(path, "rsa.pub")); ok {
		pubbytes, err := e.ReadFile(filepath.Join(path, "rsa.pub"))
		e.failOn(err)
		pub, err := signutils.ParsePublicKey(pubbytes)
		e.failOn(err)
		reg.RegisterPublicKey(name, pub)
		pubfound = true
	}
	if ok, _ := e.Exists(filepath.Join(path, "rsa.priv")); ok {
		privbytes, err := e.ReadFile(filepath.Join(path, "rsa.priv"))
		e.failOn(err)
		priv, err := signutils.ParsePrivateKey(privbytes)
		e.failOn(err)
		reg.RegisterPrivateKey(name, priv)
		if !pubfound {
			pub, _, err := rsa.GetPublicKey(priv)
			e.failOn(err)
			reg.RegisterPublicKey(name, pub)
		}
	}
}
