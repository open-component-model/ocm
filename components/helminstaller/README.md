# Example helminstaller

This folder contains an example how to build an installer for the [toi](../../docs/reference/ocm_toi-bootstrapping.md) of the ocm CLI. It also demonstrates how to use OCM with building multiarch-images. It contains
an `executorspec.yaml` as interface to the toi installer.


# Building

You can use `make` to build this component. You will have to adjust the variables at the top of the [makefile](Makefile) to your environment (at least `OCMREPO`). By default all artifacts are built in the `gen` folder of this project.

The main targets are:

* `make ca`: builds a component archive
* `make ctf`: builds a common transport archive
* `make push`: stores the component version in an OCI registry
* `make transport`: transfers the component version between two registries
* `make descriptor`: displays the component descriptor
* `make describe`: displays the component and it dependencies in a tree structure


# Usage

The helm installer is used by providing a reference to this component version along with a `packagespec.yaml` description.

```yaml
apiVersion: ocm.software/v3alpha1
kind: ComponentVersion
metadata:
  ...
spec:
  references:
  - componentName: ocm.software/toi/installers/helminstaller
    name: installer
    version: v0.2.0
...
```

See the [helmdemo](../helmdemo/README.md) for an example including a [packagespec.yaml](../helmdemo/packagespec.yaml).

