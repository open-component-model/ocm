## ocm transfer componentversions &mdash; Transfer Component Version

### Synopsis

```
ocm transfer componentversions [<options>] {<component-reference>} <target>
```

### Options

```
  -h, --help                 help for componentversions
      --lookup stringArray   repository name or spec for closure lookup fallback
  -f, --overwrite            overwrite existing component versions
  -r, --recursive            follow component reference nesting
      --repo string          repository name or spec
  -V, --resourcesByValue     transfer resources by-value
      --script string        config name of transfer handler script
  -s, --scriptFile string    filename of transfer handler script
  -t, --type string          archive format (directory, tar, tgz) (default "directory")
```

### Description


Transfer all component versions specified to the given target repository.
If only a component (instead of a component version) is specified all versions
are transferred.

If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center>
    <pre>&lt;component>[:&lt;version>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as located OCM component version references:

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</pre>
</center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center>
    <pre>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</pre>
</center>

The <code>--repo</code> option takes an OCM repository specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository type supported by the 
linked library can be used:

Dedicated OCM repository types:
- `ComponentArchive`

OCI Repository types (using standard component repository to OCI mapping):
- `ArtefactSet`
- `CommonTransportFormat`
- `DockerDaemon`
- `Empty`
- `OCIRegistry`
- `oci`
- `ociRegistry`

The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
- directory
- tar
- tgz
The default format is <code>directory</code>.

With the option <code>--recursive</code> the complete reference tree of a component reference is traversed.

If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.

It the option <code>--overwrite</code> is given, component version in the
target repository will be overwritten, if they already exist.

It the option <code>--resourcesByValue</code> is given, all referential 
resources will potentially be localized, mapped to component version local
resources in the target repository.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.

It is possible to use a dedicated transfer script based on spiff.
The option <code>--scriptFile</code> can be used to specify this script
by a file name. With <code>--script</code> it can be taken from the 
CLI config using an entry of the following format:

<pre>
type: scripts.ocm.config.ocm.gardener.cloud
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

$ ocm transfer components -t tgz ghcr.io/mandelsoft/kubelink ctf.tgz
$ ocm transfer components -t tgz --repo OCIRegistry:ghcr.io mandelsoft/kubelink ctf.tgz

```

### SEE ALSO

##### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artefacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

