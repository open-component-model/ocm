// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/listformat"
)

func New(ctx clictx.Context) *cobra.Command {
	var standard credentials.IdentityMatcherInfos
	var consumer credentials.IdentityMatcherInfos
	for _, e := range ctx.CredentialsContext().ConsumerIdentityMatchers().List() {
		if e.IsConsumerType() {
			consumer = append(consumer, e)
		} else {
			standard = append(standard, e)
		}
	}

	return &cobra.Command{
		Use:   "credential-handling",
		Short: "Provisioning of credentials for credential consumers",
		Long: `
In contrast to libraries intended for a dedicated technical environment,
for example the handling of OCI images in OCI registries, the OCM
ecosystem cannot provide a specialized credential management for a decicated
environment.

Because of its extensibility working with component versions could
require access to any kind of technical system, either for storing
the model elements in a storage backend, or for accessing content
in any kind of technical storage system. There are several kinds of
credential consumers with potentially completely different kinds of credentials.
Therefore, a common uniform credential management is required, capable to serve
all those use cases.

This credential management brings together various kinds of credential consumers,
for example the access to artifacts in OCI registries or accessing
Git repository content, and credential providers, like
vaults or local files in the filesystem (for example a technology
specific credential source like the docker config json file for 
accessing OCI registries).

The used credential management model is based on four elements:
- *Credentials:*

  Credentials are described property set (key/value pairs).
- *Consumer Ids*

  Because of the extensible nature of the OCM model, credential consumers
  must be formally identified. A consumer id described a concrete
  access, which must be authorized.

  This is again achieved by a set of simple named attributes. There is only
  one defined property, which must always be present, the <code>type</code> attibute.
  It denotes the type of the technical environment credentials are required for.
  For example, for accessing OCI or Git registries. Additionally, there may 
  be any number of arbitrary attributes used to describe the concrete
  instance of such an environment and access paths in this environment, which
  should be accessed (for example the OCI registry URL to describe the instance
  and the repository path for the set of objects, which should be accessed)

  There are two use cases for consumer ids:
  - *Credential Request.* They are used by a credential consumer to issue a 
    credential request to the credential management. Hereby, they describe the
    concrete element, which should accessed.
  - *Credential Assignment.* The credential management allows to assign 
    credentials to consumer ids

- *Credential Providers* or repositories

  Credential repositories are dedicated kinds of implementations, which provide
  access to names sets of credentials stored in any kind of technical 
  environment, for example a vault or a credentials somewhere on the local
  filesystem.

- *Identity Matchers*

  The credential management must resolve credential requests against a set
  of credential assignments. This is not necessarily a complete attribute match
  for the involved consumer ids. There is typically some kind of matching 
  involved. For example, an assigment is done for an OCI registry with a dedicated
  server url and prefix for the repository path (type is OCIRegistry, host is
  ghcr.io, prefixpath is open-component-model). The assigned credentials
  should be applicable for sub repositories. Som the assignment use a more
  general consumer id than the concrete credential request (for example for
  repository path <code>open-component-model/ocm/ocmcli</code>)

  This kind of matching depend on the used attribute and is therefore in general
  type specific. Therefore, every consumer type uses an own identity matcher,
  which is then used by the credential management to find the best matching
  assignment.

The general process for a credential management then looks as follows.
- credentials provided by credential repositories are assigned to generalized
  consumer ids.
- a concrete access operation for a technical environment calculates
  a detailed consumer id for the element, which should be accessed
- it asks the credential management for credentials for this id
- the management examines all defined assignments to find the best
  matching one based on the provided matching mechanism.
- it then returns the mapped credentials from the references repository.

The critical task for a user of the toolset is to define those assignments.
This is basically a manual task, because the credentials stored in vault
(for example) could be usable for any kind of system, which typically
cannot be derived from the credential values.

But luckily, those could partly be automated:
- there may be credential providers, which are technology specific, for example
  the docker config json is used to describe credentials for OCI registries.
  Such providers can automatically assign the found credentials to appropriate
  consumer ids.
- If the credential store has the possibility to store custom meta data for a 
  credential set, this metadata can be used to describe the intended consumer
  ids. The provider implementation then uses this info create the appropriate
  assignments.

### Consumer Types and Matchers

The following credential consumer types are used/supported:
` + listformat.FormatListElements("", consumer) + `\
Those consumer types provide their own matchers, which are often based
on some standard generic matches. Those generic matchers and their
behaviours are described in the following list:
` + listformat.FormatListElements("", standard) + `

### Credential Providers

` + ctx.CredentialsContext().RepositoryTypes().Describe(),
	}
}
