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

package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/open-component-model/ocm/pkg/common"

	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/consts"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func Merge(values ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	for _, val := range values {
		for k, v := range val {
			result[k] = v
		}
	}
	return result
}

func Execute(d driver.Driver, action string, ctx ocm.Context, octx out.Context, cv ocm.ComponentVersionAccess, cfg *Config, values map[string]interface{}, kubeconfig []byte) error {

	if action != "install" && action != "uninstall" {
		return errors.ErrNotSupported("action", action)
	}
	cfgv, err := cfg.GetValues()
	if err != nil {
		return err
	}
	values = Merge(cfgv, values)

	acc, rcv, err := utils.ResolveResourceReference(cv, cfg.Chart, nil)
	if err != nil {
		return err
	}

	if cv != rcv {
		defer rcv.Close()
	}

	fmt.Printf("Installing helm chart from resource %s@%s\n", cfg.Chart, common.VersionedElementKey(cv))
	if acc.Meta().Type != consts.HelmChart {
		return errors.Newf("resource type %q required, but found %q", consts.HelmChart, acc.Meta().Type)
	}

	// have to use the OS filesystem here for using the helm library
	file, err := ioutil.TempFile("", "helm-*")
	if err != nil {
		return err
	}

	path := file.Name()
	file.Close()
	os.Remove(path)

	_, path, err = download.For(ctx).Download(octx, acc, path, osfs.New())
	if err != nil {
		return errors.Wrapf(err, "downloading helm chart")
	}
	defer os.Remove(path)

	for i, v := range cfg.ImageMapping {
		acc, rcv, err := utils.ResolveResourceReference(cv, v.ResourceReference, nil)
		if err != nil {
			return errors.Wrapf(err, "mapping %d (%s)", i+1, &v.ResourceReference)
		}
		if rcv != cv {
			rcv.Close()
		}
		ref, err := utils.GetOCIArtefactRef(ctx, acc)
		if err != nil {
			return errors.Wrapf(err, "mapping %d: cannot resolve resource %s to an OCI Reference", i+1, v)
		}
		ix := strings.Index(ref, ":")
		if ix < 0 {
			ix = strings.Index(ref, "@")
			if ix < 0 {
				return errors.Wrapf(err, "mapping %d: image tag or digest missing (%s)", i+1, ref)
			}
		}
		repo := ref[:ix]
		tag := ref[ix+1:]
		if v.Repository != "" {
			err = Set(values, v.Repository, repo)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning repositry to property %q", v.Repository)
			}
		}
		if v.Tag != "" {
			err = Set(values, v.Tag, tag)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning tag to property %q", v.Tag)
			}
		}
		if v.Image != "" {
			err = Set(values, v.Image, ref)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning image to property %q", v.Image)
			}
		}

	}
	ns := "default"
	if cfg.Namespace != "" {
		ns = cfg.Namespace
	}
	if s, ok := values["namespace"].(string); ok && s != "" {
		ns = s
	}
	release := cfg.Release
	if s, ok := values["release"].(string); ok && s != "" {
		release = s
	}
	valuesdata, err := runtime.DefaultYAMLEncoding.Marshal(values)
	if err != nil {
		return errors.Wrapf(err, "marshal values")
	}

	dcfg := &driver.Config{
		ChartPath:       path,
		Release:         release,
		Namespace:       ns,
		CreateNamespace: cfg.CreateNamespace,
		Values:          valuesdata,
		Kubeconfig:      kubeconfig,
	}
	switch action {
	case "install":
		return d.Install(dcfg)
	case "uninstall":
		return d.Uninstall(dcfg)
	default:
		return errors.ErrNotImplemented("action", action)
	}
}
