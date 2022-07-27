## ocm ocm ocm-bootstrapping &mdash; Installation Bootstrapping Based On Component Versions

### Description


The OCM command line tool and library provides some basic bootstrapping mechanism, which can be used to
execute simple installation steps based on content described by the Open Component Model.

Therefore a dedicated resource type <code>ocmInstaller</code> is defined.
It describes a yaml file containing information about the bootstrapping mechanism:

The most important section is the <code>executors</code> sections. It describes a list of
executor definitions and the actions the executor should be used for.
The actions are arbitrary string that can be defined by the provider of the bootstrapping.
If the <code>actions</code> list omitted the executor will be used for all actions
not already accepted by an earlier entry in the list.

Every executor describes an executor image, which is taken from the component version
the vootstrapping specification is taken from. Such a ref always describes a resource
identity (minimal attribute set consists of the resource's <code>name</code> attribute). Additional
attributes are possible, to achive a unique identification of the resource.
Optionally a <code>referencePath</code> can be given, if the resource is located in some 
aggregated component version. This is just a list af identity sets uniquely identifying a nested
component version reference.

Every executor may define some config that is passed to the image execution.
After processing it is possible to return named outputs. The name of an output must be a filename.
The bootstrapp specification maps those file to logical outputs in the <code>outputs</code> section.

<center>
  &lt;provide file name by executor> -> &lt;logical output name for the bootstrap command>
</center>

Common to all executors a parameter file can be provided by the caller. The specification may 
provide a [spiff template}(https://github.com/mandelsoft/spiff) for this parameter file.
The actually provided content is merged with this template.

To validate user configuration a JSON scheme can be provided. The user input is valided first
against this scheme before the actual merhe is done.

### Image Binding

The executor image is called with the action as additional argument. It is expected 
that is defines a default entry point and a potentially empty list of standard arguments.

The other inputs and outputs are provided by a filesystem structure:
<pre>
/
└── ocm
    ├── inputs
    │   ├── config      config info from bootstrap specification
    │   ├── ocmrepo     OCM filesystem repository containing the complement component version
    │   └── parameters  merged complete parameter file
    ├── outputs
    │   ├── &lt;out>       any number of arbitrary output data provided by executor
    │   └── ...         
    └── run             typical location for the executed command
</pre>

The output names are mapped according the bootstrap specification resource.


### Examples

```

executors:
  - actions:
    - install
    imageResourceRef:
      resource:
        name: installerimage
    config:
      level: info
    outputs:
       test: bla
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

* [ocm ocm](ocm_ocm.md)	 - Dedicated command flavors for the Open Component Model
* [ocm](ocm.md)	 - Open Component Model command line client

