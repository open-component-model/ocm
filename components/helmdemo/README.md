# Example echoserver

This folder contains an example how to use the ocm CLI to build and upload a component version to an OCI registry. It refers to a public image and creates a component-descriptor with an appropriate access method. Additionally it embeds a helm chart for installation. Finally it adds a [toi package](../../docs/reference/ocm_toi-bootstrapping.md) specification and a reference to the [helminstaller](../helminstaller/README.md). This allows installing the component with the toi installer.

# Building

You can use `make` to build this component. You will have to adjust the variables at the top of the [makefile](Makefile) to your environment (at least `OCMREPO`). By default, all artifacts are built in the `gen` folder of this project.

The main targets are:

* `make ca`: builds a component archive
* `make ctf`: builds a common transport archive
* `make push`: stores the component version in an OCI registry
* `make transport`: transfers the component version between two registries
* `make descriptor`: displays the component descriptor
* `make describe`: displays the component and it dependencies in a tree structure

You can find more information [here](../../cmds/helminstaller/README.md).

# Using transported images (image mapping)

When transferring a component version between registries or from a local common transport archive to an OCI registry often images are included. When installing the applications all images in a deployment need to be adjusted so that their url points to the new location. The toi installer can perform this mapping and will dynamically adjust the helm values. This needs to be configured in the `packagespec.yaml` file. The `packagespec.yaml` inserts the content of another file `helmconfig.yaml` where the image mapping is defined.

```yaml
imageMapping:
  - tag: image.tag
    repository: image.repository
    resource:
      name: image
```

This image mapping instructs the helm-installer to replace the tags

```
image:
  repository: <add reference to image from component version here>
  tag:        <add tag from component version here>
```

with the value it finds in the resource named "image" of the current component version:

```yaml
---
name: image
type: ociImage
version: "1.0"
access:
  type: ociArtifact
  imageReference: gcr.io/google_containers/echoserver:1.10
```

## Notes

For a successful image mapping ensure that all you image resources have a globalAccess and relation external in their component descriptor. Only with global access Kubernetes is able to pull an image. Ensure that your components are transferred with the --copy-resources flag.

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
