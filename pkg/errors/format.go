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

package errors

type ErrorFormatter interface {
	Format(kind string, elem *string, ctx string) string
}

type defaultFormatter struct {
	verb        string
	msg         string
	preposition string
}

func NewDefaultFormatter(verb, msg, preposition string) ErrorFormatter {
	if verb != "" {
		verb = verb + " "
	}
	return &defaultFormatter{
		verb:        verb,
		msg:         msg,
		preposition: preposition,
	}
}

func (f *defaultFormatter) Format(kind string, elem *string, ctx string) string {
	if ctx != "" {
		ctx = " " + f.preposition + " " + ctx
	}
	elems := ""
	if elem != nil {
		elems = "\"" + *elem + "\" "
	}
	if kind != "" {
		kind = kind + " "
	}
	if kind == "" && elems == "" {
		return f.msg + ctx
	}
	return kind + elems + f.verb + f.msg + ctx
}
