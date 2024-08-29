package maven

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "maven"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The <code>repoUrl</code> is the url pointing either to the http endpoint of a maven
repository (e.g. https://repo.maven.apache.org/maven2/) or to a file system based
maven repository (e.g. file://local/directory).

This blob type specification supports the following fields:
- **<code>repoUrl</code>** *string*

  This REQUIRED property describes the url from which the resource is to be
  accessed.

- **<code>groupId</code>** *string*

  This REQUIRED property describes the groupId of a maven artifact.

- **<code>artifactId</code>** *string*
	
  This REQUIRED property describes artifactId of a maven artifact.

- **<code>version</code>** *string*

  This REQUIRED property describes the version of a maven artifact.

- **<code>classifier</code>** *string*
  
  This OPTIONAL property describes the classifier of a maven artifact.

- **<code>extension</code>** *string*

  This OPTIONAL property describes the extension of a maven artifact.
`
