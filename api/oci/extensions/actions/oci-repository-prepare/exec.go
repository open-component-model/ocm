package oci_repository_prepare

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/datacontext/action/handlers"
	"ocm.software/ocm/api/utils/misc"
)

func Execute(hdlrs handlers.Registry, host, repo string, creds misc.Properties) (*ActionResult, error) {
	return generics.CastR[*ActionResult](hdlrs.Execute(Spec(host, repo), creds))
}
