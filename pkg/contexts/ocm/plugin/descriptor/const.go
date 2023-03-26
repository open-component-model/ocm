// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package descriptor

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	KIND_PLUGIN       = "plugin"
	KIND_DOWNLOADER   = "downloader"
	KIND_UPLOADER     = "uploader"
	KIND_ACCESSMETHOD = errors.KIND_ACCESSMETHOD
	KIND_ACTION       = action.KIND_ACTION
)

var REALM = logging.SubRealm("plugins")
