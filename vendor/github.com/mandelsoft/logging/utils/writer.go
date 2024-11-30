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

package utils

import (
	"io"
	"sync"
)

// SyncWriter executes synchronized Write (and With) operations.
// logr and underlying logger implementations typically
// use a synchronized write for log records, but they do not
// offer an api to interfere with the log record write
// operations.
// But this is urgently required to inject logs provided
// by another program execution.
// A SynWriter may be used as a bad alternative to
// wrap the originally used writer allowing additional synchronized
// write operations. But this only works together with log sinks,
// if the log record is written by a single Write or ReadFrom
// operation.
// Unfortunately io.Copy prioritizes the WriteTo method
// from the reader over the ReadFrom method from the writer.
// For example, logrus uses a single (synchronized) io.Copy call to
// write a record.
type SyncWriter interface {
	Write(buf []byte) (int, error)
	WriteWith(f func(writer io.Writer) (int64, error)) (int64, error)
}

func NewSyncWriter(w io.Writer) SyncWriter {
	return &syncWriter{writer: w}
}

type syncWriter struct {
	lock   sync.Mutex
	writer io.Writer
}

func (w *syncWriter) Write(buf []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.writer.Write(buf)
}

func (w *syncWriter) ReadFrom(r io.Reader) (n int64, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return io.Copy(w.writer, r)
}

// With gets a writer function, which is executed under the writer lock.
func (w *syncWriter) WriteWith(f func(writer io.Writer) (int64, error)) (int64, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return f(w.writer)
}

func (w *syncWriter) Unwrap() io.Writer {
	return w.writer
}

func UnwrapWriter(w io.Writer) io.Writer {
	for w != nil {
		if u, ok := w.(interface{ Unwrap() io.Writer }); ok {
			w = u.Unwrap()
		} else {
			break
		}
	}
	return w
}
