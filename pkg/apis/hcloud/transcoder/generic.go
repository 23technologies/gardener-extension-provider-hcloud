/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package transcoder is used for API related object transformations
package transcoder

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

func DecodeSSHFingerprintFromPublicKey(publicKey []byte) (string, error) {
	if len(publicKey) == 0 {
		return "", fmt.Errorf("SSH public key given is empty")
	}

	publicKeyData := strings.SplitN(string(publicKey), " ", 3)
	if len(publicKeyData) < 2 {
		return "", fmt.Errorf("SSH public key has invalid format")
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

	fingerprint := strings.Join(fingerprintArray, ":")

	return fingerprint, nil
}

// MissingProviderConfig is raised when the requested ProviderConfig does not
// exist
type MissingProviderConfig struct{}

func (m *MissingProviderConfig) Error() string {
	return "Missing provider config"
}
