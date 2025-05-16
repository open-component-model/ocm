---
title: "transfer componentarchive"
menu:
  docs:
    parent: transfer
---
## ocm transfer componentarchive &mdash; (DEPRECATED) - Please Use Commontransportarchive Instead

### Synopsis

```bash
ocm transfer componentarchive [<options>] <source> <target>
```

#### Aliases

```text
componentarchive, comparch, ca
```

### Options

```text
  -L, --copy-local-resources   transfer referenced local resources by-value
  -V, --copy-resources         transfer referenced resources by-value
      --copy-sources           transfer referenced sources by-value
      --enforce                enforce transport as if target version were not present
  -h, --help                   help for componentarchive
      --lookup stringArray     repository name or spec for closure lookup fallback
      --no-update              don't touch existing versions in target
  -f, --overwrite              overwrite existing component versions
  -t, --type string            archive format (directory, tar, tgz) (default "directory")
```

### Description

Transfer a component archive to some component repository. This might
be a CTF Archive or a regular repository.
If the type CTF is specified the target must already exist, if CTF flavor
is specified it will be created if it does not exist.

Besides those explicitly known types a complete repository spec might be configured,
either via inline argument or command configuration file and name.


The <code>--type</code> option accepts a file format for the
target archive to use. It is only evaluated if the target
archive does not exist yet. The following formats are supported:
- directory
- tar
- tgz

The default format is <code>directory</code>.

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.


With the option <code>--no-update</code> existing versions in the target
repository will not be touched at all. An additional specification of the
option <code>--overwrite</code> is ignored. By default, updates of
volatile (non-signature-relevant) information is enabled, but the
modification of non-volatile data is prohibited unless the overwrite
option is given.


If the option <code>--overwrite</code> is given, component versions in the
target repository will be overwritten, if they already exist, but with different digest.
If the option <code>--enforce</code> is given, component versions in the
target repository will be transported as if they were not present on the target side,
regardless of their state (this is independent on their actual state, even identical
versions are re-transported).


If the option <code>--copy-resources</code> is given, all referential
resources will potentially be localized, mapped to component version local
resources in the target repository. If the option <code>--copy-local-resources</code>
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.


If the option <code>--copy-sources</code> is given, all referential
sources will potentially be localized, mapped to component version local
resources in the target repository.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.

### SEE ALSO

#### Parents

* [ocm transfer](ocm_transfer.md)	 &mdash; Transfer artifacts or components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client

