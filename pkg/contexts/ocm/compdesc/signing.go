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

package compdesc

import (
	"encoding/hex"
	"fmt"
	"hash"

	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
)

// NormalisationAlgorithm types and versions the algorithm used for digest generation.
type NormalisationAlgorithm = string

const (
	JsonNormalisationV1 NormalisationAlgorithm = "jsonNormalisation/v1"
)

const (
	KIND_HASH_ALGORITHM = "hash algorithm"
	KIND_SIGN_ALGORITHM = "signing algorithm"
	KIND_PUBLIC_KEY     = "public key"
	KIND_PRIVATE_KEY    = "private key"
	KIND_SIGNATURE      = "signature"
)

// isNormalizeable checks if componentReferences and resources contain digest.
// Resources are allowed to omit the digest, if res.access.type == None or res.access == nil.
// Does NOT verify if the digests are correct
func (cd *ComponentDescriptor) isNormalizeable() error {
	// check for digests on component references
	for _, reference := range cd.ComponentReferences {
		if reference.Digest == nil || reference.Digest.HashAlgorithm == "" || reference.Digest.NormalisationAlgorithm == "" || reference.Digest.Value == "" {
			return fmt.Errorf("missing digest in componentReference for %s:%s", reference.Name, reference.Version)
		}
	}
	for _, res := range cd.Resources {
		if (res.Access != nil && res.Access.GetType() != "None") && res.Digest == nil {
			return fmt.Errorf("missing digest in resource for %s:%s", res.Name, res.Version)
		}
		if (res.Access == nil || res.Access.GetType() == "None") && res.Digest != nil {
			return fmt.Errorf("digest for resource with emtpy (None) access not allowed %s:%s", res.Name, res.Version)
		}
	}
	return nil
}

// Hash return the hash for the component-descriptor, if it is normalizeable
// (= componentReferences and resources contain digest field)
func Hash(cd *ComponentDescriptor, normAlgo string, hash hash.Hash) (string, error) {
	if hash == nil {
		return metav1.NoDigest, nil
	}
	cv := DefaultSchemes[cd.SchemaVersion()]
	if cv == nil {
		if cv == nil {
			return "", errors.ErrNotSupported(errors.KIND_SCHEMAVERSION, cd.SchemaVersion())
		}
	}
	v, err := cv.ConvertFrom(cd)
	if err != nil {
		return "", err
	}
	normalized, err := v.Normalize(normAlgo)
	if err != nil {
		return "", fmt.Errorf("failed normalising component descriptor %w", err)
	}
	hash.Reset()
	if _, err = hash.Write(normalized); err != nil {
		return "", fmt.Errorf("failed hashing the normalisedComponentDescriptorJson: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Sign signs the given component-descriptor with the signer.
// The component-descriptor has to contain digests for componentReferences and resources.
func Sign(cd *ComponentDescriptor, privateKey interface{}, signer signing.Signer, hasher signing.Hasher, signatureName string) error {
	digest, err := Hash(cd, JsonNormalisationV1, hasher.Create())
	if err != nil {
		return fmt.Errorf("failed getting hash for cd: %w", err)
	}

	signature, err := signer.Sign(digest, privateKey)
	if err != nil {
		return fmt.Errorf("failed signing hash of normalised component descriptor, %w", err)
	}
	cd.Signatures = append(cd.Signatures, metav1.Signature{
		Name: signatureName,
		Digest: metav1.DigestSpec{
			HashAlgorithm:          hasher.Algorithm(),
			NormalisationAlgorithm: JsonNormalisationV1,
			Value:                  digest,
		},
		Signature: metav1.SignatureSpec{
			Algorithm: signature.Algorithm,
			Value:     signature.Value,
			MediaType: signature.MediaType,
		},
	})
	return nil
}

// Verify verifies the signature (selected by signatureName) and hash of the component-descriptor (as specified in the signature).
// Does NOT resolve resources or referenced component-descriptors.
// Returns error if verification fails.
func Verify(cd *ComponentDescriptor, registry signing.Registry, signatureName string) error {
	//find matching signature
	matchingSignature := cd.SelectSignatureByName(signatureName)
	if matchingSignature == nil {
		return errors.ErrNotFound(KIND_SIGNATURE, signatureName)
	}
	verifier := registry.GetVerifier(matchingSignature.Signature.Algorithm)
	if verifier == nil {
		return errors.ErrUnknown(KIND_SIGN_ALGORITHM, matchingSignature.Signature.Algorithm)
	}
	publicKey := registry.GetPublicKey(signatureName)
	if verifier == nil {
		return errors.ErrNotFound(KIND_PUBLIC_KEY, signatureName)
	}

	//Verify author of signature
	err := verifier.Verify(matchingSignature.Digest.Value, matchingSignature.Signature.Value, matchingSignature.Signature.MediaType, publicKey)
	if err != nil {
		return fmt.Errorf("failed verifying: %w", err)
	}

	//get hasher by algorithm name
	hasher := registry.GetHasher(matchingSignature.Digest.HashAlgorithm)
	if hasher == nil {
		return errors.ErrUnknown(KIND_HASH_ALGORITHM, matchingSignature.Digest.HashAlgorithm)
	}

	hash := hasher.Create()
	//Verify normalised cd to given (and verified) hash
	calculatedDigest, err := Hash(cd, matchingSignature.Digest.NormalisationAlgorithm, hash)
	if err != nil {
		return fmt.Errorf("failed hashing cd %s:%s: %w", cd.Name, cd.Version, err)
	}

	if calculatedDigest != matchingSignature.Digest.Value {
		return fmt.Errorf("normalised component-descriptor does not match hash from signature")
	}

	return nil
}

// SelectSignatureByName returns the Signature (Digest and SigantureSpec) matching the given name
func (cd *ComponentDescriptor) SelectSignatureByName(signatureName string) *metav1.Signature {
	for _, signature := range cd.Signatures {
		if signature.Name == signatureName {
			return &signature
		}
	}
	return nil
}
