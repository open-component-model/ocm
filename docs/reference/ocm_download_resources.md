## ocm download resources

download resources of a component version

### Synopsis


Download resources of a component version. Resources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
For a single resource to download, this is the file written for the
resource blob. If multiple resources are selected, a directory structure
is written into the given directory for every involved component version
as follows:

<center>
<code>&lt;component>/&lt;version>{/&lt;nested component>/&lt;version>}</code>
</center>

The resource files are named according to the resource identity in the
component descriptor. If this identity is just the resource name, this name
is ised. If additional identity attributes are required, this name is
append by a comma separated list of <code>&lt;name>=&lt>value></code> pairs
separated by a "-" from the plain name. This attribute list is alphabetical
order:

<center>
<code>&lt;resource name>[-[&lt;name>=&lt>value>]{,&lt;name>=&lt>value>}]</code>
</center>



```
ocm download resources [<options>]  <component> {<name> { <key>=<value> }} [flags]
```

### Options

```
  -c, --closure            follow component references
  -h, --help               help for resources
      --lookup string      repository name or spec for closure lookup fallback
  -O, --outfile string     output file or directory
  -o, --output string      output mode ()
  -r, --repo string        repository name or spec
  -s, --sort stringArray   sort fields
```

### SEE ALSO

* [ocm download](ocm_download.md)	 - 

