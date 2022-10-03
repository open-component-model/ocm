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

package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	crypto "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// PrintPrettyYaml prints the given objects as yaml if enabled.
func PrintPrettyYaml(obj interface{}, enabled bool) {
	if !enabled {
		return
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		logrus.Errorf("unable to serialize object as yaml: %s", err)

		return
	}

	//nolint: forbidigo // Intentional Println.
	fmt.Println(string(data))
}

// GetFileType returns the mimetype of a file.
func GetFileType(fs vfs.FileSystem, path string) (string, error) {
	file, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// see http://golang.org/pkg/net/http/#DetectContentType for the 512 bytes
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buf), nil
}

// CleanMarkdownUsageFunc removes markdown tags from the long usage of the command.
// With this func it is possible to generate the markdown docs but still have readable commandline help func.
// Note: currently only "<pre>" tags are removed.
func CleanMarkdownUsageFunc(cmd *cobra.Command) {
	defaultHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, s []string) {
		cmd.Long = strings.ReplaceAll(cmd.Long, "<pre>", "")
		cmd.Long = strings.ReplaceAll(cmd.Long, "</pre>", "")
		defaultHelpFunc(cmd, s)
	})
}

// RawJSON converts an arbitrary value to json.RawMessage.
func RawJSON(value interface{}) (*json.RawMessage, error) {
	jsonval, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&jsonval), nil
}

// Gzip applies gzip compression to an arbitrary byte slice.
func Gzip(data []byte, compressionLevel int) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	gzipWriter, err := gzip.NewWriterLevel(buf, compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	if _, err = gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("unable to write to stream: %w", err)
	}

	if err = gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("unable to close writer: %w", err)
	}

	return buf.Bytes(), nil
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

// RandomString creates a new random string with the given length.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		var value int
		if v, err := crypto.Int(crypto.Reader, big.NewInt(int64(len(chars)))); err == nil {
			value = int(v.Int64())
		} else {
			// insecure fallback to provide a valid result
			logrus.Warnf("failed to generate random number: %s", err)
			value = rand.Intn(len(chars)) //nolint: gosec // only used as fallback
		}
		b[i] = chars[value]
	}
	return string(b)
}

// SafeConvert converts a byte slice to string.
// If the byte slice is nil, an empty string is returned.
func SafeConvert(bytes []byte) string {
	if bytes == nil {
		return ""
	}

	return string(bytes)
}

const (
	BYTE = 1.0 << (10 * iota)
	KIBIBYTE
	MEBIBYTE
	GIBIBYTE
)

// BytesString converts bytes into a human readable string.
// This function is inspired by https://www.reddit.com/r/golang/comments/8micn7/review_bytes_to_human_readable_format/
func BytesString(bytes uint64, accuracy int) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= GIBIBYTE:
		unit = "GiB"
		value /= GIBIBYTE
	case bytes >= MEBIBYTE:
		unit = "MiB"
		value /= MEBIBYTE
	case bytes >= KIBIBYTE:
		unit = "KiB"
		value /= KIBIBYTE
	case bytes >= BYTE:
		unit = "bytes"
	case bytes == 0:
		return "0"
	}

	stringValue := strings.TrimSuffix(
		fmt.Sprintf("%.2f", value), "."+strings.Repeat("0", accuracy),
	)

	return fmt.Sprintf("%s %s", stringValue, unit)
}

// WriteFileToTARArchive writes a new file with name=filename and content=contentReader to archiveWriter.
func WriteFileToTARArchive(filename string, contentReader io.Reader, archiveWriter *tar.Writer) error {
	if filename == "" {
		return errors.New("filename must not be empty")
	}

	if contentReader == nil {
		return errors.New("contentReader must not be nil")
	}

	if archiveWriter == nil {
		return errors.New("archiveWriter must not be nil")
	}

	tempfile, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("unable to create tempfile: %w", err)
	}
	defer tempfile.Close()

	fsize, err := io.Copy(tempfile, contentReader)
	if err != nil {
		return fmt.Errorf("unable to copy content to tempfile: %w", err)
	}

	if _, err := tempfile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to beginning of tempfile: %w", err)
	}

	header := tar.Header{
		Name:    filename,
		Size:    fsize,
		Mode:    0o600,
		ModTime: time.Now(),
	}

	if err := archiveWriter.WriteHeader(&header); err != nil {
		return fmt.Errorf("unable to write tar header: %w", err)
	}

	if _, err := io.Copy(archiveWriter, tempfile); err != nil {
		return fmt.Errorf("unable to write file to tar archive: %w", err)
	}

	return nil
}

func IndentLines(orig string, gap string, skipfirst ...bool) string {
	return JoinIndentLines(strings.Split(strings.TrimPrefix(orig, "\n"), "\n"), gap, skipfirst...)
}

func JoinIndentLines(orig []string, gap string, skipfirst ...bool) string {
	skip := false
	for _, b := range skipfirst {
		skip = skip || b
	}

	s := ""
	for _, l := range orig {
		if !skip {
			s += gap
		}
		s += l + "\n"
		skip = false
	}
	return s
}

func StringMapKeys(m interface{}) []string {
	if m == nil {
		return nil
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		panic(fmt.Sprintf("%T is no map", m))
	}
	if v.Type().Key().Kind() != reflect.String {
		panic(fmt.Sprintf("map key of %T is no string", m))
	}

	keys := []string{}
	for _, k := range v.MapKeys() {
		keys = append(keys, k.Interface().(string))
	}
	sort.Strings(keys)
	return keys
}
