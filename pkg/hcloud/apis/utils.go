// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apis is the main package for HCloud specific APIs
package apis

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

// GetRegionFromZone returns the region for a given zone string
//
// PARAMETERS
// zone string Zone
func GetRegionFromZone(zone string) string {
	zoneData := strings.SplitN(zone, "-", 2)
	return zoneData[0]
}

// GetSSHFingerprint returns the calculated fingerprint for an SSH public key.
//
// PARAMETERS
// publicKey []byte SSH public key
func GetSSHFingerprint(publicKey []byte) (string, error) {
	publicKeyData := strings.SplitN(string(publicKey), " ", 3)
	if len(publicKeyData) < 2 {
		return "", errors.New("SSH public key has invalid format")
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyData[1])
	if err != nil {
		return "", err
	}

	publicKeyMD5 := md5.Sum(publicKey)
	fingerprintArray := make([]string, len(publicKeyMD5))

	for i, c := range publicKeyMD5 {
		fingerprintArray[i] = hex.EncodeToString([]byte{c})
	}

	return strings.Join(fingerprintArray, ":"), nil
}
