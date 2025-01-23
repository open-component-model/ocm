package main

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/semverutils"
)

// --- begin resource interface ---
type Resource interface {
	GetIdentity() metav1.Identity
	GetType() string
	GetAccess() string
	GetData() ([]byte, error)

	SetError(s string)
	AddDataFromMethod(ctx ocm.ContextProvider, m ocm.AccessMethod) error

	Close() error
}

// --- end resource interface ---

// --- begin resource factory ---
// ResourceFactory is used to create a particular resource object.
type ResourceFactory interface {
	Create(id metav1.Identity, typ string) Resource
}

// --- end resource factory ---

// --- begin resource implementation ---
// resource is a Resource implementation using
// the original access method to cache the content.
type resource struct {
	Identity     metav1.Identity
	ArtifactType string
	Access       string
	Data         blobaccess.BlobAccess
}

var _ Resource = (*resource)(nil)

func (r *resource) AddDataFromMethod(ctx ocm.ContextProvider, m ocm.AccessMethod) error {
	// provide an own reference to the method
	// to store this in the provided resource object.
	priv, err := m.Dup()
	if err != nil {
		return err
	}

	// release a possible former cache entry
	if r.Data != nil {
		r.Data.Close()
	}
	r.Data = priv.AsBlobAccess()
	// release obsolete blob access
	r.Access = m.AccessSpec().Describe(ctx.OCMContext())
	return nil
}

// Close releases the cached access.
func (r *resource) Close() error {
	c := r.Data
	if c == nil {
		return nil
	}
	r.Data = nil
	return c.Close()
}

// --- end resource implementation ---

// --- begin caching factory ---
// CachingFactory provides resource inmplementations
// using the original access as cache.
type CachingFactory struct{}

func (c CachingFactory) Create(id metav1.Identity, typ string) Resource {
	return &resource{
		Identity:     id,
		ArtifactType: typ,
	}
}

// --- end caching factory ---

func (r *resource) GetIdentity() metav1.Identity {
	return r.Identity
}

func (r *resource) GetType() string {
	return r.ArtifactType
}

func (r *resource) GetAccess() string {
	return r.Access
}

func (r *resource) SetError(s string) {
	r.Access = "error: " + s
}

func (r *resource) GetData() ([]byte, error) {
	if r.Data == nil {
		return nil, fmt.Errorf("no data set")
	}
	return r.Data.Get()
}

func ResourceManagement() error {
	// get the default context providing
	// all OCM entry point registrations, like
	// access method, repository types, etc.
	// The context bundles all registrations and
	// configuration settings, like credentials,
	// which should be used when working with the OCM
	// ecosystem.
	ctx := ocm.DefaultContext()

	// --- begin decouple ---
	// gathering resources, this is completely hidden
	// behind an implementation.
	resources, err := GatherResources(ctx, CachingFactory{})
	if err != nil {
		return err
	}

	var list errors.ErrorList

	list.Add(HandleResources(resources))

	// we are done, so close the resources, again.
	for i, r := range resources {
		list.Addf(nil, r.Close(), "closing resource %d", i)
	}
	return list.Result()
	// --- end decouple ---
}

// --- begin handle ---
func HandleResources(resources []Resource) error {
	var list errors.ErrorList
	fmt.Printf("*** resources:\n")
	for i, r := range resources {
		fmt.Printf("  %2d: extra identity: %s\n", i+1, r.GetIdentity())
		fmt.Printf("      resource type:  %s\n", r.GetType())
		fmt.Printf("      access:         %s\n", r.GetAccess())
	}

	return list.Result()
}

// --- end handle ---

func GatherResources(ctx ocm.Context, factory ResourceFactory) ([]Resource, error) {
	var resources []Resource

	spec := ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")

	// And the context can now be used to map the descriptor
	// into a repository object, which then provides access
	// to the OCM elements stored in this repository.
	// --- begin repository ---
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot setup repository")
	}

	// to release potentially allocated temporary resources,
	// many objects must be closed, if they should not be used
	// anymore.
	// This is typically done by a `defer` statement placed after a
	// successful object retrieval.
	defer repo.Close()
	// --- end repository ---

	// Now, we look up the OCM CLI component.
	// All kinds of repositories, regardless of their type
	// feature the same interface to work with OCM content.
	// --- begin lookup component ---
	c, err := repo.LookupComponent("ocm.software/ocmcli")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot lookup component")
	}
	defer c.Close()
	// --- end lookup component ---

	// Now we look for the versions of the component
	// available in this repository.
	versions, err := c.ListVersions()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot query version names")
	}

	// OCM version names must follow the SemVer rules.
	// Therefore, we can simply order the versions and print them.
	err = semverutils.SortVersions(versions)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot sort versions")
	}
	versions = []string{"...", "0.9.0", "0.10.0", "0.11.0", "0.12.0", "0.12.1", "0.13.0", "0.14.0", "0.15.0", "0.17.0", "0.18.0", "..."}
	fmt.Printf("versions for component ocm.software/ocmcli: %s\n", strings.Join(versions, ", "))

	// Now, we have a look at the latest version. it is
	// the last one in the list.
	// --- begin lookup version ---
	// to retrieve the latest version use
	// cv, err := c.LookupVersion(versions[len(versions)-1])
	cv, err := c.LookupVersion("0.17.0")
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get latest version")
	}
	defer cv.Close()
	// --- end lookup version ---

	cd := cv.GetDescriptor()
	fmt.Printf("looking up resources of the latest version:\n")
	fmt.Printf("  version:  %s\n", cv.GetVersion())
	fmt.Printf("  provider: %s\n", cd.Provider.Name)

	// and list all the included resources.
	// Resources have some metadata, like the resource identity and a resource type.
	// And they describe how the content of the resource (as blob) can be accessed.
	// This is done by an *access specification*, again a serializable descriptor,
	// like the repository specification.
	// --- begin resources ---
	for _, r := range cv.GetResources() {
		res := factory.Create(
			r.Meta().GetIdentity(cv.GetDescriptor().Resources),
			r.Meta().GetType(),
		)
		acc, err := r.Access()
		if err != nil {
			res.SetError(err.Error())
		} else {
			m, err := acc.AccessMethod(cv)
			if err == nil {
				// delegate data handling to target
				// we don't know, how this is implemented.
				err = res.AddDataFromMethod(ctx, m)
				if err != nil {
					res.SetError(err.Error())
				}
				// release local usage of the access method object
				m.Close()
			} else {
				res.SetError(err.Error())
			}
		}
		resources = append(resources, res)
	}
	// --- end resources ---
	return resources, nil
}
