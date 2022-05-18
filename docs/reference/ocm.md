## ocm

Open Component Model command line client

### Synopsis

```
ocm [<options>] <sub command> ...
```

### Options

```
      --config string      configuration file
  -C, --cred stringArray   credential setting
  -h, --help               help for ocm
```

### Description


The Open Component Model command line client support the work with OCM
artefacts, like Component Archives, Common Transport Archive,  
Component Repositories, and component versions.

Additionally it provides some limited support for the docker daemon, OCI artefacts and
registries.

With the open <code>--cred</code> it is possible to specify arbitrary credentials
for various environments on the command line. Nevertheless it is always preferrable
to use the cli config file.
Every credential setting is related to a dedicated consumer and provides a set of
credential attributes. All this can be specified by a sequence of <code>--cred</code>
options. 

Every option value has the format
<center>
<code>--cred [:]&lt;attr>=&lt;value></code>
</center>

Consumer identity attribues are prefixed with the colon (:). A credential settings
always start with a sequence of at least one identity attributes, followed by a
sequence of credential attributes.
If a credential attribute is followed by an identity attribute a new credential setting
is started.

The first credential setting may omit identity attributes. In this case it is used as
default credential, always used if no dedicated match is found.

For example:
<center>
<code>--cred :type=ociRegistry --cred hostname:ghcr.io --cred usename=mandelsoft --cred password=xyz </code>
</center>


### SEE ALSO

* [ocm <b>add</b>](ocm_add.md)	 - Add resources or sources to a component archive
* [ocm <b>componentarchive</b>](ocm_componentarchive.md)	 - Commands acting on component archives
* [ocm <b>components</b>](ocm_components.md)	 - Commands acting on components
* [ocm <b>create</b>](ocm_create.md)	 - Create transport or component archive
* [ocm <b>describe</b>](ocm_describe.md)	 - Describe artefacts
* [ocm <b>download</b>](ocm_download.md)	 - Download oci artefacts, resources or complete components
* [ocm <b>get</b>](ocm_get.md)	 - Get information about artefacts and components
* [ocm <b>oci</b>](ocm_oci.md)	 - Dedicated command flavors for the OCI layer
* [ocm <b>ocm</b>](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm <b>references</b>](ocm_references.md)	 - Commands related to component references in component versions
* [ocm <b>resources</b>](ocm_resources.md)	 - Commands acting on component resources
* [ocm <b>show</b>](ocm_show.md)	 - Show tags or versions
* [ocm <b>sources</b>](ocm_sources.md)	 - Commands acting on component sources
* [ocm <b>transfer</b>](ocm_transfer.md)	 - Transfer artefacts or components
* [ocm <b>version</b>](ocm_version.md)	 - displays the version

