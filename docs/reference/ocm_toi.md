## ocm toi &mdash; Dedicated Command Flavors For The TOI Layer

### Synopsis

```
ocm toi [<options>] <sub command> ...
```

### Options

```
  -h, --help   help for toi
```

### Description


TOI is an abbreviation for (T)iny (O)CM (I)nstallation. It is a simple
application framework on top of the Open Component Model, that can
be used to describe image based installation executors and installation
packages (see topic [ocm toi bootstrapping](ocm_toi_bootstrapping.md) in form of resources
with a dedicated type. All involved resources are hereby taken from a component
version of the Open Component Model, which supports all the OCM features, like
transportation.

The framework consists of a generic bootstrap command
([ocm toi bootstrap componentversions](ocm_toi_bootstrap_componentversions.md)) and an arbitrary set of image
based executors, that are executed in containers and fed with the required
installation data by th generic command.


### SEE ALSO

##### Parents

* [ocm](ocm.md)	 &mdash; Open Component Model command line client


##### Sub Commands

* ocm toi <b>bootstrap</b>	 &mdash; bootstrap components
* ocm toi <b>configuration</b>	 &mdash; TOI Commands acting on config
* ocm toi <b>describe</b>	 &mdash; describe packages
* ocm toi <b>package</b>	 &mdash; TOI Commands acting on components



##### Additional Help Topics

* [ocm toi <b>bootstrapping</b>](ocm_toi_bootstrapping.md)	 &mdash; Tiny OCM Installer based on component versions
* [ocm toi <b>ocm-references</b>](ocm_toi_ocm-references.md)	 &mdash; notation for OCM references


##### Additional Links

* [<b>ocm toi bootstrap componentversions</b>](ocm_toi_bootstrap_componentversions.md)

