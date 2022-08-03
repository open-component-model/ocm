// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package helm

import (
	"context"
	"fmt"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
)

func Install(path string, release string, namespace string, createNamespace bool, values []byte, kubeconfig []byte) error {
	opt := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace,
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true,
			DebugLog: func(format string, v ...interface{}) {
				fmt.Printf(format+"\n", v...)
			},
		},
		KubeContext: "",
		KubeConfig:  kubeconfig,
	}

	helmClient, err := helmclient.NewClientFromKubeConf(opt, helmclient.Burst(100), helmclient.Timeout(10*time.Second))
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
