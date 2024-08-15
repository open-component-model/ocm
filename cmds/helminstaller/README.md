# Helm Installer for the TOI Framework

The helm installer provides an executor plugin for the [TOI framework](../../docs/reference/ocm_toi-bootstrapping.md).
It can be used to install a top level helm chart, which might optionally
by composed of a set of sub charts.

## Component

The helm installer plugin for the TOI framework if provided
by component `ocm.software/toi/installers/helminstaller`.

It provides a single `toiExecutor` resource with the
name `toiexecutor`.

## Configuration

The executor configuration supports the following field:

- `chart` *[ResourceReference](../../docs/reference/ocm_toi-bootstrapping.md#resourcereference)* (**required**) The resource containing the top level helmchart to work on.

- `subCharts` *map[string]ResourceReference* (**optional**) A set of resources describing charts used as sub charts for the top level chart. The map key is the name of the folder created for the sub chart.

- `release` *string* (**optional**) The default release name to use for installation.

- `namespace` *string* (**optional**) The default namespace to use for the installation.

- `createNamespace` *boolean* (**optional**) If set to true the namespace will be created.

- `imageMapping` *[]ImageMapping* list of localization rules for images.

- `values` **yaml* The default values used for installation. They will be overwritten by the given installation values (at top-level).

- `kubeConfigName` *string* (**optional** default: `target`) The credential name to lookup for the kubeconfig used to access the target Kubernetes cluster.

### Image Mappings

Image mappings describe the injection of OCI image locations, names and versions
taken from the component version info dedicated properties of the installation
values. The helmchart MUST take all image locations used in the templates
from dedicated values. This is required to support the transport of
component versions into local repositories environments, which should be used
on the installation site.

An image mapping consists of resource reference fields to refer to an OCI image resource used to extract the image location plus the following additional fields:

- `tag` *string*  (**optional**) The property of the values used to inject the tag/version of the image.

- `repository` *string*  (**optional**) The property of the values used to inject the base URL (location+repository) of the image.

- `image` *string*  (**optional**) The property of the values used to inject the complete image name.

At least the `image` attribute or the `tag` and `repositories` attributes must be used to provide a complete image location.

### Configuring Subcharts

Subcharts are configured as usual with the values for the parent chart (see <https://helm.sh/docs/chart_template_guide/subcharts_and_globals/>).

The key of the subchart is used as top-level values key to add settings for the subchart.
Similar to the parent chart, images used by subcharts must be localized via [image mappings](#image-mappings), also. The subchart values must accept tag, repository and/or image value
fields for used images. They are set by concatenating the key of the subchart with the name of the value field.
