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

// Package hcloud provides types and functions used for HCloud interaction
package hcloud

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Token struct {
	Token string
}

// Credentials contains the necessary HCloud credential information.
type Credentials struct {
	hcloud    *Token
	hcloudMCM *Token
	hcloudCCM *Token
	hcloudCSI *Token
}

func (c *Credentials) HcloudMCM() Token {
	if c.hcloudMCM != nil {
		return *c.hcloudMCM
	}
	return *c.hcloud
}

func (c *Credentials) HcloudCCM() Token {
	if c.hcloudCCM != nil {
		return *c.hcloudCCM
	}
	return *c.hcloud
}

func (c *Credentials) HcloudCSI() Token {
	if c.hcloudCSI != nil {
		return *c.hcloudCSI
	}
	return *c.hcloud
}

// GetCredentials computes for a given context and infrastructure the corresponding credentials object.
func GetCredentials(ctx context.Context, c client.Client, secretRef corev1.SecretReference) (*Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, c, &secretRef)
	if err != nil {
		return nil, err
	}
	return ExtractCredentials(secret)
}

func extractUserPass(secret *corev1.Secret, tokenKey string) (*Token, error) {
	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", tokenKey)
	}

	return &Token{Token: string(token)}, nil
}

// ExtractCredentials generates a credentials object for a given provider secret.
func ExtractCredentials(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret does not contain any data")
	}

	hcloud, hcloudErr := extractUserPass(secret, HcloudToken)

	mcm, err := extractUserPass(secret, HcloudTokenMCM)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or machine controller manager specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}
	ccm, err := extractUserPass(secret, HcloudTokenCCM)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or cloud controller manager specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}
	csi, err := extractUserPass(secret, HcloudTokenCSI)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or cloud controller manager specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}

	return &Credentials{
		hcloud:    hcloud,
		hcloudMCM: mcm,
		hcloudCCM: ccm,
		hcloudCSI: csi,
	}, nil
}
