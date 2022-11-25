## ocm attributes &mdash; Configuration Attributes Used To Control The Behaviour

### Description


The OCM library supports are set of attributes, which can be used to influence
the bahaviour of various functions. The CLI also supports setting of those
attributes using the config file (see [ocm configfile](ocm_configfile.md)) or by
command line options of the main command (see [ocm](ocm.md)).

The following options are available in the currently used version of the
OCM library:
- <code>github.com/mandelsoft/oci/cache</code> [<code>cache</code>]: *string*

  Filesystem folder to use for caching OCI blobs

- <code>github.com/mandelsoft/ocm/compat</code> [<code>compat</code>]: *bool*

  Compatibility mode: Avoid generic local access methods and prefer type specific ones.

- <code>github.com/mandelsoft/ocm/keeplocalblob</code> [<code>keeplocalblob</code>]: *bool*

  Keep local blobs when importing OCI artifacts to OCI registries from <code>localBlob</code>
  access methods. By default they will be expanded to OCI artifacts with the
  access method <code>ociRegistry</code>. If this option is set to true, they will be stored
  as local blobs, also. The access method will still be <code>localBlob</code> but with a nested
  <code>ociRegistry</code> access method for describing the global access.

- <code>github.com/mandelsoft/ocm/ociuploadrepo</code> [<code>ociuploadrepo</code>]: *oci base repository ref*

  Upload local OCI artifact blobs to a dedicated repository.

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

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

