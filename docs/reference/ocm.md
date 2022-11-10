## ocm &mdash; Open Component Model Command Line Client

### Synopsis

```
ocm [<options>] <sub command> ...
```

### Options

```
  -X, --attribute stringArray   attribute setting
      --config string           configuration file
  -C, --cred stringArray        credential setting
  -h, --help                    help for ocm
      --logconfig string        log config
  -L, --logfile string          set log file
  -l, --loglevel string         set log level
  -v, --verbose                 enable verbose logging
```

### Description


The Open Component Model command line client support the work with OCM
artefacts, like Component Archives, Common Transport Archive,  
Component Repositories, and component versions.

Additionally it provides some limited support for the docker daemon, OCI artefacts and
registries.

It can be used in two ways:
- *verb/operation first*: here the sub commands follow the pattern *&lt;verb> &lt;object kind> &lt;arguments>*
- *area/kind first*: here the area and/or object kind is given first followed by the operation according to the pattern
  *[&lt;area>] &lt;object kind> &lt;verb/operation> &lt;arguments>*

The command accepts some top level options, they can only be given before the sub commands.

With the option <code>--cred</code> it is possible to specify arbitrary credentials
for various environments on the command line. Nevertheless it is always preferrable
to use the cli config file.
Every credential setting is related to a dedicated consumer and provides a set of
credential attributes. All this can be specified by a sequence of <code>--cred</code>
options. 

Every option value has the format

<center>
    <pre>--cred [:]&lt;attr>=&lt;value></pre>
</center>

Consumer identity attributes are prefixed with the colon (:). A credential settings
always start with a sequence of at least one identity attributes, followed by a
sequence of credential attributes.
If a credential attribute is followed by an identity attribute a new credential setting
is started.

The first credential setting may omit identity attributes. In this case it is used as
default credential, always used if no dedicated match is found.

For example:

<center>
    <pre>--cred :type=ociRegistry --cred hostname=ghcr.io --cred usename=mandelsoft --cred password=xyz</pre>
</center>

With the option <code>-X</code> it is possible to pass global settings of the 
form 

<center>
    <pre>-X &lt;attribute>=&lt;value></pre>
</center>

The value can be a simple type or a json string for complex values. The following
attributes are supported:
- <code>github.com/mandelsoft/oci/cache</code> [<code>cache</code>]: *string*

  Filesystem folder to use for caching OCI blobs

- <code>github.com/mandelsoft/ocm/compat</code> [<code>compat</code>]: *bool*

  Compatibility mode: Avoid generic local access methods and prefer type specific ones.

- <code>github.com/mandelsoft/ocm/keeplocalblob</code> [<code>keeplocalblob</code>]: *bool*

  Keep local blobs when importing OCI artefacts to OCI registries from <code>localBlob</code>
  access methods. By default they will be expanded to OCI artefacts with the
  access method <code>ociRegistry</code>. If this option is set to true, they will be stored
  as local blobs, also. The access method will still be <code>localBlob</code> but with a nested
  <code>ociRegistry</code> access method for describing the global access.

- <code>github.com/mandelsoft/ocm/ociuploadrepo</code> [<code>ociuploadrepo</code>]: *oci base repository ref*

  Upload local OCI artefact blobs to a dedicated repository.

- <code>github.com/mandelsoft/ocm/plugindir</code> [<code>plugindir</code>]: *plugin directory*

  Directory to look for OCM plugin executables.

- <code>github.com/mandelsoft/ocm/signing</code>: *JSON*

  Public and private Key settings given as JSON document with the following
  format:
  
  <pre>
  {
    "publicKeys"": [
       "&lt;provider>": {
         "data": ""&lt;base64>"
       }
    ],
    "privateKeys"": [
       "&lt;provider>": {
         "path": ""&lt;file path>"
       }
    ]
  </pre>
  
  One of following data fields are possible:
  - <code>data</code>:       base64 encoded binary data
  - <code>stringdata</code>: plain text data
  - <code>path</code>:       a file path to read the data from

- <code>github.com/mandelsoft/tempblobcache</code> [<code>blobcache</code>]: *string* Foldername for temporary blob cache

  The temporary blob cache is used to accessing large blobs from remote sytems.
  The are temporarily stored in the filesystem, instead of the memory, to avoid
  blowing up the memory consumption.

### SEE ALSO



##### Sub Commands

* [ocm <b>add</b>](ocm_add.md)	 &mdash; Add resources or sources to a component archive
* [ocm <b>bootstrap</b>](ocm_bootstrap.md)	 &mdash; bootstrap components
* [ocm <b>clean</b>](ocm_clean.md)	 &mdash; Cleanup/re-organize elements
* [ocm <b>create</b>](ocm_create.md)	 &mdash; Create transport or component archive
* [ocm <b>describe</b>](ocm_describe.md)	 &mdash; Describe artefacts
* [ocm <b>download</b>](ocm_download.md)	 &mdash; Download oci artefacts, resources or complete components
* [ocm <b>get</b>](ocm_get.md)	 &mdash; Get information about artefacts and components
* [ocm <b>show</b>](ocm_show.md)	 &mdash; Show tags or versions
* [ocm <b>sign</b>](ocm_sign.md)	 &mdash; Sign components
* [ocm <b>transfer</b>](ocm_transfer.md)	 &mdash; Transfer artefacts or components
* [ocm <b>verify</b>](ocm_verify.md)	 &mdash; Verify component version signatures
* [ocm <b>version</b>](ocm_version.md)	 &mdash; displays the version



##### Additional Help Topics

* [ocm <b>attributes</b>](ocm_attributes.md)	 &mdash; configuration attributes used to control the behaviour
* [ocm <b>configfile</b>](ocm_configfile.md)	 &mdash; configuration file
* [ocm <b>oci-references</b>](ocm_oci-references.md)	 &mdash; notation for OCI references
* [ocm <b>ocm-references</b>](ocm_ocm-references.md)	 &mdash; notation for OCM references
* [ocm <b>toi-bootstrapping</b>](ocm_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions
