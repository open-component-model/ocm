// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spiff

import (
	"fmt"

	"github.com/mandelsoft/spiff/spiffing"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/v2/pkg/errors"
)

func ValidateSourceByScheme(src spiffing.Source, schemedata []byte) error {
	data, err := src.Data()
	if err != nil {
		return err
	}
	return ValidateByScheme(data, schemedata)
}

func ValidateByScheme(src []byte, schemedata []byte) error {
	data, err := yaml.YAMLToJSON(src)
	if err != nil {
		return errors.Wrapf(err, "converting data to json")
	}
	schemedata, err = yaml.YAMLToJSON(schemedata)
	if err != nil {
		return errors.Wrapf(err, "converting scheme to json")
	}
	documentLoader := gojsonschema.NewBytesLoader(data)

	scheme, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemedata))
	if err != nil {
		return errors.Wrapf(err, "invalid scheme")
	}
	res, err := scheme.Validate(documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		errs := res.Errors()
		errMsg := errs[0].String()
		for i := 1; i < len(errs); i++ {
			errMsg = fmt.Sprintf("%s;%s", errMsg, errs[i].String())
		}
		return errors.New(errMsg)
	}

	return nil
}
