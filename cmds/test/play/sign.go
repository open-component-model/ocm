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

package main

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

func sign() {

	capriv, capub, err := rsa.Handler{}.CreateKeyPair()

	CheckErr(err, "ca keypair")

	subject := pkix.Name{
		CommonName: "ca-authority",
	}
	caData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, capub, nil, capriv, true)
	CheckErr(err, "ca cert")

	ca, err := x509.ParseCertificate(caData)
	CheckErr(err, "ca")

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	CheckErr(err, "keypair")

	subject = pkix.Name{
		CommonName:    "mandelsoft",
		StreetAddress: []string{"some street 24"},
	}
	certData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, pub, ca, capriv, false)
	CheckErr(err, "ca cert")

	cert, err := x509.ParseCertificate(certData)
	CheckErr(err, "ca cert")

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	err = signing.VerifyCert(nil, pool, "mandelsoft", cert)
	CheckErr(err, "verify cert")

	err = signing.VerifyCert(nil, pool, "", cert)
	CheckErr(err, "verify anon cert")

	hasher := crypto.SHA256
	hash := hasher.New()
	hash.Write([]byte("test"))
	digest := hash.Sum(nil)
	sig, err := rsa.Handler{}.Sign(hex.EncodeToString(digest), hasher, "", priv)
	CheckErr(err, "sign")

	fmt.Printf("sig: %s\n", sig)
}
