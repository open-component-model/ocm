// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/errors"
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
	s := cli.EnvSettings{}

	dl := &chartDownloader{
		ChartDownloader: &downloader.ChartDownloader{
			Out:              out,
			Verify:           0,
			Keyring:          "",
			Getters:          getter.All(&s),
			Options:          nil,
			RepositoryConfig: "",
			RepositoryCache:  "",
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

	chart, p, err := dl.DownloadTo(ref, version, dl.root)
	if err != nil {
		return nil, err
	}
	if p != nil {
		dl.prov = dl.chart + ".prov"
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
		Name:                  "default",
		URL:                   repourl,
		Username:              creds[ATTR_USERNAME],
		Password:              creds[ATTR_PASSWORD],
		CertFile:              "",
		KeyFile:               "",
		CAFile:                "",
		InsecureSkipTLSverify: false,
		PassCredentialsAll:    false,
	}

	config := vfs.Join(d.fs, d.root, ".config")
	err := d.fs.MkdirAll(config, 0o700)
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
		err = d.writeFile("keyring", config, &entry.KeyFile, d.keyring, "key file")
		if err != nil {
			return err
		}
	}
	if len(creds[ATTR_CERTIFICATE]) != 0 {
		err = d.writeFile("cert", config, &entry.CertFile, []byte(creds[ATTR_CERTIFICATE]), "certificate file")
		if err != nil {
			return err
		}
	}
	rf.Add(&entry)

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
