package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

const (
	COMP = "acme.org/mytestcomponent"
	VERS = "1.0.0"
)

func CTFExample() (rerr error) {
	var finalize finalizer.Finalizer

	// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

	defer finalize.FinalizeWithErrorPropagation(&rerr)

	octx := ocm.DefaultContext()

	memfs := memoryfs.New()

	repo, err := ctf.Open(octx, accessobj.ACC_CREATE, "test", 0o700, accessio.PathFileSystem(memfs))
	if err != nil {
		return err
	}
	finalize.Close(repo)

	for i := 1; i <= 1; i++ {
		loop := finalize.Nested()
		cname := fmt.Sprintf("%s%d", COMP, i)
		comp, err := repo.LookupComponent(cname)
		if err != nil {
			return errors.Wrapf(err, "cannot lookup component %s", cname)
		}
		loop.Close(comp)

		compvers, err := comp.NewVersion(VERS, false)
		if err != nil {
			return errors.Wrapf(err, "cannot create new version %s", VERS)
		}
		loop.Close(compvers)

		// add provider information
		compvers.GetDescriptor().Provider = metav1.Provider{Name: "acme.org"}

		// add a new resource artifact with the local identity `name="test"`.
		err = compvers.SetResourceBlob(
			&compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: "test",
				},
				Type:     resourcetypes.BLOB,
				Relation: metav1.LocalRelation,
			},
			blobaccess.ForString(mime.MIME_TEXT, "testdata"),
			"", nil,
		)
		if err != nil {
			return errors.Wrapf(err, "cannot add resource")
		}

		if err = comp.AddVersion(compvers); err != nil {
			return errors.Wrapf(err, "cannot add new version")
		}
		fmt.Printf("added component %s version %s\n", cname, VERS)
		if err := loop.Finalize(); err != nil {
			return err
		}
	}
	return nil
}
