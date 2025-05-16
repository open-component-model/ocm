---
title: "get credentials"
url: "/docs/cli-reference/get/credentials/"
---

## ocm get credentials &mdash; Get Credentials For A Dedicated Consumer Spec

### Synopsis

```bash
ocm get credentials {<consumer property>=<value>}
```

#### Aliases

```text
credentials, creds, cred
```

### Options

```text
  -h, --help             help for credentials
  -m, --matcher string   matcher type override
  -s, --sloppy           sloppy matching of consumer type
```

### Description

Try to resolve a given consumer specification against the configured credential
settings and show the found credential attributes.

Matchers exist for the following usage contexts or consumer types:
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



The following standard identity matchers are supported:
  - <code>exact</code>: exact match of given pattern set
  - <code>hostpath</code>: Host and path based credential matcher

    This matcher works on the following properties:

    - *<code>type</code>* (required if set in pattern): the identity type
    - *<code>hostname</code>* (required if set in pattern): the hostname of a server
    - *<code>scheme</code>* (optional): the URL scheme of a server
    - *<code>port</code>* (optional): the port of a server
    - *<code>pathprefix</code>* (optional): a path prefix to match. The
      element with the most matching path components is selected (separator is <code>/</code>).


  - <code>partial</code> (default): complete match of given pattern ignoring additional attributes

The used matcher is derived from the consumer attribute <code>type</code>.
For all other consumer types a matcher matching all attributes will be used.
The usage of a dedicated matcher can be enforced by the option <code>--matcher</code>.

### SEE ALSO

#### Parents

* [ocm get](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

