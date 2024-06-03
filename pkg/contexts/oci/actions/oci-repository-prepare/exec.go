package oci_repository_prepare

import (
	"github.com/mandelsoft/goutils/generics"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/handlers"
)

func Execute(hdlrs handlers.Registry, host, repo string, creds common.Properties) (*ActionResult, error) {
	return generics.CastR[*ActionResult](hdlrs.Execute(Spec(host, repo), creds))
}
