---
title: "ocm configfile - Configuration File"
linkTitle: "configfile"
url: "/docs/cli-reference/configfile/"
sidebar:
  collapsed: true
menu:
  docs:
    name: "configfile"
---

### Description

The command line client supports configuring by a given configuration file.
If existent, by default, the file <code>$HOME/.ocmconfig</code> will be read.
Using the option <code>--config</code> an alternative file can be specified.

The file format is yaml. It uses the same type mechanism used for all
kinds of typed specification in the ocm area. The file must have the type of
a configuration specification. Instead, the command line client supports
a generic configuration specification able to host a list of arbitrary configuration
specifications. The type for this spec is <code>generic.config.ocm.software/v1</code>.

The following configuration types are supported:

- <code>attributes.config.ocm.software</code>
  The config type <code>attributes.config.ocm.software</code> can be used to define a list
  of arbitrary attribute specifications:

  <pre>
      type: attributes.config.ocm.software
      attributes:
         &lt;name>: &lt;yaml defining the attribute>
         ...
  </pre>
- <code>blobLimits.ocireg.ocm.config.ocm.software</code>
  The config type <code>blobLimits.ocireg.ocm.config.ocm.software</code> can be used to set some
  blob layer limits for particular OCI registries used to host OCM repositories.
  The <code>blobLimits</code> field maps a OCI registry address to the blob limit to use:

  <pre>
      type: blobLimits.ocireg.ocm.config.ocm.software
      blobLimits:
          dummy.io: 65564
          dummy.io:8443: 32768 // with :8443 specifying the port and 32768 specifying the byte limit
  </pre>

  If blob limits apply to a registry, local blobs with a size larger than
  the configured limit will be split into several layers with a maximum
  size of the given value.

  These settings can be overwritten by explicit settings in an OCM
  repository specification for those repositories.

  The most specific entry will be used. If a registry with a dedicated
  port is requested, but no explicit configuration is found, the
  setting for the sole hostname is used (if configured).
- <code>cli.ocm.config.ocm.software</code>
  The config type <code>cli.ocm.config.ocm.software</code> is used to handle the
  main configuration flags of the OCM command line tool.

  <pre>
      type: cli.ocm.config.ocm.software
      aliases:
         &lt;name>: &lt;OCI registry specification>
         ...
  </pre>
- <code>credentials.config.ocm.software</code>
  The config type <code>credentials.config.ocm.software</code> can be used to define a list
  of arbitrary configuration specifications:

  <pre>
      type: credentials.config.ocm.software
      consumers:
        - identity:
            &lt;name>: &lt;value>
            ...
          credentials:
            - &lt;credential specification>
            ... credential chain
      repositories:
         - repository: &lt;repository specification>
           credentials:
            - &lt;credential specification>
            ... credential chain
      aliases:
         &lt;name>:
           repository: &lt;repository specification>
           credentials:
            - &lt;credential specification>
            ... credential chain
  </pre>
- <code>downloader.ocm.config.ocm.software</code>
  The config type <code>downloader.ocm.config.ocm.software</code> can be used to define a list
  of preconfigured download handler registrations (see [ocm ocm-downloadhandlers](ocm_ocm-downloadhandlers.md)),
  the default priority is 200:

  <pre>
      type: downloader.ocm.config.ocm.software
      description: "my standard download handler configuration"
      registrations:
        - name: oci/artifact
          artifactType: ociImage
          mimeType: ...
          description: ...
          priority: ...
          config: ...
        ...
  </pre>
- <code>generic.config.ocm.software</code>
  The config type <code>generic.config.ocm.software</code> can be used to define a list
  of arbitrary configuration specifications and named configuration sets:

  <pre>
      type: generic.config.ocm.software
      configurations:
        - type: &lt;any config type>
          ...
        ...
      sets:
         standard:
            description: my selectable standard config
            configurations:
              - type: ...
                ...
              ...
  </pre>

  Configurations are directly applied. Configuration sets are
  just stored in the configuration context and can be applied
  on-demand. On the CLI, this can be done using the main command option
  <code>--config-set &lt;name></code>.
- <code>hasher.config.ocm.software</code>
  The config type <code>hasher.config.ocm.software</code> can be used to define
  the default hash algorithm used to calculate digests for resources.
  It supports the field <code>hashAlgorithm</code>, with one of the following
  values:
    - <code>NO-DIGEST</code>
    - <code>SHA-256</code> (default)
    - <code>SHA-512</code>
- <code>keys.config.ocm.software</code>
  The config type <code>keys.config.ocm.software</code> can be used to define
  public and private keys. A key value might be given by one of the fields:
  - <code>path</code>: path of file with key data
  - <code>data</code>: base64 encoded binary data
  - <code>stringdata</code>: data a string parsed by key handler

  <pre>
      type: keys.config.ocm.software
      privateKeys:
         &lt;name>:
           path: &lt;file path>
         ...
      publicKeys:
         &lt;name>:
           data: &lt;base64 encoded key representation>
         ...
      rootCertificates:
        - path: &lt;file path>

      issuers:
         &lt;name>:
           commonName: acme.org
  </pre>

  Issuers define an expected distinguished name for
  public key certificates optionally provided together with
  signatures. They support the following fields:
  - <code>commonName</code> *string*
  - <code>organization</code> *string array*
  - <code>organizationalUnit</code> *string array*
  - <code>country</code> *string array*
  - <code>locality</code> *string array*
  - <code>province</code> *string array*
  - <code>streetAddress</code> *string array*
  - <code>postalCode</code> *string array*

  At least the given values must be present in the certificate
  to be accepted for a successful signature validation.
- <code>logging.config.ocm.software</code>
  The config type <code>logging.config.ocm.software</code> can be used to configure the logging
  aspect of a dedicated context type:

  <pre>
      type: logging.config.ocm.software
      contextType: attributes.context.ocm.software
      settings:
        defaultLevel: Info
        rules:
          - ...
  </pre>

  The context type attributes.context.ocm.software is the root context of a
  context hierarchy.

  If no context type is specified, the config will be applies to any target
  acting as logging context provider, which is not a non-root context.
- <code>memory.credentials.config.ocm.software</code>
  The config type <code>memory.credentials.config.ocm.software</code> can be used to define a list
  of arbitrary credentials stored in a memory based credentials repository:

  <pre>
      type: memory.credentials.config.ocm.software
      repoName: default
      credentials:
        - credentialsName: ref
          reference:  # refer to a credential set stored in some other credential repository
            type: Credentials # this is a repo providing just one explicit credential set
            properties:
              username: <my-user>
              password: <my-secret-password>
        - credentialsName: direct
          credentials: # direct credential specification
              username: <my-user>
              password: <my-secret-password>
  </pre>
- <code>merge.config.ocm.software</code>
  The config type <code>merge.config.ocm.software</code> can be used to set some
  assignments for the merging of (label) values. It applies to a value
  merge handler registry, either directly or via an OCM context.

  <pre>
      type: merge.config.ocm.software
      labels:
      - name: acme.org/audit/level
        merge:
          algorithm: acme.org/audit
          config: ...
      assignments:
         label:acme.org/audit/level@v1:
            algorithm: acme.org/audit
            config: ...
            ...
  </pre>
- <code>oci.config.ocm.software</code>
  The config type <code>oci.config.ocm.software</code> can be used to define
  OCI registry aliases:

  <pre>
      type: oci.config.ocm.software
      aliases:
         &lt;name>: &lt;OCI registry specification>
         ...
  </pre>
- <code>ocm.cmd.config.ocm.software</code>
  The config type <code>ocm.cmd.config.ocm.software</code> can be used to
  configure predefined aliases for dedicated OCM repositories and
  OCI registries.

  <pre>
     type: ocm.cmd.config.ocm.software
     ocmRepositories:
         &lt;name>: &lt;specification of OCM repository>
     ...
     ociRepositories:
         &lt;name>: &lt;specification of OCI registry>
     ...
  </pre>
- <code>ocm.config.ocm.software</code>
  The config type <code>ocm.config.ocm.software</code> can be used to set some
  configurations for an OCM context;

  <pre>
      type: ocm.config.ocm.software
      aliases:
         myrepo:
            type: &lt;any repository type>
            &lt;specification attributes>
            ...
      resolvers:
        - repository:
            type: &lt;any repository type>
            &lt;specification attributes>
            ...
          prefix: ghcr.io/open-component-model/ocm
          priority: 10
  </pre>

  With aliases repository alias names can be mapped to a repository specification.
  The alias name can be used in a string notation for an OCM repository.

  Resolvers define a list of OCM repository specifications to be used to resolve
  dedicated component versions. These settings are used to compose a standard
  component version resolver provided for an OCM context. Optionally, a component
  name prefix can be given. It limits the usage of the repository to resolve only
  components with the given name prefix (always complete name segments).
  An optional priority can be used to influence the lookup order. Larger value
  means higher priority (default 10).

  All matching entries are tried to lookup a component version in the following
  order:
  - highest priority first
  - longest matching sequence of component name segments first.

  If resolvers are defined, it is possible to use component version names on the
  command line without a repository. The names are resolved with the specified
  resolution rule.
  They are also used as default lookup repositories to lookup component references
  for recursive operations on component versions (<code>--lookup</code> option).
- <code>plugin.config.ocm.software</code>
  The config type <code>plugin.config.ocm.software</code> can be used to configure a
  plugin.

  <pre>
      type: plugin.config.ocm.software
      plugin: &lt;plugin name>
      config: &lt;arbitrary configuration structure>
      disableAutoRegistration: &lt;boolean flag to disable auto registration for up- and download handlers>
  </pre>
- <code>rootcerts.config.ocm.software</code>
  The config type <code>rootcerts.config.ocm.software</code> can be used to define
  general root certificates. A certificate value might be given by one of the fields:
  - <code>path</code>: path of file with key data
  - <code>data</code>: base64 encoded binary data
  - <code>stringdata</code>: data a string parsed by key handler

  <pre>
      rootCertificates:
        - path: &lt;file path>
  </pre>
- <code>scripts.ocm.config.ocm.software</code>
  The config type <code>scripts.ocm.config.ocm.software</code> can be used to define transfer scripts:

  <pre>
      type: scripts.ocm.config.ocm.software
      scripts:
        &lt;name>:
          path: &lt;>file path>
        &lt;other name>:
          script: &lt;>nested script as yaml>
  </pre>
- <code>transport.ocm.config.ocm.software</code>
  The config type <code>transport.ocm.config.ocm.software</code> can be used to define transfer scripts:

  <pre>
      type: transport.ocm.config.ocm.software
      recursive: true
      overwrite: true
      localResourcesByValue: false
      resourcesByValue: true
      sourcesByValue: false
      keepGlobalAccess: false
      stopOnExistingVersion: false
      omitAccessTypes:
      - s3
  </pre>
- <code>uploader.ocm.config.ocm.software</code>
  The config type <code>uploader.ocm.config.ocm.software</code> can be used to define a list
  of preconfigured upload handler registrations (see [ocm ocm-uploadhandlers](ocm_ocm-uploadhandlers.md)),
  the default priority is 200:

  <pre>
      type: uploader.ocm.config.ocm.software
      description: "my standard upload handler configuration"
      registrations:
        - name: oci/artifact
          artifactType: ociImage
          config:
            ociRef: ghcr.io/open-component-model/...
        ...
  </pre>

### Examples

```yaml
type: generic.config.ocm.software/v1
configurations:
  - type: credentials.config.ocm.software
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: "~/.docker/config.json"
          propagateConsumerIdentity: true
   - type: attributes.config.ocm.software
     attributes:  # map of attribute settings
       compat: true
#  - type: scripts.ocm.config.ocm.software
#    scripts:
#      "default":
#         script:
#           process: true
```

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm ocm-downloadhandlers</b>](ocm_ocm-downloadhandlers.md)	 &mdash; List of all available download handlers
* [<b>ocm ocm-uploadhandlers</b>](ocm_ocm-uploadhandlers.md)	 &mdash; List of all available upload handlers

