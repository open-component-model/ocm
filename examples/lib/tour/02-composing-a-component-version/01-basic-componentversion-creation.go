package main

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/elements"
	"ocm.software/ocm/api/ocm/elements/artifactblob/dockermultiblob"
	"ocm.software/ocm/api/ocm/elements/artifactblob/textblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/semverutils"
)

// setupVersion configures a component version.
// This can be called on any component version, regardless of
// its origin.
func setupVersion(cv ocm.ComponentVersionAccess) error {
	// The provider is a structure with a name and some labels.
	// We just set the name here by directly setting the `Name` attribute.
	// --- begin setup provider ---
	provider := &compdesc.Provider{
		Name: "acme.org",
	}
	fmt.Printf("  setting provider...\n")
	err := cv.SetProvider(provider)
	if err != nil {
		return errors.Wrapf(err, "cannot set provider")
	}
	// --- end setup provider ---

	////////////////////////////////////////////////////////////////////////////
	// Ok, a component version not describing any resources
	// is pretty useless.
	// So, lets add some resources, now.

	fmt.Printf("  setting external resource 'image'...\n")
	// first, we add some resource already located in
	// some external registry.
	// A resources has some metadata, like an identity
	// and a type.
	// The identity is just a set of string properties,
	// at least containing the `name` property.
	// additional identity properties can be added via
	// options.
	// The type represents the logical meaning of the
	// resource, here an `ociImage`.
	// --- begin setup resource meta ---
	meta, err := elements.ResourceMeta("image", resourcetypes.OCI_IMAGE)
	if err != nil {
		// without metadata options, there will be never be an error,
		// bit to be complete, we just handle the error case, here.
		return errors.Wrapf(err, "invalid resource meta")
	}
	// --- end setup resource meta ---

	// And most importantly, it requires content.
	// Content can be already present in some external
	// repository. As long, as there is an access type
	// for this kind of repository, we can just refer to it.
	// Here, we just use an image provided by the
	// OCM ecosystem.
	// Supported access types can be found under
	// .../api/ocm/extensions/accessmethods.
	// --- begin setup image access ---
	acc := ociartifact.New("ghcr.io/open-component-model/ocm/ocm.software/toi/installers/helminstaller/helminstaller:0.4.0")
	// --- end setup image access ---

	// Once we have both, the metadata and the content specification,
	// we can now add the resource to our component version.
	// The SetResource methods will replace an existing resource with the same
	// identity, or add the resource, if no such resource exists in the component
	// version.
	// --- begin setup resource ---
	err = cv.SetResource(meta, acc)
	if err != nil {
		return errors.Wrapf(err, "cannot add access to ocmcli-image)")
	}
	// --- end setup resource ---

	////////////////////////////////////////////////////////////////////////////
	// Now, we will add a second resource, some unspecific yaml data.
	// Therefore, we use the generic YAML resource type.
	// In practice, you should always use a resource type describing
	// the real meaning of the content, for example something like
	// `kubernetesManifest`. This enables tools working with specific content
	// to understand the resource set of a component version.

	fmt.Printf("  setting blob resource 'descriptor'...\n")
	// --- begin setup second meta ---
	meta, err = elements.ResourceMeta("descriptor", resourcetypes.OCM_YAML)
	if err != nil {
		return errors.Wrapf(err, "invalid resource meta")
	}
	// --- end setup second meta ---

	basic := true
	yamldata := `
type: mySpecialDocument
data: some very important data required to understand this component
`

	if basic {
		// Besides referring to external resources, another possibility
		// to add content is to directly provide the content blob. The
		// used abstraction here is blobaccess.BlobAccess.
		//
		// Any blob content, which can be provided by an implementation of this
		// interface, can be added as resource to a component version.
		// The library provides various access implementations for blobs
		// taken from the local host or from other repositories.
		// For example, this could be some file system content.
		// To describe blobs taken from external repositories
		// an access type specification can be mapped to a blob access.
		// Hereby, blobs are stored along with the component descriptor
		// instead of storing a reference to content in an external repository.
		//
		// The most simple form is to directly provide a byte sequence,
		// for example some YAML data.
		// A blob always must provide a mime type, describing the
		// technical format of the blob's byte sequence.
		// This is different
		// from the resource type. A logical resource, like a *Helm chart* can be
		// represented
		// in different technical formats, for example a Helm chart archive
		// or as OCI image archive. While the type described the
		// logical content, the meaning of the resource, its mime type
		// described the technical blob format used to represent
		// the resource as byte sequence.
		// --- begin string blob access ---
		blob := blobaccess.ForString(mime.MIME_YAML, yamldata)
		// --- end string blob access ---

		// when storing the blob, it is possible to provide some
		// optional additional information:
		// - a name of the resource described by the blob, which could
		//   be used to do a later upload into an external repository
		//   (for example the image repository of an OCI image stored
		//   as local blob)
		// - an additional access type, which provides an alternative
		//   global technology specific access to the same content
		//   (we don't use it, here).
		// --- begin setup by blob access ---
		err = cv.SetResourceBlob(meta, blob, "", nil)
		if err != nil {
			return errors.Wrapf(err, "cannot add yaml document")
		}
		// --- end setup by blob access ---

		// Resources added by blobs will be stored along with the component
		// version metadata in the same repository, no external
		// repository is required.
	} else {
		// The above blob example describes the basic operations,
		// which can be used to compose any kind of resource
		// from any kind of source.
		// For selected use cases there are convenience helpers available,
		// which can be used to compose a resource access object.
		// This is basically the same interface returned by GetResource
		// functions on the component version from the last example.
		// Such objects can directly be used to add/modify a resource in a
		// component version.
		//
		// The above case could also be written as follows:
		// --- begin setup by access ---
		res := textblob.ResourceAccess(cv.GetContext(), meta, yamldata,
			textblob.WithMimeType(mime.MIME_YAML))
		err = cv.SetResourceByAccess(res)
		if err != nil {
			return errors.Wrapf(err, "cannot add yaml document")
		}
		// --- end setup by access ---

		// The resource access is an abstraction of external access via access
		// methods or direct blob access objects and additionally
		// contain all the required resource metadata.
	}

	// There are even more complex blob sources, for example
	// for Helm charts stored in the file system, or even for images
	// generated by docker builds.
	// Here, we just compose a multi-platform image built with `buildx`
	// from these sources (components/ocmcli) featuring two flavors.
	// (you have to execute `make image.multi` in components/ocmcli
	// before executing this example.
	fmt.Printf("  setting blob resource 'ocmcli'...\n")
	// --- begin setup by docker ---
	meta, err = elements.ResourceMeta("ocmcli", resourcetypes.OCI_IMAGE)
	if err != nil {
		return errors.Wrapf(err, "invalid resource meta")
	}
	res := dockermultiblob.ResourceAccess(cv.GetContext(), meta,
		dockermultiblob.WithPrinter(common.StdoutPrinter),
		dockermultiblob.WithHint("ocm.software/ocmci"),
		dockermultiblob.WithVersion(current_version),
		dockermultiblob.WithVariants(
			fmt.Sprintf("ocmcli-image:%s-linux-amd64", current_version),
			fmt.Sprintf("ocmcli-image:%s-linux-arm64", current_version),
		),
	)
	err = cv.SetResourceByAccess(res)
	if err != nil {
		return errors.Wrapf(err, "cannot add ocmcli")
	}
	// --- end setup by docker ---
	return err
}

// addVersion configures and adds a new version to the
// given repository.
// This can be called for any repository object, regardless of its
// underlying storage backend.
func addVersion(repo ocm.Repository, name, version string) error {
	// basically, any other OCM repository implementation could
	// be used here.

	// now we compose a new component version, first we create
	// a new version backed by this repository.
	// The result is a memory based representation, which is not yet persisted.
	// --- begin new version ---
	cv, err := repo.NewComponentVersion(name, version)
	if err != nil {
		return errors.Wrapf(err, "cannot create new version")
	}
	defer cv.Close()
	// --- end new version ---

	err = setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "cannot setup new version")
	}

	// finally, we add the new version to the repository.
	fmt.Printf("adding component version\n")
	err = repo.AddComponentVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "cannot save version")
	}
	return nil
}

func describeVersion(cv ocm.ComponentVersionAccess) error {
	// many elements of the API keep trak of their context
	ctx := cv.GetContext()

	// Have a look at the component descriptor
	cd := cv.GetDescriptor()
	fmt.Printf("resources of the latest version of %s:\n", cv.GetName())
	fmt.Printf("  version:  %s\n", cv.GetVersion())
	fmt.Printf("  provider: %s\n", cd.Provider.Name)

	// and list all the included resources.
	for i, r := range cv.GetResources() {
		fmt.Printf("  %2d: name:           %s\n", i+1, r.Meta().GetName())
		fmt.Printf("      extra identity: %s\n", r.Meta().GetExtraIdentity())
		fmt.Printf("      resource type:  %s\n", r.Meta().GetType())
		acc, err := r.Access()
		if err != nil {
			fmt.Printf("      access:         error: %s\n", err)
		} else {
			fmt.Printf("      access:         %s\n", acc.Describe(ctx))
		}
	}
	return nil
}

func listVersions(repo ocm.Repository, list ...string) error {
	// now we just list the found components and their latest version.

	// because we will loop over versions, we want to
	// close the objects acquired in the loop at the end of
	// the loop, but on error returns, also.
	// This can be achieved using a Finalizer object.
	var finalize finalizer.Finalizer
	defer finalize.Finalize()

	for _, name := range list {
		// one instance per loop step, which can separately finalized.
		// It is still bound to the top level finalizer and
		// will be finalized on return statements.
		nested := finalize.Nested()

		// Do you remember the code from example 1?
		// This is basically the same, used to examine the
		// created version

		c, err := repo.LookupComponent(name)
		if err != nil {
			return errors.Wrapf(err, "cannot get component %s", name)
		}
		nested.Close(c, "component %s", name)

		// Now we look for the versions of the component
		// available in this repository.
		versions, err := c.ListVersions()
		if err != nil {
			return errors.Wrapf(err, "cannot query version names for %s", name)
		}

		// OCM version names must follow the semver rules.
		err = semverutils.SortVersions(versions)
		if err != nil {
			return errors.Wrapf(err, "cannot sort versions for %s", name)
		}
		fmt.Printf("versions for component %s: %s\n", name, strings.Join(versions, ", "))

		cv, err := repo.LookupComponentVersion(name, versions[len(versions)-1])
		if err != nil {
			return errors.Wrapf(err, "cannot get latest version for %s", name)
		}
		nested.Close(cv, "component version", common.VersionedElementKey(cv).String())

		nested.Finalize()
	}
	return nil
}

func ComposingAComponentVersionA() error {
	// yes, we need an OCM context, again
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// To compose and store a new component version
	// we need some OCM repository to
	// store the component. The most simple
	// external repository could be the file system.
	// For this purpose OCM defines a distribution format, the
	// Common Transport Format (CTF),
	// which is an extension of the OCI distribution
	// specification.
	// There are three flavors, Directory, Tar or TGZ.
	// The implementation provides a regular OCM repository
	// interface, like the one used in the previous example.
	// --- begin create ctf ---
	repo, err := ctfocm.Open(ctx, ctfocm.ACC_WRITABLE|ctfocm.ACC_CREATE, "/tmp/example02.ctf", 0o0744, ctfocm.FormatDirectory)
	if err != nil {
		return errors.Wrapf(err, "cannot create transport repository")
	}
	defer repo.Close()
	// --- end create ctf ---

	// now we create a first component version in this repository.
	err = addVersion(repo, "acme.org/example02", "v0.1.0")
	if err != nil {
		return err
	}

	// list the versions as known from example 1
	return listVersions(repo)
}
