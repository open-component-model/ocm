---
title: "bootstrap package"
url: "/docs/cli-reference/bootstrap/package/"
sidebar:
  collapsed: true
---

## ocm bootstrap package &mdash; Bootstrap Component Version

### Synopsis

```bash
ocm bootstrap package [<options>] <action> {<component-reference>} {<resource id field>}
```

#### Aliases

```text
package, pkg, componentversion, cv, component, comp, c
```

### Options

```text
      --config stringToString   driver config (default [])
  -C, --create-env string       create local filesystem contract to call executor command locally
  -c, --credentials string      credentials file
  -h, --help                    help for package
      --lookup stringArray      repository name or spec for closure lookup fallback
  -o, --outputs string          output file/directory
  -p, --parameters string       parameter file
      --repo string             repository name or spec
```

### Description

Use the simple TOI bootstrap mechanism to execute actions for a TOI package resource
based on the content of an OCM component version and some command input describing
the dedicated installation target.

The package resource must have the type <code>toiPackage</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic [ocm toi-bootstrapping](ocm_toi-bootstrapping.md).

This resource finally describes an executor image, which will be executed in a
container with the installation source and (instance specific) user settings.
The container is just executed, the framework make no assumption about the
meaning/outcome of the execution. Therefore, any kind of actions can be described and
issued this way, not only installation handling.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no output file is provided, the yaml representation of the outputs are
printed to standard out. If the output file is a directory, for every output a
dedicated file is created, otherwise the yaml representation is stored to the
file.

If no credentials file name is provided (option -c) the file
<code>TOICredentials</code> is used, if present. If no parameter file name is
provided (option -p) the file <code>TOIParameters</code> is used, if present.

Using the credentials file it is possible to configure credentials required by
the installation package or executor. Additionally arbitrary consumer ids
can be forwarded to executor, which might be required by accessing blobs
described by external access methods.

The credentials file uses the following yaml format:
- <code>credentials</code> *map[string]CredentialsSpec*

  The resolution of credentials requested by the package (by name).

- <code>forwardedConsumers</code> *[]ForwardSpec* (optional)

  An optional list of consumer specifications to be forwarded to the OCM
  configuration provided to the executor.

The *CredentialsSpec* uses the following format:

- <code>consumerId</code> *map[string]string*

  The consumer id used to look up the credentials.

- <code>consumerType</code> *string* (optional) (default: partial)

  The type of the matcher used to match the consumer id.

- <code>reference</code> *yaml*

  A generic credential specification as used in the ocm config file.

- <code>credentials</code> *map[string]string*

  Direct credential fields.

One of <code>consumerId</code>, <code>reference</code> or <code>credentials</code> must be configured.

The *ForwardSpec* uses the following format:

- <code>consumerId</code> *map[string]string*

  The consumer id to be forwarded.

- <code>consumerType</code> *string* (optional) (default: partial)

  The type of the matcher used to match the consumer id.

If provided by the package it is possible to download template versions
for the parameter and credentials file using the command [ocm bootstrap configuration](ocm_bootstrap_configuration.md).

Using the option <code>--config</code> it is possible to configure options
for the execution environment (so far only docker is supported).
The following options are possible:
  - <code>CLEANUP_CONTAINERS</code>
  - <code>DOCKER_DRIVER_QUIET</code>
  - <code>NETWORK_MODE</code>
  - <code>PULL_POLICY</code>
  - <code>USERNS_MODE</code>


Using the option <code>--create-env  &lt;toi root folder></code> it is possible to
create a local execution environment for an executor according to the executor
image contract (see [ocm toi-bootstrapping](ocm_toi-bootstrapping.md)). If the executor executable is
built based on the toi executor support package, the executor can then be called
locally with

<center>
    <pre>&lt;executor> --bootstraproot &lt;given toi root folder></pre>
</center>


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

Using the JSON variant any repository types supported by the
linked library can be used:

OCI Repository types (using standard component repository to OCI mapping):

  - <code>CommonTransportFormat</code>: v1
  - <code>OCIRegistry</code>: v1
  - <code>oci</code>: v1
  - <code>ociRegistry</code>

\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.

### Examples

```bash
$ ocm toi bootstrap package ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
```

### SEE ALSO

#### Parents

* [ocm bootstrap](ocm_bootstrap.md)	 &mdash; bootstrap components
* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm toi-bootstrapping</b>](ocm_toi-bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions
* [<b>ocm bootstrap configuration</b>](ocm_bootstrap_configuration.md)	 &mdash; bootstrap TOI configuration files

