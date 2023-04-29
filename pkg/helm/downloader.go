// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"io"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/helm/credentials"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type chartDownloader struct {
	*downloader.ChartDownloader
	*chartAccess
	creds   common.Properties
	cacert  []byte
	keyring []byte
}

func DownloadChart(out io.Writer, ref, version, url string, opts ...Option) (ChartAccess, error) {
	acc, err := newTempChartAccess(osfs.New())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			acc.Close()
		}
	}()

	s := cli.EnvSettings{}

	dl := &chartDownloader{
		ChartDownloader: &downloader.ChartDownloader{
			Out:     out,
			Getters: getter.All(&s),
		},
		chartAccess: acc,
	}
	for _, o := range opts {
		err := o.apply(dl)
		if err != nil {
			return nil, err
		}
	}

	err = dl.complete(url)
	if err != nil {
		return nil, err
	}

	chart, _, err := dl.DownloadTo("repo/"+ref, version, dl.root)
	if err != nil {
		return nil, err
	}
	prov := chart + ".prov"
	if filepath.Exists(prov) {
		dl.prov = prov
	}
	dl.chart = chart
	return dl.chartAccess, nil
}

func (d *chartDownloader) complete(repourl string) error {
	rf := repo.NewFile()

	creds := d.creds
	if d.creds == nil {
		creds = common.Properties{}
	}

	entry := repo.Entry{
		Name:     "repo",
		URL:      repourl,
		Username: creds[credentials.ATTR_USERNAME],
		Password: creds[credentials.ATTR_PASSWORD],
	}

	config := vfs.Join(d.fs, d.root, ".config")
	err := d.fs.MkdirAll(config, 0o700)
	if err != nil {
		return err
	}
	cache := vfs.Join(d.fs, d.root, ".cache")
	err = d.fs.MkdirAll(cache, 0o700)
	if err != nil {
		return err
	}

	if len(d.cacert) != 0 {
		err = d.writeFile("cacert", config, &entry.CAFile, d.cacert, "CA file")
		if err != nil {
			return err
		}
	}
	if len(d.keyring) != 0 {
		err = d.writeFile("keyring", config, &d.Keyring, d.keyring, "keyring file")
		if err != nil {
			return err
		}
		d.Verify = downloader.VerifyIfPossible
	}
	if len(creds[credentials.ATTR_CERTIFICATE]) != 0 {
		err = d.writeFile("cert", config, &entry.CertFile, []byte(creds[credentials.ATTR_CERTIFICATE]), "certificate file")
		if err != nil {
			return err
		}
	}
	if len(creds[credentials.ATTR_PRIVATE_KEY]) != 0 {
		err = d.writeFile("private-key", config, &entry.KeyFile, []byte(creds[credentials.ATTR_PRIVATE_KEY]), "private key file")
		if err != nil {
			return err
		}
	}
	rf.Add(&entry)

	cr, err := repo.NewChartRepository(&entry, d.Getters)
	if err != nil {
		return errors.Wrapf(err, "cannot get chart repository %q", repourl)
	}

	d.RepositoryCache, cr.CachePath = cache, cache

	_, err = cr.DownloadIndexFile()
	if err != nil {
		return errors.Wrapf(err, "cannot download repository index for %q", repourl)
	}

	data, err := runtime.DefaultYAMLEncoding.Marshal(rf)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal repository file")
	}
	err = d.writeFile("repository", config, &d.RepositoryConfig, data, "repository config")
	if err != nil {
		return err
	}

	return nil
}

func (d *chartDownloader) writeFile(name, root string, path *string, data []byte, desc string) error {
	*path = vfs.Join(d.fs, root, name)
	err := vfs.WriteFile(d.fs, *path, data, 0o600)
	if err != nil {
		return errors.Wrapf(err, "cannot write %s %q", desc, *path)
	}
	return nil
}
