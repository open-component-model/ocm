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

package rsa

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

func GetPublicKey(key interface{}) (*rsa.PublicKey, []string, error) {
	var err error
	if data, ok := key.([]byte); ok {
		key, err = ParseKey(data)
		if err != nil {
			return nil, nil, err
		}
	}
	switch k := key.(type) {
	case *rsa.PublicKey:
		return k, nil, nil
	case *rsa.PrivateKey:
		return &k.PublicKey, nil, nil
	case *x509.Certificate:
		switch p := k.PublicKey.(type) {
		case *rsa.PublicKey:
			return p, k.DNSNames, nil
		}
		return nil, nil, fmt.Errorf("unknown key public key %T in certificate", k)
	default:
		return nil, nil, fmt.Errorf("unknown key specification %T", k)
	}
}

func GetPrivateKey(key interface{}) (*rsa.PrivateKey, error) {
	if data, ok := key.([]byte); ok {
		return ParsePrivateKey(data)
	}
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return k, nil
	default:
		return nil, fmt.Errorf("unknown key specification %T", k)
	}
}

func WriteKeyData(key interface{}, w io.Writer) error {
	block := PemBlockForKey(key)
	return pem.Encode(w, block)
}

func KeyData(key interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	block := PemBlockForKey(key)
	err := pem.Encode(buf, block)
	return buf.Bytes(), err
}

func PemBlockForKey(priv interface{}, gen ...bool) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PublicKey:
		if len(gen) > 0 && gen[0] {
			bytes, err := x509.MarshalPKIXPublicKey(k)
			if err != nil {
				panic(err)
			}
			return &pem.Block{Type: "PUBLIC KEY", Bytes: bytes}
		}
		return &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(k)}
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	default:
		panic("invalid key")
		return nil
	}
}

func ParseKey(data []byte) (interface{}, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid key format (expected pem block)")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "CERTIFICATE":
		return x509.ParseCertificate(block.Bytes)
	}
	return ParsePublicKey(data)
}

func ParsePublicKey(data []byte) (interface{}, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid public key format (expected pem block)")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DER encoded public key: %s", err)
		}
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unknown type of public key")
	}
}

func ParsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid private key format (expected pem block)")
	}
	x509Encoded := block.Bytes
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(x509Encoded)
	default:
		untypedPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing key %w", err)
		}
		key, ok := untypedPrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not of type *rsa.GetPrivateKey: %T", untypedPrivateKey)
		}
		return key, nil
	}
}
