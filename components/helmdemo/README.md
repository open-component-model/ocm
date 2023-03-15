# Example echoserver

This folder contains an example how to use the ocm CLI to build and upload a component version to an OCI registry. It refers to a public image and creates a component-descriptor with an appropriate access method. Additionally it embeds a helm chart for installation. Finally it adds a [toi package](../../docs/reference/ocm_toi-bootstrapping.md) specification and a reference to the [helminstaller](../helminstaller/README.md). This allows installing the component with the toi installer.

# Building

You can use `make` to build this component. You will have to adjust the variables at the top of the [makefile](Makefile) to your environment (at least `OCMREPO`). By default all artifacts are built in the `gen` folder of this project.

The main targets are:

* `make ca`: builds a component archive
* `make ctf`: builds a common transport archive
* `make push`: stores the component version in an OCI registry
* `make transport`: transfers the component version between two registries
* `make descriptor`: displays the component descriptor
* `make describe`: displays the component and it dependencies in a tree structure

You can find more information [here](../../cmds/helminstaller/README.md).

# Installation

You can use the `ocm bootstrap` command to install this component with the [toi installer](../../docs/reference/ocm_toi-bootstrapping.md) on a Kubernetes cluster.

You will need a credentials file containing the kubeconfig for the target cluster:

`credentials.yaml`:

```yaml
credentials:
  target:
    credentials:
      KUBECONFIG: (( read("/<path-to-kuebconfig.yaml>", "text") ))

```

To set parameters for the target installation a configuration file can be created:

`params.yaml`:

```yaml
# Namespace in target cluster to install the component:
namespace: echo

# Name of the helm release to be created:
release: myechoserver

# User provided helm values:
replicaCount: 1
ingress:
  enabled: true
```

You can then install the echoserver with the command (`OCMREPO` and `VERSION` need to be adjusted):

```shell
OCMREPO=ghcr.io/open-component-model
VERSION=0.2.0
ocm bootstrap component install -p params.yaml -c credentials.yaml ${OCMREPO}//ocm.software/toi/demo/helmdemo:${VERSION}
```
