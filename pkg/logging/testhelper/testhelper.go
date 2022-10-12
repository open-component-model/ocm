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

package testhelper

import (
	"github.com/mandelsoft/logging"
)

func LogTest(ctx logging.ContextProvider, prefix ...string) {
	LoggerTest(ctx.LoggingContext().Logger(), prefix...)
}

func LoggerTest(logger logging.Logger, prefix ...string) {
	p := ""
	for _, e := range prefix {
		p += e
	}
	logger.Trace(p + "trace")
	logger.Debug(p + "debug")
	logger.Info(p + "info")
	logger.Warn(p + "warn")
	logger.Error(p + "error")
}
