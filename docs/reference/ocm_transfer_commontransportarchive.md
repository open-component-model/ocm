## ocm transfer commontransportarchive &mdash; Transfer Transport Archive

### Synopsis

```
ocm transfer commontransportarchive [<options>] <ctf> <target>
```

### Options

```
  -V, --copy-resources      transfer referenced resources by-value
  -h, --help                help for commontransportarchive
  -f, --overwrite           overwrite existing component versions
      --script string       config name of transfer handler script
  -s, --scriptFile string   filename of transfer handler script
  -t, --type string         archive format (directory, tar, tgz) (default "directory")
```

### Description


Transfer content of a Common Transport Archive to the given target repository.

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
resources in the target repository.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.

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

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artefacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

