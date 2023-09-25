## ocm transfer commontransportarchive &mdash; Transfer Transport Archive

### Synopsis

```
ocm transfer commontransportarchive [<options>] <ctf> <target>
```

### Options

```
  -L, --copy-local-resources      transfer referenced local resources by-value
  -V, --copy-resources            transfer referenced resources by-value
  -h, --help                      help for commontransportarchive
      --lookup stringArray        repository name or spec for closure lookup fallback
  -f, --overwrite                 overwrite existing component versions
  -r, --recursive                 follow component reference nesting
      --script string             config name of transfer handler script
  -s, --scriptFile string         filename of transfer handler script
  -E, --stop-on-existing          stop on existing component version in target repository
  -t, --type string               archive format (directory, tar, tgz) (default "directory")
      --uploader <name>=<value>   repository uploader (<name>:<artifact type>[:<media type>]=<JSON target config) (default [])
```

### Description


Transfer content of a Common Transport Archive to the given target repository.


With the option <code>--recursive</code> the complete reference tree of a component reference is traversed.

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.


It the option <code>--overwrite</code> is given, component version in the
target repository will be overwritten, if they already exist.


It the option <code>--copy-resources</code> is given, all referential 
resources will potentially be localized, mapped to component version local
resources in the target repository. It the option <code>--copy-local-resources</code> 
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.


It the option <code>--stop-on-existing</code> is given together with the <code>--recursive</code>
option, the recursion is stopped for component versions already existing in the 
target repository. This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.


If the <code>--uploader</code> option is specified, appropriate uploaders
are configured for the transport target. It has the following format

<center>
    <pre>&lt;name>:&lt;artifact type>:&lt;media type>=&lt;yaml target config></pre>
</center>

The uploader name may be a path expression with the following possibilities:
- <code>ocm/ociRegistry</code>: oci Registry upload for local OCI artifact blobs.
  The media type is optional. If given ist must be an OCI artifact media type.
- <code>plugin/<plugin name>[/<uploader name]</code>: uploader provided by plugin.


It is possible to use a dedicated transfer script based on spiff.
The option <code>--scriptFile</code> can be used to specify this script
by a file name. With <code>--script</code> it can be taken from the 
CLI config using an entry of the following format:

<pre>
type: scripts.ocm.config.ocm.software
scripts:
  &lt;name>: 
    path: &lt;filepath> 
    script:
      &lt;scriptdata>
</pre>

Only one of the fields <code>path</code> or <code>script</code> can be used.

If no script option is given and the cli config defines a script <code>default</code>
this one is used.


### Examples

```
$ ocm transfer ctf ctf.tgz ghcr.io/mandelsoft/components
```

### SEE ALSO

##### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

