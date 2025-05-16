---
title: "credential handling"
url: "/docs/cli-reference/credential-handling/"
---

## ocm credential-handling &mdash; Provisioning Of Credentials For Credential Consumers

### Description

In contrast to libraries intended for a dedicated technical environment,
for example the handling of OCI images in OCI registries, the OCM
ecosystem cannot provide a specialized credential management for a dedicated
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
  one defined property, which must always be present, the <code>type</code> attribute.
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
  involved. For example, an assignment is done for an OCI registry with a dedicated
  server url and prefix for the repository path (type is OCIRegistry, host is
  ghcr.io, prefix path is open-component-model). The assigned credentials
  should be applicable for sub repositories. So the assignment uses a more
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
  - <code>Buildcredentials.ocm.software</code>: Gardener config credential matcher

    It matches the <code>Buildcredentials.ocm.software</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type Buildcredentials.ocm.software evaluate the following credential properties:

      - <code>key</code>: secret key use to access the credential server


  - <code>Git</code>: Git credential matcher

    It matches the <code>Git</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type Git evaluate the following credential properties:

      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password
      - <code>token</code>: HTTP token authentication
      - <code>privateKey</code>: Private Key authentication certificate


  - <code>Github</code>: GitHub credential matcher

    This matcher is a hostpath matcher.

    Credential consumers of the consumer type Github evaluate the following credential properties:

      - <code>token</code>: GitHub personal access token


  - <code>HashiCorpVault</code>: HashiCorp Vault credential matcher

    This matcher matches credentials for a HashiCorp vault instance.
    It uses the following identity attributes:
      - <code>hostname</code>: vault server host
      - <code>scheme</code>: (optional) URL scheme
      - <code>port</code>: (optional) server port
      - <code>namespace</code>: vault namespace
      - <code>mountPath</code>: mount path
      - <code>pathprefix</code>: path prefix for secret


    Credential consumers of the consumer type HashiCorpVault evaluate the following credential properties:

      - <code>authmeth</code>: auth method
      - <code>token</code>: vault token
      - <code>roleid</code>: app-role role id
      - <code>secretid</code>: app-role secret id

    The only supported auth methods, so far, are <code>token</code> and <code>approle</code>.


  - <code>HelmChartRepository</code>: Helm chart repository

    It matches the <code>HelmChartRepository</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type HelmChartRepository evaluate the following credential properties:

      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password
      - <code>certificate</code>: TLS client certificate
      - <code>privateKey</code>: TLS private key
      - <code>certificateAuthority</code>: TLS certificate authority


  - <code>MavenRepository</code>: MVN repository

    It matches the <code>MavenRepository</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type MavenRepository evaluate the following credential properties:

      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password


  - <code>NpmRegistry</code>: NPM registry

    It matches the <code>NpmRegistry</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type NpmRegistry evaluate the following credential properties:

      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password
      - <code>email</code>: NPM registry, require an email address
      - <code>token</code>: the token attribute. May exist after login at any npm registry. Check your .npmrc file!


  - <code>OCIRegistry</code>: OCI registry credential matcher

    It matches the <code>OCIRegistry</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type OCIRegistry evaluate the following credential properties:

      - <code>username</code>: the basic auth username
      - <code>password</code>: the basic auth password
      - <code>identityToken</code>: the bearer token used for non-basic auth authorization
      - <code>certificateAuthority</code>: the certificate authority certificate used to verify certificates


  - <code>S3</code>: S3 credential matcher

    This matcher is a hostpath matcher.

    Credential consumers of the consumer type S3 evaluate the following credential properties:

      - <code>awsAccessKeyID</code>: AWS access key id
      - <code>awsSecretAccessKey</code>: AWS secret for access key id
      - <code>token</code>: AWS access token (alternatively)


  - <code>Signingserver.gardener.cloud</code>: signing service credential matcher

    This matcher matches credentials for a Signing Service instance.
    It uses the following identity attributes:
      - <code>hostname</code>: signing server host
      - <code>scheme</code>: (optional) URL scheme
      - <code>port</code>: (optional) server port
      - <code>pathprefix</code>: path prefix for the server URL


    Credential consumers of the consumer type Signingserver.gardener.cloud evaluate the following credential properties:

      - <code>clientCert</code>: client certificate for authentication
      - <code>privateKey</code>: private key for client certificate
      - <code>caCerts</code>: root certificate for signing server


  - <code>wget</code>: wget credential matcher

    It matches the <code>wget</code> consumer type and additionally acts like
    the <code>hostpath</code> type.

    Credential consumers of the consumer type wget evaluate the following credential properties:

      - <code>username</code>: the basic auth user name
      - <code>password</code>: the basic auth password
      - <code>identityToken</code>: the bearer token used for non-basic auth authorization
      - <code>certificateAuthority</code>: the certificate authority certificate used to verify certificates presented by the server
      - <code>certificate</code>: the certificate used to present to the server
      - <code>privateKey</code>: the private key corresponding to the certificate


\
Those consumer types provide their own matchers, which are often based
on some standard generic matches. Those generic matchers and their
behaviors are described in the following list:
  - <code>exact</code>: exact match of given pattern set
  - <code>hostpath</code>: Host and path based credential matcher

    This matcher works on the following properties:

    - *<code>type</code>* (required if set in pattern): the identity type
    - *<code>hostname</code>* (required if set in pattern): the hostname of a server
    - *<code>scheme</code>* (optional): the URL scheme of a server
    - *<code>port</code>* (optional): the port of a server
    - *<code>pathprefix</code>* (optional): a path prefix to match. The
      element with the most matching path components is selected (separator is <code>/</code>).


  - <code>partial</code>: complete match of given pattern ignoring additional attributes


### Credential Providers

Credential providers offer sets of named credentials from various sources,
which might be directly mapped to consumer identities (if supported
by the provider type).

The type <code>Credentials</code> can be used to inline
credentials in credential configuration objects
to configure mappings of consumer identities to a credential
set (see [ocm configfile](ocm_configfile.md)).

The following types are currently available:

- Credential provider <code>Credentials</code>

  This repository type can be used to specify a single inline credential
  set. The default name is the empty string or <code>Credentials</code>.

  The following versions are supported:
  - Version <code>v1</code>

    The repository specification supports the following fields:
      - <code>properties</code>: *map[string]string*: direct credential fields


- Credential provider <code>DockerConfig</code>

  This repository type can be used to access credentials stored in a file
  following the docker config json format. It take into account the
  credentials helper section, also. If enabled, the described
  credentials will be automatically assigned to appropriate consumer ids.

  The following versions are supported:
  - Version <code>v1</code>

    The repository specification supports the following fields:
      - <code>dockerConfigFile</code>: *string*: the file path to a docker config file
      - <code>dockerConfig</code>: *json*: an embedded docker config json
      - <code>propagateConsumerIdentity</code>: *bool*(optional): enable consumer id propagation


- Credential provider <code>HashiCorpVault</code>

  This repository type can be used to access credentials stored in a HashiCorp
  Vault.

  It provides access to list of secrets stored under a dedicated path in
  a vault namespace. This list can either explicitly be specified, or
  it is taken from the metadata of a specified secret.

  The following custom metadata attributes are evaluated:
  - <code>secrets</code> this attribute may contain a comma separated list of
    vault secrets, which should be exposed by this repository instance.
    The names are evaluated under the path prefix used for the repository.
  - <code>consumerId</code> this attribute may contain a JSON encoded
    consumer id , this secret should be assigned to.
  - <code>type</code> if no special attribute is defined this attribute
    indicated to use the complete custom metadata as consumer id.

  It uses the HashiCorpVault identity matcher and consumer type
  to requests credentials for the access.


  This matcher matches credentials for a HashiCorp vault instance.
  It uses the following identity attributes:
    - <code>hostname</code>: vault server host
    - <code>scheme</code>: (optional) URL scheme
    - <code>port</code>: (optional) server port
    - <code>namespace</code>: vault namespace
    - <code>mountPath</code>: mount path
    - <code>pathprefix</code>: path prefix for secret


  It requires the following credential attributes:

    - <code>authmeth</code>: auth method
    - <code>token</code>: vault token
    - <code>roleid</code>: app-role role id
    - <code>secretid</code>: app-role secret id

  The only supported auth methods, so far, are <code>token</code> and <code>approle</code>.

  The following versions are supported:
  - Version <code>v1</code>

    The repository specification supports the following fields:
      - <code>serverURL</code>: *string* (required): the URL of the vault instance
      - <code>namespace</code>: *string* (optional): the namespace used to evaluate secrets
      - <code>mountPath</code>: *string* (optional): the mount path to use (default: secrets)
      - <code>path</code>: *string* (optional): the path prefix used to lookup secrets
      - <code>secrets</code>: *[]string* (optional): list of secrets
      - <code>propagateConsumerIdentity</code>: *bool*(optional): evaluate metadata for consumer id propagation

    If the secrets list is empty, all secret entries found in the given path
    is read.


- Credential provider <code>NPMConfig</code>

  This repository type can be used to access credentials stored in a file
  following the NPM npmrc format (~/.npmrc). It take into account the
  credentials helper section, also. If enabled, the described
  credentials will be automatically assigned to appropriate consumer ids.

  The following versions are supported:
  - Version <code>v1</code>

    The repository specification supports the following fields:
      - <code>npmrcFile</code>: *string*: the file path to a NPM npmrc file
      - <code>propagateConsumerIdentity</code>: *bool*(optional): enable consumer id propagation


### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

