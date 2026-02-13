## ocm attributes &mdash; Configuration Attributes Used To Control The Behaviour

### Description

The OCM library supports a set of attributes, which can be used to influence
the behaviour of various functions. The CLI also supports setting of those
attributes using the config file (see [ocm configfile](ocm_configfile.md)) or by
command line options of the main command (see [ocm](ocm.md)).

The following options are available in the currently used version of the
OCM library:
- <code>github.com/mandelsoft/logforward</code> [<code>logfwd</code>]: *logconfig* Logging config structure used for config forwarding

  This attribute is used to specify a logging configuration intended
  to be forwarded to other tools.
  (For example: TOI passes this config to the executor)

- <code>github.com/mandelsoft/oci/cache</code> [<code>cache</code>]: *string*

  Filesystem folder to use for caching OCI blobs

- <code>github.com/mandelsoft/ocm/compat</code> [<code>compat</code>]: *bool*

  Compatibility mode: Avoid generic local access methods and prefer type specific ones.

- <code>github.com/mandelsoft/ocm/hasher</code>: *JSON*

  Preferred hash algorithm to calculate resource digests. The following
  digesters are supported:
    - <code>NO-DIGEST</code>
    - <code>SHA-256</code> (default)
    - <code>SHA-512</code>

- <code>github.com/mandelsoft/ocm/keeplocalblob</code> [<code>keeplocalblob</code>]: *bool*

  Keep local blobs when importing OCI artifacts to OCI registries from <code>localBlob</code>
  access methods. By default, they will be expanded to OCI artifacts with the
  access method <code>ociRegistry</code>. If this option is set to true, they will be stored
  as local blobs, also. The access method will still be <code>localBlob</code> but with a nested
  <code>ociRegistry</code> access method for describing the global access.

- <code>github.com/mandelsoft/ocm/mapocirepo</code> [<code>mapocirepo</code>]: *bool|YAML*

  When uploading an OCI artifact blob to an OCI based OCM repository and the
  artifact is uploaded as OCI artifact, the repository path part is shortened,
  either by hashing all but the last repository name part or by executing
  some prefix based name mappings.

  If a boolean is given the short hash or none mode is enabled.
  The YAML flavor uses the following fields:
  - *<code>mode</code>* *string*: <code>hash</code>, <code>shortHash</code>, <code>prefixMapping</code>
    or <code>none</code>. If unset, no mapping is done.
  - *<code>prefixMappings</code>*: *map[string]string* repository path prefix mapping.
  - *<code>prefix</code>*: *string* repository prefix to use (replaces potential sub path of OCM repo).
    or <code>none</code>.
  - *<code>prefixMapping</code>*: *map[string]string* repository path prefix mapping.

  Notes:

  - The mapping only occurs in transfer commands and only when transferring to OCI registries (e.g.
    when transferring to a CTF archive this option will be ignored).
  - The mapping in mode <code>prefixMapping</code> requires a full prefix of the composed final name.
    Partial matches are not supported. The host name of the target will be skipped.
  - The artifact name of the component-descriptor is not mapped.
  - If the mapping is provided on the command line it must be JSON format and needs to be properly
    escaped (see example below).

  Example:

  Assume a component named <code>github.com/my_org/myexamplewithalongname</code> and a chart name
  <code>echo</code> in the <code>Charts.yaml</code> of the chart archive. The following input to a
  <code>resource.yaml</code> creates a component version:

  <pre>
  name: mychart
  type: helmChart
  input:
    type: helm
    path: charts/mychart.tgz
  ---
  name: myimage
  type: ociImage
  version: 0.1.0
  input:
    type: ociImage
    repository: ocm/ocm.software/ocmcli/ocmcli-image
    path: ghcr.io/acme/ocm/ocm.software/ocmcli/ocmcli-image:0.1.0
  </pre>

  The following command:

  <pre>
  ocm "-X mapocirepo={\"mode\":\"mapping\",\"prefixMappings\":{\"acme/github.com/my_org/myexamplewithalongname/ocm/ocm.software/ocmcli\":\"acme/cli\", \"acme/github.com/my_org/myexamplewithalongnameabc123\":\"acme/mychart\"}}" transfer ctf -f --copy-resources ./ctf ghcr.io/acme
  </pre>

  will result in the following artifacts in <code>ghcr.io/my_org</code>:

  <pre>
  mychart/echo
  cli/ocmcli-image
  </pre>

  Note that the host name part of the transfer target <code>ghcr.io/acme</code> is excluded from the
  prefix but the path <code>acme</code> is considered.

  The same using a config file <code>.ocmconfig</code>:
  <pre>
  type: generic.config.ocm.software/v1
  configurations:
  ...
  - type: attributes.config.ocm.software
    attributes:
  	...
  	mapocirepo:
  	  mode: mapping
  	  prefixMappings:
  	    acme/github.com/my\_org/myexamplewithalongname/ocm/ocm.software/ocmcli: acme/cli
  		acme/github.com/my\_org/myexamplewithalongnameabc123: acme/mychart
  </pre>

  <pre>
  ocm transfer ca -f --copy-resources ./ca ghcr.io/acme
  </pre>

- <code>github.com/mandelsoft/ocm/ociuploadrepo</code> [<code>ociuploadrepo</code>]: *oci base repository ref*

  Upload local OCI artifact blobs to a dedicated repository.

- <code>github.com/mandelsoft/ocm/plugindir</code> [<code>plugindir</code>]: *plugin directory*

  Directory to look for OCM plugin executables.

- <code>github.com/mandelsoft/ocm/rootcerts</code> [<code>rootcerts</code>]: *JSON*

  General root certificate settings given as JSON document with the following
  format:

  <pre>
  {
    "rootCertificates": [
       {
         "data": ""&lt;base64>"
       },
       {
         "path": ""&lt;file path>"
       }
    ]
  }
  </pre>

  One of following data fields are possible:
  - <code>data</code>:       base64 encoded binary data
  - <code>stringdata</code>: plain text data
  - <code>path</code>:       a file path to read the data from

- <code>github.com/mandelsoft/ocm/signing</code>: *JSON*

  Public and private Key settings given as JSON document with the following
  format:

  <pre>
  {
    "publicKeys": [
       "&lt;provider>": {
         "data": ""&lt;base64>"
       }
    ],
    "privateKeys"": [
       "&lt;provider>": {
         "path": ""&lt;file path>"
       }
    ]
  }
  </pre>

  One of following data fields are possible:
  - <code>data</code>:       base64 encoded binary data
  - <code>stringdata</code>: plain text data
  - <code>path</code>:       a file path to read the data from

- <code>github.com/mandelsoft/tempblobcache</code> [<code>blobcache</code>]: *string* Foldername for temporary blob cache

  The temporary blob cache is used to accessing large blobs from remote systems.
  The are temporarily stored in the filesystem, instead of the memory, to avoid
  blowing up the memory consumption.

- <code>ocm.software/cliconfig</code> [<code>cliconfig</code>]: *cliconfig* Configuration Object passed to command line plugin.



- <code>ocm.software/compositionmode</code> [<code>compositionmode</code>]: *bool* (default: false)

  Composition mode decouples a component version provided by a repository
  implementation from the backend persistence. Added local blobs will
  and other changes will not be forwarded to the backend repository until
  an AddVersion is called on the component.
  If composition mode is disabled blobs will directly be forwarded to
  the backend and descriptor updated will be persisted on AddVersion
  or closing a provided existing component version.

- <code>ocm.software/ocm/api/datacontext/attrs/httptimeout</code> [<code>timeout</code>]: *string*

  Configures the timeout duration for HTTP client requests used to access
  OCI registries and other remote endpoints. The value is specified as a
  Go duration string (e.g. "30s", "5m", "1h").

  If not set, the default timeout of 30s is used.

- <code>ocm.software/ocm/api/ocm/extensions/attrs/maxworkers</code> [<code>maxworkers</code>]: *integer* or *"auto"*

  Specifies the maximum number of concurrent workers to use for resource and source,
  as well as reference transfer operations.

  Supported values:
    - A positive integer: use exactly that number of workers.
    - The string "auto": automatically use the number of logical CPU cores.
    - Zero or omitted: fall back to single-worker mode (1). This is the default.
      This mode guarantees deterministic ordering of operations.

  Precedence:
    1. Attribute set in the current OCM context.
    2. Environment variable OCM_TRANSFER_WORKER_COUNT.
    3. Default value (1).

  WARNING: This is an experimental feature and may cause unexpected behavior
  depending on workload concurrency. Values above 1 may result in non-deterministic
  transfer ordering.

- <code>ocm.software/ocm/oci/preferrelativeaccess</code> [<code>preferrelativeaccess</code>]: *bool*

  If an artifact blob is uploaded to the technical repository
  used as OCM repository, the uploader should prefer to return
  a relative access method.

- <code>ocm.software/signing/sigstore</code> [<code>sigstore</code>]: *sigstore config* Configuration to use for sigstore based signing.


  Configuration applies to <code>sigstore</code> (legacy) and <code>sigstore-v2</code> signing algorithms.
  The algorithms affect how signatures are stored in Rekor:
  - <code>sigstore</code>: stores only public key in Rekor entry (non-compliant Sigstore Bundle spec).
  - <code>sigstore-v2</code>: stores Fulcio certificate in Rekor entry (compliant Sigstore Bundle spec).

  The following fields are used.
  - *<code>fulcioURL</code>* *string*  default is https://fulcio.sigstore.dev
  - *<code>rekorURL</code>* *string*  default is https://rekor.sigstore.dev
  - *<code>OIDCIssuer</code>* *string*  default is https://oauth2.sigstore.dev/auth
  - *<code>OIDCClientID</code>* *string*  default is sigstore
### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

