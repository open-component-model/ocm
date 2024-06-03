package helm

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/helm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	ocihelm "github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm"
	"github.com/open-component-model/ocm/pkg/helm"
	"github.com/open-component-model/ocm/pkg/helm/loader"
	"github.com/open-component-model/ocm/pkg/utils"
)

func (o *Options) OCIContext() oci.Context {
	if o.Context == nil {
		return oci.DefaultContext()
	}
	return o.Context
}

func BlobAccessForHelmChart(path string, opts ...Option) (blob blobaccess.BlobAccess, name, version string, err error) {
	eff := optionutils.EvalOptions(opts...)
	ctx := eff.OCIContext()
	fs := utils.FileSystem(eff.FileSystem)
	printer := eff.Printer
	if printer == nil {
		printer = common.NewPrinter(nil)
	}

	var chartLoader loader.Loader
	if eff.HelmRepository == "" {
		if ok, err := vfs.Exists(fs, path); !ok || err != nil {
			return nil, "", "", errors.NewEf(err, "invalid file path %q", path)
		}
		chartLoader = loader.VFSLoader(path, fs)
	} else {
		cert := []byte(eff.CACert)
		if eff.CACertFile != "" {
			cert, err = vfs.ReadFile(fs, eff.CACertFile)
			if err != nil {
				return nil, "", "", errors.Wrapf(err, "cannot read root certificates from %q", eff.CACertFile)
			}
		}

		acc, err := helm.DownloadChart(printer, ctx, path, eff.Version, eff.HelmRepository,
			helm.WithCredentials(identity.GetCredentials(ctx, eff.HelmRepository, path)),
			helm.WithRootCert(cert))
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "cannot download chart %s:%s from %s", path, eff.Version, eff.HelmRepository)
		}
		chartLoader = loader.AccessLoader(acc)
	}

	defer errors.PropagateError(&err, chartLoader.Close)

	chart, err := chartLoader.Chart()
	if err != nil {
		return nil, "", "", err
	}
	vers := chart.Metadata.Version
	if vers == "" || optionutils.AsValue(eff.OverrideVersion) {
		vers = eff.Version
	}
	if vers == "" {
		return nil, "", "", fmt.Errorf("no version found or specified")
	}

	blob, err = chartLoader.ChartArtefactSet()
	if err == nil && blob == nil {
		blob, err = ocihelm.SynthesizeArtifactBlob(chartLoader)
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "cannot synthesize artifact blob")
		}
	}
	return blob, chart.Name(), vers, err
}

func BlobAccessProviderForHelmChart(name string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, _, _, err := BlobAccessForHelmChart(name, opts...)
		return b, err
	})
}
