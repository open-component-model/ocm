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
```

### Description


The Open Component Model command line client support the work with OCM
artefacts, like Component Archives, Common Transport Archive,  
Component Repositories, and component versions.

Additionally it provides some limited support for the docker daemon, OCI artefacts and
registries.

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
- <code>github.com/mandelsoft/ocm/signing</code>: *bool*
  Public and private Key settings.
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

### SEE ALSO



##### Sub Commands

* [ocm <b>add</b>](ocm_add.md)	 - Add resources or sources to a component archive
* [ocm <b>cache</b>](ocm_cache.md)	 - Cache related commands
* [ocm <b>clean</b>](ocm_clean.md)	 - Cleanup/re-organize elements
* [ocm <b>componentarchive</b>](ocm_componentarchive.md)	 - Commands acting on component archives
* [ocm <b>componentversions</b>](ocm_componentversions.md)	 - Commands acting on components
* [ocm <b>create</b>](ocm_create.md)	 - Create transport or component archive
* [ocm <b>describe</b>](ocm_describe.md)	 - Describe artefacts
* [ocm <b>download</b>](ocm_download.md)	 - Download oci artefacts, resources or complete components
* [ocm <b>get</b>](ocm_get.md)	 - Get information about artefacts and components
* [ocm <b>oci</b>](ocm_oci.md)	 - Dedicated command flavors for the OCI layer
* [ocm <b>ocm</b>](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm <b>references</b>](ocm_references.md)	 - Commands related to component references in component versions
* [ocm <b>resources</b>](ocm_resources.md)	 - Commands acting on component resources
* [ocm <b>show</b>](ocm_show.md)	 - Show tags or versions
* [ocm <b>sign</b>](ocm_sign.md)	 - Sign components
* [ocm <b>sources</b>](ocm_sources.md)	 - Commands acting on component sources
* [ocm <b>transfer</b>](ocm_transfer.md)	 - Transfer artefacts or components
* [ocm <b>verify</b>](ocm_verify.md)	 - Verify component version signatures
* [ocm <b>version</b>](ocm_version.md)	 - displays the version



##### Additional Help Topics

* [ocm <b>configfile</b>](ocm_configfile.md)	 - configuration file
* [ocm <b>oci-references</b>](ocm_oci-references.md)	 - notation for OCI references
* [ocm <b>ocm-references</b>](ocm_ocm-references.md)	 - notation for OCM references

