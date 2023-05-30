# Example: subcharts

This folder contains an example how to deploy an application consistsing of multiple helm charts. It consists of a helm chart, a [toi package](../../docs/reference/ocm_toi-bootstrapping.md) and a reference to the [helm helminstaller](../helminstaller/README.md).

The helm chart has multiple other charts as dependencies and has no templates. Its only purpose is to act as an anchor for deployment. The toi package is used to install the component with the helm-installer using the `ocm bootstrap` command.

## Helm Charts with dependencies
Goal of this deployment is to install two different components having their own helm-charts: echoserver and pod-info. Both have their own component version and a helm-chart resource. The chart dependencies are declared in the `packagespec.yaml` as `subcharts`:

```yaml
executors:
  - ...
    config:
      chart:
        resource:
          name: subchartsapp-chart
      subcharts:
        podinfo:
          resource:
            name: podinfo-chart
          referencePath:
          - name: podinfo
        echoserver:
          resource:
            name: echo-chart
          referencePath:
          - name: echoserver
      ...
```

For a full configuration see: [packagespec.yaml](packagespec.yaml)

For echoserver a dependency is declared to a resource named `echo-chart` contained in the reference named `echoserver` of this component version (`referencePath`):

```yaml
components:
- name: ocm.software/toi/demo/subcharts/subcharts
  version: ...
  ...
  componentReferences:
  - ...
  - name: echoserver
    componentName:  ...
    version: ...
  - ...
```

For a full configuration see: [packagespec.yaml](packagespec.yaml)

In the echoserver component version the resource `echo-chart` will be used:

```yaml
  component:
    name: ocm.software/toi/demo/subcharts/echoserver
    version: "1.10"
    ...
    resources:
    - ...
    - name: echo-chart
      version: "1.10"
      relation: local
      type: helmChart
      access:
        globalAccess:
          imageReference: <myrepo>/ocm.software/toi/demo/subcharts/echoserver/echoserver:0.1.0
          type: ociArtifact
        localReference: sha256:105804915dd4ab173a15d952721b05b87ce4f139358a05e3b6dc884a31d98502
        mediaType: application/vnd.oci.image.manifest.v1+tar+gzip
        referenceName: ocm.software/toi/demo/subcharts/echoserver/echoserver:0.1.0
        type: localBlob
```

Same mechanism for podinfo:

```yaml
components:
- name: ocm.software/toi/demo/subcharts/subcharts
  version: ...
  ...
  componentReferences:
  - ....
  - name: podinfo
    componentName: .../podinfo
    version: ...
  - ...

```
The `Chart.yaml` file used for the helm deployment is dynamically modified to add the `dependencies`  according to the helm documentation. There is no need to declare them manually (but it does not harm to do so). See [helm documentation]([https:/](https://helm.sh/docs/helm/helm_dependency/)) for details.


## Using transported images (image mapping)

When transferring a component version between registries or from a local common transport archive to an OCI registry often images are included. When installing the applications all images in a deployment need to be adjusted so that their url points to the new location. The toi installer can perform this mapping and will dynamically adjust the helm values. This needs to be configured in the `packagespec.yaml` file:

```yaml
executors:
  - ...
    config:
      ...
      imageMapping:
      - tag: echoserver.image.tag
        repository: echoserver.image.repository
        resource:
          name: echo-image
        referencePath:
        - name: echoserver
      - tag: podinfo.image.tag
        repository: podinfo.image.repository
        resource:
          name: podinfo-image
        referencePath:
        - name: podinfo
```
This configuration will set the tags in the generated helm values:

```yaml
echoserver:
  image: <add reference to image from component version here>
  tag:   <add tag from component version here>
podinfo:
  image: <add reference to image from component version here>>
  tag:   <add tag from component version here>
```

The first image mapping instructs the helm-installer to replace the image tag with the value it finds in the component reference named `echoserver` in the resource named "echo-image".

```
  component:
    ...
    componentReferences:
    ---
    - componentName: ocm.software/toi/demo/subcharts/echoserver
      name: echoserver
      version: "1.10"
```
and then in echoserver:

```yaml
component:
    name: ocm.software/toi/demo/subcharts/echoserver
    version: "1.10"
    ...
    resources:
    - access:
        globalAccess:
          imageReference: <OCMREPO>/echoserver:1.10
          type: ociArtifact
        localReference: sha256:4b93359cc643b5d8575d4f96c2d107b4512675dcfee1fa035d0c44a00b9c027c
        mediaType: application/vnd.docker.distribution.manifest.v2+tar+gzip
        referenceName: google-containers/echoserver:1.10
        type: localBlob
      name: echo-image
      relation: external
      type: ociImage
      version: "1.10"
```

The same lookup mechanism is used for the reference named `podinfo` and the resource `podinfo-image`.

### Testing

To test that the mapping works as expected you should check the pods after the deployment to contain the correct image URL.

```shell
kubectl get pods
NAME                                  READY   STATUS    RESTARTS   AGE
myrelease-echoserver-6d6b8c4664-ghzpp   1/1     Running   0          3h8m
myrelease-podinfo-666bcbddd4-xfdrr      1/1     Running   0          3h8m

kubectl describe pod myrelease-echoserver-6d6b8c4664-ghzpp

Name:         myrelease-echoserver-6d6b8c4664-ghzpp
Namespace:    subcharts
...
Image:          <myrepo>/google-containers/echoserver:1.10
```

Check that the image is taken from your target OCI registry.

### Notes

For a successful image mapping ensure that all you image resources have a `globalAccess` and relation `external` in their component descriptor. Only with global access Kubernetes is able to pull an image. Ensure that your components are transferred with the `--copy-resources` flag.

## Providing helm values for subcharts:

Usually deployments will get a configuration specific for this target enviroment. In helm this is done with values specific for a helm release.
When using the toi installer this works in the same way by providing the values with other parameters like name of the helm release and the target namespace in a <parameters.yaml> file:

```yaml
namespace: subcharts
release: myrelease
# User provided helm values (have to provided at top-level here)
podinfo:
  serviceAccount:
    enabled: True
    imagePullSecrets:
    - name: myoci-secret
echoserver:
  imagePullSecrets:
  - name: myoci-secret
```
Values for subcharts are located under a parent tag identifying the sub-chart. See the [Helm Documentation](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/) for details.


# Building

You can use `make` to build this component. You will have to adjust the variables at the top of the [makefile](Makefile) to your environment (at least `OCMREPO`). By default, all artifacts are built in the `gen` folder of this project.

The main targets are:

* `make ctf`: builds a common transport archive
* `make push`: stores the component version in an OCI registry
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
namespace: subcharts

# Name of the helm release to be created:
release: mysubcharts

# User provided helm values (have to provided at top-level here)
podinfo:
  serviceAccount:
    enabled: True
    imagePullSecrets:
    - name: gcr-secret
echoserver:
  imagePullSecrets:
  - name: gcr-secret
```

You can then install the echoserver with the command (`OCMREPO` and `VERSION` need to be adjusted):

```shell
OCMREPO=ghcr.io/open-component-model
VERSION=v0.3.0
ocm bootstrap component install -p params.yaml -c credentials.yaml $OCMREPO}//ocm.software/toi/demo/subcharts/subcharts:${VERSION}
```

