## ocm &mdash; Open Component Model Command Line Client

### Synopsis

```bash
ocm [<options>] <sub command> ...
```

### Options

```text
  -X, --attribute stringArray     attribute setting
      --ca-cert stringArray       additional root certificate authorities (for signing certificates)
      --config stringArray        configuration file
      --config-set strings        apply configuration set
  -C, --cred stringArray          credential setting
  -h, --help                      help for ocm
  -I, --issuer stringArray        issuer name or distinguished name (DN) (optionally for dedicated signature) ([<name>:=]<dn>)
      --logJson                   log as json instead of human readable logs
      --logconfig string          log config
  -L, --logfile string            set log file
      --logkeys stringArray       log tags/realms(with leading /) to be enabled ([/[+]]name{,[/[+]]name}[=level])
  -l, --loglevel string           set log level
  -K, --private-key stringArray   private key setting
  -k, --public-key stringArray    public key setting
      --timeout string            client timeout (default 30s, e.g. 30s, 5m)
  -v, --verbose                   deprecated: enable logrus verbose logging
      --version                   show version
```

### Description

The Open Component Model command line client supports the work with OCM
artifacts, like Common Transport Archive,
Component Repositories, and Component Versions.

Additionally it provides some limited support for the docker daemon, OCI artifacts and
registries.

It can be used in two ways:
- *verb/operation first*: here the sub commands follow the pattern *&lt;verb> &lt;object kind> &lt;arguments>*
- *area/kind first*: here the area and/or object kind is given first followed by the operation according to the pattern
  *[&lt;area>] &lt;object kind> &lt;verb/operation> &lt;arguments>*

The command accepts some top level options, they can only be given before the sub commands.

A configuration according to [ocm configfile](ocm_configfile.md) is read from a <code>.ocmconfig</code> file
located in the <code>HOME</code> directory. With the option <code>--config</code> other
file locations can be specified. If nothing is specified and no file is found at the default
location a default configuration is composed according to known type specific
configuration files.

The following configuration sources are used:
  - The docker configuration file at <code>~/.docker/config.json</code> is
    read to feed in the configured credentials for OCI registries.

  - The npm configuration file at <code>~/.npmrc</code> is
    read to feed in the configured credentials for NPM registries.



With the option <code>--cred</code> it is possible to specify arbitrary credentials
for various environments on the command line. Nevertheless it is always preferable
to use the cli config file.
Every credential setting is related to a dedicated consumer and provides a set of
credential attributes. All this can be specified by a sequence of <code>--cred</code>
options.

Every option value has the format

<center>
    <pre>--cred [:]&lt;attr>=&lt;value></pre>
</center>

Consumer identity attributes are prefixed with the colon ':'. A credential settings
always start with a sequence of at least one identity attributes, followed by a
sequence of credential attributes.
If a credential attribute is followed by an identity attribute a new credential setting
is started.

The first credential setting may omit identity attributes. In this case it is used as
default credential, always used if no dedicated match is found.

For example:

<center>
    <pre>--cred :type=OCIRegistry --cred :hostname=ghcr.io --cred username=mandelsoft --cred password=xyz</pre>
</center>

With the option <code>-X</code> it is possible to pass global settings of the
form

<center>
    <pre>-X &lt;attribute>=&lt;value></pre>
</center>

The <code>--log*</code> options can be used to configure the logging behaviour.
For details see [ocm logging](ocm_logging.md).

There is a quick config option <code>--logkeys</code> to configure simple
tag/realm based condition rules. The comma-separated names build an AND rule.
Hereby, names starting with a slash (<code>/</code>) denote a realm (without
the leading slash). A realm is a slash separated sequence of identifiers. If
the realm name starts with a plus (<code>+</code>) character the generated rule
will match the realm and all its sub-realms, otherwise, only the dedicated
realm is affected. For example <code>/+ocm=trace</code> will enable all log output of the
OCM library.

A tag directly matches the logging tags. Used tags and realms can be found under
topic [ocm logging](ocm_logging.md). The ocm coding basically uses the realm <code>ocm</code>.
The default level to enable is <code>info</code>. Separated by an equal sign (<code>=</code>)
optionally a dedicated level can be specified. Log levels can be (<code>error</code>,
<code>warn</code>, <code>info</code>, <code>debug</code> and <code>trace</code>.
The default level is <code>warn</code>.
The <code>--logconfig*</code> options can be used to configure a complete
logging configuration (yaml/json) via command line. If the argument starts with
an <code>@</code>, the logging configuration is taken from a file.

The value can be a simple type or a JSON/YAML string for complex values
(see [ocm attributes](ocm_attributes.md)). The following attributes are supported:
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

For several options (like <code>-X</code>) it is possible to pass complex values
using JSON or YAML syntax. To pass those arguments the escaping of the used shell
must be used to pass quotes, commas, curly brackets or newlines. for the *bash*
the easiest way to achieve this is to put the complete value into single quotes.

<center>
<code>-X 'mapocirepo={"mode": "shortHash"}'</code>.
</center>

Alternatively, quotes and opening curly brackets can be escaped by using a
backslash (<code>&bsol;</code>).
Often a tagged value can also be substituted from a file with the syntax

<center>
<code>&lt;attr>=@&lt;filepath></code>
</center>

The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>&lt;name>=&lt;filepath></code>. The name is the name
of the key and represents the context is used for (For example the signature
name of a component version)

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.

With <code>--issuer</code> it is possible to declare expected issuer
constraints for public key certificates provided as part of a signature
required to accept the provisioned public key (besides the successful
validation of the certificate). By default, the issuer constraint is
derived from the signature name. If it is not a formal distinguished name,
it is assumed to be a plain common name.

With <code>--ca-cert</code> it is possible to define additional root
certificates for signature verification, if public keys are provided
by a certificate delivered with the signature.

### SEE ALSO



##### Sub Commands

* [ocm <b>add</b>](ocm_add.md)	 &mdash; Add elements to a component repository or component version
* [ocm <b>bootstrap</b>](ocm_bootstrap.md)	 &mdash; bootstrap components
* [ocm <b>check</b>](ocm_check.md)	 &mdash; check components in OCM repository
* [ocm <b>clean</b>](ocm_clean.md)	 &mdash; Cleanup/re-organize elements
* [ocm <b>controller</b>](ocm_controller.md)	 &mdash; Commands acting on the ocm-controller
* [ocm <b>create</b>](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm <b>describe</b>](ocm_describe.md)	 &mdash; Describe various elements by using appropriate sub commands.
* [ocm <b>download</b>](ocm_download.md)	 &mdash; Download oci artifacts, resources or complete components
* [ocm <b>execute</b>](ocm_execute.md)	 &mdash; Execute an element.
* [ocm <b>get</b>](ocm_get.md)	 &mdash; Get information about artifacts and components
* [ocm <b>hash</b>](ocm_hash.md)	 &mdash; Hash and normalization operations
* [ocm <b>install</b>](ocm_install.md)	 &mdash; Install new OCM CLI components
* [ocm <b>list</b>](ocm_list.md)	 &mdash; List information about components
* [ocm <b>set</b>](ocm_set.md)	 &mdash; Set information about OCM repositories
* [ocm <b>show</b>](ocm_show.md)	 &mdash; Show tags or versions
* [ocm <b>sign</b>](ocm_sign.md)	 &mdash; Sign components or hashes
* [ocm <b>transfer</b>](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm <b>verify</b>](ocm_verify.md)	 &mdash; Verify component version signatures
* [ocm <b>version</b>](ocm_version.md)	 &mdash; displays the version



##### Area Overview

* [ocm <b>cache</b>](ocm_cache.md)	 &mdash; Cache related commands
* [ocm <b>credentials</b>](ocm_credentials.md)	 &mdash; Commands acting on credentials
* [ocm <b>oci</b>](ocm_oci.md)	 &mdash; Dedicated command flavors for the OCI layer
* [ocm <b>ocm</b>](ocm_ocm.md)	 &mdash; Dedicated command flavors for the Open Component Model
* [ocm <b>toi</b>](ocm_toi.md)	 &mdash; Dedicated command flavors for the TOI layer


##### Additional Help Topics

* [ocm <b>attributes</b>](ocm_attributes.md)	 &mdash; configuration attributes used to control the behaviour
* [ocm <b>configfile</b>](ocm_configfile.md)	 &mdash; configuration file
* [ocm <b>credential-handling</b>](ocm_credential-handling.md)	 &mdash; Provisioning of credentials for credential consumers
* [ocm <b>logging</b>](ocm_logging.md)	 &mdash; Configured logging keys
* [ocm <b>oci-references</b>](ocm_oci-references.md)	 &mdash; notation for OCI references
* [ocm <b>ocm-accessmethods</b>](ocm_ocm-accessmethods.md)	 &mdash; List of all supported access methods
* [ocm <b>ocm-downloadhandlers</b>](ocm_ocm-downloadhandlers.md)	 &mdash; List of all available download handlers
* [ocm <b>ocm-labels</b>](ocm_ocm-labels.md)	 &mdash; Labels and Label Merging
* [ocm <b>ocm-pubsub</b>](ocm_ocm-pubsub.md)	 &mdash; List of all supported publish/subscribe implementations
* [ocm <b>ocm-references</b>](ocm_ocm-references.md)	 &mdash; notation for OCM references
* [ocm <b>ocm-uploadhandlers</b>](ocm_ocm-uploadhandlers.md)	 &mdash; List of all available upload handlers
* [ocm <b>toi-bootstrapping</b>](ocm_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions
