// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"context"
	"time"

	"github.com/mandelsoft/logging"
	helmclient "github.com/mittwald/go-helm-client"
)

func Install(l logging.Logger, path string, release string, namespace string, createNamespace bool, values []byte, kubeconfig []byte) error {
	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace,
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
			DebugLog:         DebugFunction(l),
		},
		KubeContext: "",
		KubeConfig:  kubeconfig,
	}

	helmClient, err := helmclient.NewClientFromKubeConf(opt, helmclient.Burst(100), helmclient.Timeout(30*time.Second))
	if err != nil {
		return err
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     release,
		ChartName:       path,
		Namespace:       namespace,
		ValuesYaml:      string(values),
		UpgradeCRDs:     true,
		CreateNamespace: createNamespace,
		Timeout:         100 * time.Second,
		Wait:            true,
	}

	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		return err
	}

	return nil
}
