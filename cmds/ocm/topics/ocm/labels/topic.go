package topicocmlabels

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/valuemergehandler"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "ocm-labels",
		Short: "Labels and Label Merging",

		Long: `
Labels are a set of arbitrary properties, which can be attached to elements
of a component version:
- a component version itself
- the provider of a component version
- resources
- sources
- component references

The dedicated elements support this by providing a field <code>labels</code>,
which is a list of label definitions. Every label definition has several fields:

- *<code>name</code>* *string* 

  The name of the label also determines the interpretation of its value. All labels
  with a dedicated name must have the same globally unique meaning, enabling a
  common understanding of label content for tools working of such properties of an
  element.

  There are several predefined labels, they just use flat names. To guarantee
  globally unique meanings of labels a label name may have a hierarchical
  structure. Names defined in dedicated definition realms must be prefixed by
  a DNS domain-like string identifying the organization of realm defining the
  label's value structure. For example: <code>acme.org/maturity/level</code>.

  Hereby, the name defines the meaning of the value and its value structure.
  To support the evolution of the value structure a label field optionally
  contains a <code>version</code> field, which finally defines the concrete
  value structure in the context of the meaning of the label name. A version
  is just a number prefixed with a <code>v</code>. If not specified, the
  version <code>v1</code> is assumed.

- *<code>version</code>* *string* (optional) (default: <code>v1</code>)

  The format version of the label value in the context of the label name.

- *<code>value</code>* *any*

  The value of the label according to the specified format version of the
  label in the context of its name.

- *<code>signing</code>* *bool* (optional)

  By default, labels are not signature-relevant and they will nor influence the
  digest of the component version. This allows adding, deleting or modifying
  labels as part of a process chain during the lifecycle of a component version.

  Labels which should describe relevant and unmodifiable content can be marked
  to be signing relevant by setting this label field to <code>true</code>.

- *<code>merge</code>* *merge spec* (optional)
  
  Modifiable labels can be changed independently in any transport target
  location of a component version. This might require to update label values
  when importing a new setting for a component version. This means a merging
  of content to reflect the combination of changes in the transport source and
  target.

  This is supported by the possibility to specify merge algorithms.
  The can be bound to a dedicated label incarnation or to the label name.

### Merge Specification

A merge specification consists of two fields:

- *<code>algorithm</code>* *string* (optional) (default: <code>default</code>)

  The name of the algorithm to be used for the merge process.

- *<code>config</code>* *any* (optional)

  An algorithm specific configuration to control the merge process.

There is an often used configuration field <code>overwrite</code> with a common
meaning for all algorithms supporting it. It controls the conflict resolution
and has the following values:

- *<code>none</code>*: conflicting values prevent the merging. An update
  transfer process will be aborted.

- *<code>local</code>*: a conflict will be resolved to the local change
  (in the target environment)

- *<code>inbound</code>*: a conflict will be resolved to the value provided
  by the source environment

- &lt;empty>: use a default provided by the dedicated algorithm.

The default behaviour might mean to apply a cascaded merge specification, if
the merge specification supports to specify appropriate fields to specify
this specification (for example a field <code>entries</code>).

### Determining a Merge Specification

A merge specification directly attached to a label is always preferred.
If no algorithm is specified a merge assignment for the label name and
its version is evaluated. The assignment hint is composed with

<center>
 <code>label:</code>&lt;*label name*><code>@</code>%lt;version>
</center>

The label version is defaulted to <code>v1</code>.

### Supported Merge Algorithms

There are some built-in algorithms featuring a flat name. But it will be
possible to add arbitrary algorithms using the plugin concept. 

The following algorithms are possible:
` + valuemergehandler.Usage(ctx.OCMContext()),
	}
}
