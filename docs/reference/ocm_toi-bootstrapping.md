## ocm toi-bootstrapping &mdash; Tiny OCM Installer Based On Component Versions

### Description


TOI is a small toolset on top of the Open Component Model. It provides
a possibility to run images taken from a component version with user
configuration and feed them with the content of this component version.
It is some basic mechanism which can be used to execute simple installation
steps based on content described by the Open Component Model
(see [ocm bootstrap componentversions](ocm_bootstrap_componentversions.md)).

Therefore, a dedicated resource type <code>toiPackage</code> is defined.
It is selected by a resource identity pattern. The first resource matching the pattern
is used. A possible use case could be to provide different packages for
different environments. The resource can use an identity attribute
<code>platform=&lt;value></code>. By specifying just the platform attribute,
the appropriate package will be chosen.

The bootstrap command uses this resource to determine a TOI executor together
executor configuration and additional client specific settings to describe
a dedicated installation.

To do this the package describes dedicated actions that can be executed by
the bootstrap command. Every action refers to an executor, which is executed
to perform the action. Finally, an executor is an image following the TOI
specification for passing information into the image execution and receiving
results from the execution. Such an image is described in two ways:
- it either describes a resource of type <code>ociImage</code> or
- it describes a resource of type <code>toiExecutor</code>, which defines
  the image to use and some default settings and further describes the features
  and requirements of the executor image.

The package described credentials requirements and required user configuration
which must passed along with the bootstrap command. After validation of the
input finally a container with the selected executor image is created, that
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

- **<code>executors</code>** *[]ExecutorSpecification*

- **<code>configTemplate</code>** (optional) *yaml*

  This a [spiff](https://github.com/mandelsoft/spiff) template used to generate
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

#### *ExecutorSpecification*

The executor specification describes the available actions and their mapping
to executors. It uses the following fields:

- **<code>actions</code>** *[]string*

  The list of actions this executor can be used for. If nothings is specified
  the executor will be used for all actions. The first matching executor entry
  will be used to execute an action by the bootstrap command

- **<code>resourceRef</code>** *[]ResourceReference*

  An OCM resource reference describing a component version resource relative to
  the component version containing the package resource.

- **<code>config</code>** (optional) *yaml*

  This is optional static executor config passed to the executor as is. It is
  to describe the set of elements on which the actual execution of the executor
  should work on.

- **<code>parameterMapping</code>** (optional) *spiff yaml*

  This is an optional spiff template used to process the actual parameter set
  passed by the caller to transform it to the requirements of the actual executor.

  A package has global parameter setting, but possibly multiple different
  executors for different action. They might have different requirements/formats
  concerning the parameter input. There the executor specification allows to
  map the provided user input, accordingly

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

#### *Identity*

An identity specification is a <code>map[string]string</code>. It describes
the identity attributes of a desired resource in a a component version.
It always has at least one identity attribute <code>name</code>, which
is the resource name field of the desired resource. If this resource
defines additional identity attributes, the complete set must be specified.

### The <code>toiExecutor</code> Resource

Instead of directly describing an image resource i the package file, it is
possible to refer to a resource of type toiExecutor. This
is a yaml file with the media type <code>application/x-yaml</code>,
<code>text/yaml</code> or 
<code>application/vnd.toi.ocm.software.package.v1+yaml</code>) containing
common information about the executor executor. If used by the package,
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
  will be filtered and checked for the the actual executor.

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
    │   ├── config      config info from package specification
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

```

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

```

### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client

