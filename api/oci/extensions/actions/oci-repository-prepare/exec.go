package oci_repository_prepare

import (
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/datacontext/action/handlers"
	common "ocm.software/ocm/api/utils/misc"
)

func Execute(hdlrs handlers.Registry, host, repo string, creds common.Properties) (*ActionResult, error) {
	return generics.CastR[*ActionResult](hdlrs.Execute(Spec(host, repo), creds))
}
