/*
 * Copyright 2023 Mandelsoft. All rights reserved.
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

package logrusl

import (
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/logrusl/adapter"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/mandelsoft/logging/utils"
	"github.com/sirupsen/logrus"
)

// Settings is a composition environment to configure a
// logrus.Logger or a logging.Context.
type Settings struct {
	Writer    io.Writer
	Formatter logrus.Formatter
}

func (s Settings) WithWriter(w io.Writer) Settings {
	s.Writer = w
	return s
}

func (s Settings) WithFormatter(f logrus.Formatter) Settings {
	s.Formatter = f
	return s
}

func (s Settings) Human(padded ...bool) Settings {
	s.Formatter = adapter.NewTextFmtFormatter(padded...)
	return s
}

func (s Settings) JSON() Settings {
	s.Formatter = adapter.NewJSONFormatter()
	return s
}

func (s Settings) NewLogr() logr.Logger {
	return logrusr.New(s.NewLogrus())
}

func (s Settings) NewLogrus() *logrus.Logger {
	logger := adapter.NewLogger()
	logger.Out = s.Writer
	if logger.Out == nil {
		logger.Out = utils.NewSyncWriter(os.Stderr)
	}
	logger.Formatter = s.Formatter
	if logger.Formatter == nil {
		logger.Formatter = adapter.NewTextFormatter()
	}
	return logger
}

func (s Settings) New() logging.Context {
	return logging.New(logrusr.New(s.NewLogrus()))
}

////////////////////////////////////////////////////////////////////////////////

func New() logging.Context {
	return Settings{}.New()
}

func Adapter() Settings {
	return Settings{}
}

func WithWriter(w io.Writer) Settings {
	return Settings{}.WithWriter(w)
}

func WithFormatter(f logrus.Formatter) Settings {
	return Settings{}.WithFormatter(f)
}

func Human(padded ...bool) Settings {
	return Settings{}.Human(padded...)
}

func JSON() Settings {
	return Settings{}.JSON()
}
