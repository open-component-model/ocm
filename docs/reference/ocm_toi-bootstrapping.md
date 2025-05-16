---
title: "toi bootstrapping"
menu:
  docs:
    parent: cli-reference
---
## ocm toi-bootstrapping &mdash; Tiny OCM Installer Based On Component Versions

### Description

Tiny OCM Installer (TOI) is a small toolset on top of the Open Component Model.
It provides a possibility to run images taken from a component version with user
configuration and feed them with the content of this component version.
It is some basic mechanism, which can be used to execute simple installation
steps based on content described by the Open Component Model
(see [ocm bootstrap package](ocm_bootstrap_package.md)).

Therefore, a dedicated resource type <code>toiPackage</code> is defined,
which describes an installation package to be handled by TOI.
When calling the [ocm bootstrap package](ocm_bootstrap_package.md) command it is selected by a resource
identity pattern. The first resource in given component version matching the
pattern is used. A possible use case could be to provide different packages for
different environments. The resource can use an identity attribute
<code>platform=&lt;value></code>. By specifying just the platform attribute,
the appropriate package will be chosen.

The bootstrap command uses this package resource to determine a TOI executor
together with executor configuration and additional client specific settings to
describe a dedicated installation.

To do this the package describes dedicated actions that can be executed by
the bootstrap command. Every action (for example <code>install</code>) refers to an
executor, which is executed to perform the action.

An executor is basically an image following the TOI
specification for passing information into the image execution and receiving
results from the execution. An executor specification can be described in two ways:
- it either directly describes a resource of type <code>ociImage</code> or
- it describes a resource of type <code>toiExecutor</code>, which defines
  the image to use and some default settings. It furthermore describes the features
  and requirements of the executor image.

The package describes configuration values for every configured executor as well
as general credentials requirements and required user configuration
which must be passed along with the bootstrap command. The executor specification
may then optionally map this package global settings into executor specific views.

After validation of the input and its mapping to an executor specific format,
finally, a container with the selected executor image is created, that
contains the content of the initial component version in form of a Common
Transport Archive and all the specified configuration data.

The execution of the container may do the needful to achieve the goal of the
requested action and provide some labeled output files, which will be passed
to the caller.

### The <code>toiPackage</code> Resource

This resource describes an installable software package, whose content is
contained in the component version, which contains the package resource.

It is a plain yaml resource with the media types media type <code>application/x-yaml</code>,
<code>text/yaml</code> or

<code>application/vnd.toi.ocm.software.package.v1+yaml</code>) containing
information required to control the instantiation of an executor.

It has the following format:

- **<code>description</code>** (optional) *string*

  A short description of the installation package and some configuration hints.

- **<code>executors</code>** *[]ExecutorSpecification*

- **<code>configTemplate</code>** (optional) *yaml*

  This is a [spiff](https://github.com/mandelsoft/spiff) template used to generate
  The user config that is finally passed to the executor. If no template
  is specified the user parameter input will be processed directly without template.

- **<code>configScheme</code>** (optional) *yaml*

  This is a [JSONSCHEMA](https://json-schema.org/) used to validate the user
  input prior to merging with the template

- **<code>templateLibraries</code>** (optional) *[]ResourceReference*

  This is a list of resources whose content is used as additional stubs
  for the template processing.

- **<code>credentials</code>** (optional) *map[string]CredentialRequest*

  Here the package may request the provisioning of some credentials with a
  dedicated name/purpose and structure. If specified the bootstrap command
  requites the specification of a credentials file providing the information
  how to satisfy those credential requests.

- **<code>additionalResources</code>** (optional) *map[string]AdditionalResource)*

  A set of additional resources specified by an OCM resource reference or
  direct data as byte, string or yaml.
  The key describes the meaning of the resource. The following keys have
  a special meaning:

  - **<code>configFile</code>**: an example template for a parameter file
  - **<code>credentialsFile</code>**: an example template for a credentials file

  Those templates can be downloaded with [ocm bootstrap configuration](ocm_bootstrap_configuration.md).

#### *ExecutorSpecification*

The executor specification describes the available actions and their mapping
to executors. It uses the following fields:

- **<code>actions</code>** *[]string*

  The list of actions this executor can be used for. If nothings is specified
  the executor will be used for all actions. The first matching executor entry
  will be used to execute an action by the bootstrap command

- **<code>resourceRef</code>** *ResourceReference*

  An OCM resource reference describing a component version resource relative to
  the component version containing the package resource.

- **<code>config</code>** (optional) *yaml*

  This is optional static executor config passed to the executor as is. It is
  to describe the set of elements on which the actual execution of the executor
  should work on.

- **<code>parameterMapping</code>** (optional) *spiff yaml*

  This is an optional spiff template used to process the actual package parameter
  set passed by the caller to transform it to the requirements of the actual executor.

  A package has a global parameter setting, but possibly multiple different
  executors for different actions. They might have different requirements/formats
  concerning the parameter input. Therefore, the executor specification allows to
  map the provided user input, accordingly.

- **<code>credentialMapping</code>** (optional) *map[string]string*

  This is an optional mapping to map credential names used by the package
  to the needs of dedicated executors.

  A package has global parameter setting, but possibly multiple different
  executors for different action. They might have different requirements/formats
  concerning the parameter input. There the executor specification allows to
  map the provided user input, accordingly

- **<code>image</code>** (development) *object*

  Instead of a <code>resourceRef</code> it is possible to directly specify an
  absolute image.

  ATTENTION: this is intended for development purposes, ONLY. Do not use it
  for final component versions.

  It has the field <code>ref</code> and the optional field <code>digest</code>.

- **<code>outputs</code>** (optional) *map[string]string*

  This field can be used to map the names of outputs provided by a dedicated
  executor outputs to package outputs.

#### *ResourceReference*

An OCM resource reference describes a resource of a component version. It is
always evaluated relative to the component version providing the resource
that contains the resource reference. It uses the following fields:

- **<code>resourcePath</code>** (optional) *[]Identity*

  This is sequence of reference identities used to follow a chain of
  component version references starting with the actual component version.
  If not specified the specified resource will be taken from the actual
  component version.

- **<code>resource</code>** *Identity*

  This is the identity of the resource in the selected component version.

#### *AdditionalResource*

This field has either the fields of a *ResourceReference* to refer to the
content of an OCM resource or the field:

- **<code>content</code>** *string|[]byte|YAML*

  Either a resource reference or the field <code>content</code> must be given.
  The content field may contain a string or an inline YAML document.
  For larger content the resource reference form should be preferred.

#### *Identity*

An identity specification is a <code>map[string]string</code>. It describes
the identity attributes of a desired resource in a component version.
It always has at least one identity attribute <code>name</code>, which
is the resource name field of the desired resource. If this resource
defines additional identity attributes, the complete set must be specified.

#### Input Mapping for Executors

An optional <code>parameterMapping</code> in the executor section
can be used to process the global package user-specified parameters
to provide specific values expected by the executor.

This is done by a _spiff_ template. Here special functions
are provided to access specific content:

- <code>hasCredentials(string[,string]) bool</code>

  This function can be used to check whether dedicated credentials
  are effectively provided for the actual installation.

  The name is the name of the credentials as described in the credentials
  request section optionally mapped to the name used for the executor
  (field <code>credentialMapping</code>).

  If the second argument is given, it checks for the named property
  in the credential set.

- <code>getCredentials(string[,string]) map[string]string | string</code>

  This functions provides the property set of the provided credentials.

  If the second argument is given, it returns the named property in the
  selected credential set.

  If the property name is an asterisks (<code>*</code>) a single property
  is expected, whose value is returned.

#### User Config vs Executor Config

An executor is typically able to handle a complete class of installations.
It describes a dedicated installation mechanism, but not a dedicated
installation source. Although, there might be specialized images
for dedicated installation sources, in general the idea is to provide
more general executors, for example an helm executor, which is able to
handle any helm chart, not just a dedicated helm deployment.

Because of this, there is a clear separation between an installation specific
configuration, which is provided by the user calling the TOI commands, and
the parameterization of the executor, which is completely specified in the
package.

The task of the package is to represent a dedicated deployment source. As such
it has to provide information to tell the executor what to install, while
the user configuration is used to describe the instance specific settings.

Back to the example of a helm installer executor, the executor config contained
in the package resource describes the helm chart, which should be installed
and the way how the user input is mapped to chart values. Here, also the
localizations are described in an executor specific way.

Therefore, an executor expects a dedicated configuration format, which can be
specified in the executor resource in form of a JSON scheme.

The package then may provide a package specific scheme for the instance
configuration. This value-set is dependent on the installation source (the helm
chart in this example).

For further details you have to refer to the dedicated executor and package
definitions.


### The <code>toiExecutor</code> Resource
Instead of directly describing an image resource in the package file, it is
possible to refer to a resource of type toiExecutor. This
is a yaml file with the media type <code>application/x-yaml</code>,
<code>text/yaml</code> or
<code>application/vnd.toi.ocm.software.package.v1+yaml</code>) containing
common information about the executor. If this flavor is used by the package,
this information is used to validate settings in the package specification.

It has the following format:

- **<code>imageRef</code>** *ResourceReference*

  This field reference the image resource relative to the component version
  providing the executor resource

- **<code>configTemplate</code>** (optional) *yaml*

  This a [spiff](https://github.com/mandelsoft/spiff) template used to generate
  The executor config from the package specification that is finally passed to
  the executor. If no template is specified the executor config specified in
  the package will be processed directly without template.

- **<code>configScheme</code>** (optional) *yaml*

  This is a [JSONSCHEMA](https://json-schema.org/) used to validate the executor
  config from the package prior to merging with the template

- **<code>templateLibraries</code>** (optional) *[]ResourceReference*

  This is a list of resources whose content is used as additional stubs
  for the template processing.

- **<code>credentials</code>** (optional) *map[string]CredentialRequest*

  Here the executor may request the provisioning of some credentials with a
  dedicated name/purpose and structure. If specified it will be propagated
  to a using package. It this uses an own credentials section, this one
  will be filtered and checked for the actual executor.

- **<code>outputs</code>** (optional) *map[string]OutputSpecification*

  This field can be used to describe the provided outputs of this executor.
  The *OutputSpecification* contains only the field <code>description</code>,
  so far. It is intended to be extended to contain further information to more
  formally describe the type of output.

- **<code>image</code>** (development) *object*

  Instead of an <code>imageRef</code> it is possible to directly specify an
  absolute image.

  ATTENTION: this is intended for development purposes, ONLY. Do not use it
  for final component versions.

  It has the field <code>ref</code> and the optional field <code>digest</code>.

### Client Parameters

Common to all executors a parameter file can be provided by the caller. The
package specification may provide a [spiff template](https://github.com/mandelsoft/spiff)
for this parameter file. It can be used, for example to provide useful defaults.
The actually provided content is merged with this template.

To validate user configuration a JSON scheme can be provided. The user input is
validated first against this scheme before the actual merge is done.

### Credentials

Additionally credentials can be requested to be provided by a client.
This is done with the <code>credentials</code> field. It is a map
of credentials names and their meaning and/or handling.

It uses the following fields:

- **<code>description</code>** *string*

  This field should describe the purpose of the credential.

- **<code>properties</code>** *map[string]string*

  This field should describe the used credential fields

- **<code>consumerId</code>** *map[string]*

  This field can be used to optionally define a consumer id that should be set
  in the OCM support library, if used by the executor. At least the field
  <code>type</code> and one additional field must be set.

Credentials are provided in an ocm config file (see [ocm configfile](ocm_configfile.md)).
It uses a memory credential repository with the name <code>default</code>
to store the credentials under the given name. Additionally appropriate
consumer ids will be propagated, if requested in the credentials request config.

### Executor Image Contract

The executor image is called with the action as additional argument. It is
expected that is defines a default entry point and a potentially empty list of
standard arguments.

It is called with two arguments:
- name of the action to execute
- identity of the component version containing the package the executor
  is executed for.

  This can be used to access the component descriptor to get access to
  further described resources in the executor config

The container used to execute the executor image gets prepared a standard
filesystem structure used to provide all the executor inputs before the
execution and reading provided executor outputs after the execution.

<pre>
/
└── toi
    ├── inputs
    │   ├── config      configuration from package specification
    │   ├── ocmrepo     OCM filesystem repository containing the complete
    │   │               component version of the package
    │   └── parameters  merged complete parameter file
    ├── outputs
    │   ├── &lt;out>       any number of arbitrary output data provided
    │   │               by executor
    │   └── ...
    └── run             good practice: typical location for the executed command
</pre>

After processing it is possible to return named outputs. The name of an output
must be a filename. The executor section in the package specification maps those
files to logical outputs in the <code>outputs</code> section.

<center>
  &lt;file name by executor> -> &lt;logical output name>
</center>

Basically the output may contain any data, but is strongly recommended
to use yaml or json files, only. This enables further formal processing
by the TOI toolset.

### Examples

```yaml
description: |
  This package is just an example.
executors:
  - actions:
    - install
    resourceRef:
      resource:
        name: installerimage
    config:
      level: info
#   parameterMapping:  # optional spiff mapping of Package configuration to
#      ....            # executor parameters
    outputs:
       test: bla
credentials:
  target:
    description: kubeconfig for target kubernetes cluster
    consumerId:
      type: Kubernetes
      purpose: target
configTemplate:
  parameters:
    username: admin
    password: (( &merge ))
configScheme:
  type: object
  required:
    - parameters
  additionalProperties: false
  properties:
    parameters:
      type: object
      required:
      - password
      additionalProperties: false
      properties:
        username:
          type: string
        password:
          type: string
additionalResources:
  configFile:
    resource:
      name: config-file
```

### SEE ALSO

#### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client



##### Additional Links

* [<b>ocm bootstrap package</b>](ocm_bootstrap_package.md)	 &mdash; bootstrap component version
* [<b>ocm bootstrap configuration</b>](ocm_bootstrap_configuration.md)	 &mdash; bootstrap TOI configuration files
* [<b>ocm configfile</b>](ocm_configfile.md)	 &mdash; configuration file

