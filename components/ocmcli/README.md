# Example ocmcli

This folder contains an example how to build a component version for the OCM cli.


# Building

You can use `make` to build this component. You will have to adjust the variables at the top of the [makefile](Makefile) to your environment (at least `OCMREPO`). By default all artifacts are built in the `gen` folder of this project.

The main targets are:

* `make ca`: builds a component archive
* `make ctf`: builds a common transport archive
* `make push`: stores the component version in an OCI registry
* `make transport`: transfers the component version between two registries
* `make descriptor`: displays the component descriptor
* `make describe`: displays the component and it dependencies in a tree structure
