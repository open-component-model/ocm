package ocm

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "ocm"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
This input type allows to get a resource artifact from an OCM repository.

This blob type specification supports the following fields:
- **<code>ocmRepository</code>** *repository specification*

  This REQUIRED property describes the OCM repository specification

- **<code>component</code>** *string*

  This REQUIRED property describes the component na,e

- **<code>version</code>** *string*

  This REQUIRED property describes the version of a maven artifact.

- **<code>resourceRef</code>** *relative resource reference*
  
  This REQUIRED property describes the  resource reference for the desired
  resource relative to the given component version .
`
