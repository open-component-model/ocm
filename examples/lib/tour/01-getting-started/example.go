package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/extraid"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/semverutils"
)

func GettingStarted() error {
	// get the default context providing
	// all OCM entry point registrations, like
	// access method, repository types, etc.
	// The context bundles all registrations and
	// configuration settings, like credentials,
	// which should be used when working with the OCM
	// ecosystem.
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// The context acts as the central entry
	// point to get access to OCM elements.
	// First, get a repository, to look for
	// component versions. We use the OCM
	// repository providing the standard OCM
	// components.

	// for every storage technology used to store
	// OCM components, there is a serializable
	// descriptor object, the repository specification.
	// It describes the information required to access
	// the repository and can be used to store the serialized
	// form as part of other resources, for example
	// Kubernetes resources.
	// The available repository implementations can be found
	// under .../api/ocm/extensions/repositories.
	// --- begin repository spec ---
	spec := ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")
	// --- end repository spec ---

	// And the context can now be used to map the descriptor
	// into a repository object, which then provides access
	// to the OCM elements stored in this repository.
	// --- begin repository ---
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot setup repository")
	}
	// --- end repository ---

	// to release potentially allocated temporary resources,
	// many objects must be closed, if they should not be used
	// anymore.
	// This is typically done by a `defer` statement placed after a
	// successful object retrieval.
	// --- begin close ---
	defer repo.Close()
	// --- end close ---

	// Now, we look up the OCM CLI component.
	// All kinds of repositories, regardless of their type
	// feature the same interface to work with OCM content.
	// --- begin lookup component ---
	c, err := repo.LookupComponent("ocm.software/ocmcli")
	if err != nil {
		return errors.Wrapf(err, "cannot lookup component")
	}
	defer c.Close()
	// --- end lookup component ---

	// Now we look for the versions of the component
	// available in this repository.
	// --- begin versions ---
	versions, err := c.ListVersions()
	if err != nil {
		return errors.Wrapf(err, "cannot query version names")
	}
	// --- end versions ---

	// OCM version names must follow the SemVer rules.
	// Therefore, we can simply order the versions and print them.
	// --- begin semver ---
	err = semverutils.SortVersions(versions)
	if err != nil {
		return errors.Wrapf(err, "cannot sort versions")
	}
	fmt.Printf("versions for component ocm.software/ocmcli: %s\n", strings.Join(versions, ", "))
	// --- end semver ---

	// Now, we have a look at the latest version. it is
	// the last one in the list.
	// --- begin lookup version ---
	// to retrieve the latest version use
	// cv, err := c.LookupVersion(versions[len(versions)-1])
	cv, err := c.LookupVersion("0.17.0")
	if err != nil {
		return errors.Wrapf(err, "cannot get latest version")
	}
	defer cv.Close()
	// --- end lookup version ---

	fmt.Printf("--- begin version ---\n")
	// The component version object provides access
	// to the component descriptor.
	// --- begin component descriptor ---
	cd := cv.GetDescriptor()
	fmt.Printf("resources of the latest version:\n")
	fmt.Printf("  version:  %s\n", cv.GetVersion())
	fmt.Printf("  provider: %s\n", cd.Provider.Name)
	// --- end component descriptor ---

	// and list all the included resources.
	// Resources have some metadata, like the resource identity and a resource type.
	// And they describe how the content of the resource (as blob) can be accessed.
	// This is done by an *access specification*, again a serializable descriptor,
	// like the repository specification.
	// --- begin resources ---
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
	// --- end resources ---
	fmt.Printf("--- end version ---\n")

	// Get the executable for the actual environment.
	// The identity of a resource described by a component version
	// consists of a set of properties. The property `name` is mandatory.
	// But there may be more identity attributes
	// finally stored as ``extraIdentity` in the component descriptor.
	// A convention is to use dedicated identity properties to indicate the
	// operating system and the architecture for executables.
	// --- begin find executable ---
	id := metav1.NewIdentity("ocmcli",
		extraid.ExecutableOperatingSystem, runtime.GOOS,
		extraid.ExecutableArchitecture, runtime.GOARCH,
	)

	res, err := cv.GetResource(id)
	if err != nil {
		return errors.Wrapf(err, "resource %s", id)
	}
	// --- end find executable ---

	// download to /tmp/ocmcli using basic model
	// operations.
	fmt.Printf("downloading OCM cli resource %s...\n", id)
	basic := true

	var reader io.ReadCloser
	if basic {
		// these are the basic model operations to get a reader
		// for the resource blob.
		// First, get the access method for the resource.
		// Second, request a reader for the blob.
		// --- begin getting access ---
		var m ocm.AccessMethod
		m, err = res.AccessMethod()
		if err != nil {
			return errors.Wrapf(err, "cannot get access method")
		}
		// --- end getting access ---

		// the method needs to be closed, because the method
		// object may cache the technical blob representation
		// generated by accessing the underlying access technology.
		// (for example, accessing an OCI image requires a sequence of
		// backend requests for the manifest, the layers, etc, which will
		// then be packaged into a tar archive returned as blob).
		// This caching may not be required, if the backend directly
		// returns a blob.
		// --- begin closing access ---
		defer m.Close()
		// --- end closing access ---

		// the method now also provides information about the returned
		// blob format in form of a mime type.
		// --- begin getting reader ---
		fmt.Printf("  found blob with mime type %s\n", m.MimeType())
		reader, err = m.Reader()
		// --- end getting reader ---
	} else {
		// because this is a common operation, there is a
		// utility function handling this code sequence.
		// --- begin utility function ---
		reader, err = utils.GetResourceReader(res)
		// --- end utility function ---
	}
	// --- begin closing reader ---
	if err != nil {
		return errors.Wrapf(err, "cannot get resource reader")
	}
	defer reader.Close()
	// --- end closing reader ---

	// --- begin copy ---
	file, err := os.OpenFile("/tmp/ocmcli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o766)
	if err != nil {
		return errors.Wrapf(err, "cannot open output file")
	}
	defer file.Close()

	n, err := io.Copy(file, reader)
	if err != nil {
		return errors.Wrapf(err, "write executable")
	}
	fmt.Printf("%d bytes written\n", n)
	// --- end copy ---

	// alternatively, a registered downloader for executables can be used.
	// download.DownloadResource is used to download resources with specific handlers for the
	// selected resource and mime type combinations.
	// The executable downloader is registered by default and automatically
	// sets the `X `flag for the written file.
	// --- begin download ---
	_, err = download.DownloadResource(ctx, res, "/tmp/ocmcli", download.WithPrinter(common.NewPrinter(os.Stdout)))
	if err != nil {
		return errors.Wrapf(err, "download failed")
	}
	// --- end download ---

	return nil
}
