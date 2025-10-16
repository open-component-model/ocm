package main

import (
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/elements"
	"ocm.software/ocm/api/ocm/elements/artifactblob/dockermultiblob"
	"ocm.software/ocm/api/ocm/elements/artifactblob/textblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/semverutils"
	"ocm.software/ocm/examples/lib/helper"
)

// setupVersion configures a component version.
// This can be called on any omponent version, regardless of
// its origin.
func setupVersion(cv ocm.ComponentVersionAccess) error {
	provider := &compdesc.Provider{
		Name: "acme.org",
	}
	fmt.Printf("  setting provider...\n")
	err := cv.SetProvider(provider)
	if err != nil {
		return errors.Wrapf(err, "cannot set provider")
	}

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
	// with at least containing the name property.
	// additional identity properties can be added via
	// options.
	meta, err := elements.ResourceMeta("image", resourcetypes.OCI_IMAGE)
	if err != nil {
		// without metadata options, there will be never be an error,
		// bit to be complete, we just handle the error case, here.
		return errors.Wrapf(err, "invalid resource meta")
	}

	// And most important it requires content.
	// Content can be already present in some external
	// repository. As long, as there is an access type
	// for this kind of repository, we can just refer to it.
	// Here, we just use an image provided by the
	// OCM ecosystem.
	// Supported access types can be found under
	// .../api/ocm/extensions/accessmethods.
	acc := ociartifact.New("ghcr.io/open-component-model/ocm/ocm.software/toi/installers/helminstaller/helminstaller:0.4.0")

	// Once we have both, the metadata and the content specification,
	// we can now add the resource.
	// The SetResource methods will replace an existing resource with the same
	// identity, or add the resource, if no such resource exists in the component
	// version.
	err = cv.SetResource(meta, acc)
	if err != nil {
		return errors.Wrapf(err, "cannot add access to ocmcli-image)")
	}

	////////////////////////////////////////////////////////////////////////////
	// Now, we will add a second resource, some unspecific yaml data.
	// Therefore, we use the generic YAML resource type.
	// In practice, you should always use a resource type describing
	// the real meaning of the content, for example something like
	// `kubernetesManifest`, This enables tools working with specific content
	// to understand the resource set of a component version.

	fmt.Printf("  setting blob resource 'descriptor'...\n")
	meta, err = elements.ResourceMeta("descriptor", resourcetypes.OCM_YAML)
	if err != nil {
		return errors.Wrapf(err, "invalid resource meta")
	}

	basic := true
	yamldata := `
type: mySpecialDocument
data: some very important data required to understand this component
`

	if basic {
		// Besides referring to external resources, another possibility
		// to add content is to directly provide the content blob. The
		// used abstraction here is blobaccess.BlobAccess.
		// Any blob content provided by an implementation of this
		// interface can be added as resource.
		// There are various access implementations for blobs
		// taken from the local host, for example, from the file system,
		// or from other repositories (for example by mapping
		// an access type specification into a blob access).
		// The most simple form is to directly provide a byte sequence,
		// for example some YAML data.
		// A blob always must provide a mime type, describing the
		// technical format of the blob's byte sequence.
		blob := blobaccess.ForString(mime.MIME_YAML, yamldata)

		// when storing the blob, it is possible to provide some
		// optional additional information:
		// - a name of the resource described by the blob, which could
		//   be used to do a later upload into an external repository
		//   (for example the image repository of an OCI image stored
		//   as local blob)
		// - an additional access type, which provides an alternative
		//   global technology specific access to the same content.
		// we don't use it, here.
		err = cv.SetResourceBlob(meta, blob, "", nil)
		if err != nil {
			return errors.Wrapf(err, "cannot add yaml document")
		}
	} else {
		// The above blob example describes the basic operations,
		// which can be used to compose any kind of resource
		// from any kind of source.
		// For selected use cases there are convenience helpers,
		// which can be used to compose a resource access object.
		// This is basically the same interface returned by GetResource
		// functions on the component version from the last example.
		// Such objects can directly be used to add/modify a resource in a
		// component version.
		// The above case could be written as follows, also:
		res := textblob.ResourceAccess(cv.GetContext(), meta, yamldata,
			textblob.WithMimeType(mime.MIME_YAML))
		err = cv.SetResourceByAccess(res)
		if err != nil {
			return errors.Wrapf(err, "cannot add yaml document")
		}
	}

	// There are even more complex blob sources, for example
	// for Helm charts stored in the file system, or even for images
	// generated by docker builds.
	// Here, we just compose a multi-platform image built with buildx
	// from these sources (components/ocmcli) featuring two flavors.
	// (you have to execute `make image.multi` in components/ocmcli
	// before executing this example.
	fmt.Printf("  setting blob resource 'ocmcli'...\n")
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
	cv, err := repo.NewComponentVersion(name, version)
	if err != nil {
		return errors.Wrapf(err, "cannot create new version")
	}
	defer cv.Close()

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
		displayDigest(r.Meta().Digest)
	}
	if len(cd.Signatures) > 0 {
		fmt.Printf("signatures:\n")
		for i, s := range cd.Signatures {
			fmt.Printf("  %2d: name:           %s\n", i+1, s.Name)
			displayDigest(&s.Digest)
			fmt.Printf("      signature:\n")
			fmt.Printf("        algorithm: %s\n", s.Signature.Algorithm)
			fmt.Printf("        mediaType: %s\n", s.Signature.MediaType)
			fmt.Printf("        value:     %s\n", s.Signature.Value)
		}
	}
	return nil
}

func displayDigest(d *metav1.DigestSpec) {
	if d != nil {
		fmt.Printf("      digest:\n")
		fmt.Printf("        algorithm:     %s\n", d.HashAlgorithm)
		fmt.Printf("        normalization: %s\n", d.NormalisationAlgorithm)
		fmt.Printf("        value:         %s\n", d.Value)
	}
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

func ReadConfiguration(ctx ocm.Context, cfg *helper.Config) error {
	if cfg.OCMConfig != "" {
		fmt.Printf("*** applying config from %s\n", cfg.OCMConfig)

		_, err := utils.Configure(ctx, cfg.OCMConfig)
		if err != nil {
			return errors.Wrapf(err, "error in ocm config %s", cfg.OCMConfig)
		}
	}
	return nil
}

func saveKey(key signutils.GenericPrivateKey) {
	block := signutils.PemBlockForPrivateKey(key)
	os.WriteFile("key.pem", pem.EncodeToMemory(block), 0o0600)
}

func lookupKey() signutils.GenericPrivateKey {
	data, err := os.ReadFile("key.pem")
	if err != nil {
		return nil
	}
	key, _ := signutils.GetPrivateKey(data)
	if err != nil {
		return nil
	}
	return key
}
