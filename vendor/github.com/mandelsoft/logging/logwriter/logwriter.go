/*
 * Copyright 2024 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package logwriter

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/logging/logrusr"
)

// LogWriter is an optional interface the logr.LogSink  of a passed logr.Logger
// can implement to expose the technical writer used as final
// log sink.
type LogWriter interface {
	LogWriter() io.Writer
}

// DetermineLogWriter tries to determine the technical writer
// used as log sink by unwrapping the sink and
// checking for the logging.LogWriter interface
// or a logrus wrapper.
func DetermineLogWriter(s logr.LogSink) io.Writer {
	for {
		w, ok := logrusr.LogWriter(s)
		if ok {
			return w
		}
		if lw, ok := s.(LogWriter); ok && w != nil {
			return lw.LogWriter()
		}
		u, ok := s.(interface{ Unwrap() logr.LogSink })
		if ok {
			m := u.Unwrap()
			if m != s && m != nil {
				s = m
				continue
			}
		}
		return nil
	}
}
