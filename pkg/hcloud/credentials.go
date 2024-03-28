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
	commonToken *Token
	ccmToken    *Token
	csiToken    *Token
	mcmToken    *Token
}

// CCM returns the token used for the Cloud Controller Manager.
func (c *Credentials) CCM() Token {
	if c.ccmToken != nil {
		return *c.ccmToken
	}
	return *c.commonToken
}

// CSI returns the token used for the Container Storage Interface driver.
func (c *Credentials) CSI() Token {
	if c.csiToken != nil {
		return *c.csiToken
	}
	return *c.commonToken
}

// MCM returns the token used for the Machine Controller Manager.
func (c *Credentials) MCM() Token {
	if c.mcmToken != nil {
		return *c.mcmToken
	}
	return *c.commonToken
}

// GetCredentials computes for a given context and infrastructure the corresponding credentials object.
//
// PARAMETERS
// ctx       context.Context        Execution context
// c         client.Client          Controller client
// secretRef corev1.SecretReference Secret reference to read credentials from
func GetCredentials(ctx context.Context, c client.Client, secretRef corev1.SecretReference) (*Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, c, &secretRef)
	if err != nil {
		return nil, err
	}
	return ExtractCredentials(secret)
}

// extractToken returns the token with the given key from the secret.
//
// PARAMETERS
// secret   *corev1.Secret Secret to get token from
// tokenKey string         Token key
func extractToken(secret *corev1.Secret, tokenKey string) (*Token, error) {
	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("missing %q field in secret", tokenKey)
	}

	return &Token{Token: string(token)}, nil
}

// ExtractCredentials generates a credentials object for a given provider secret.
//
// PARAMETERS
// secret   *corev1.Secret Secret to extract tokens from
func ExtractCredentials(secret *corev1.Secret) (*Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret does not contain any data")
	}

	commonToken, hcloudErr := extractToken(secret, HcloudToken)

	ccmToken, err := extractToken(secret, HcloudTokenCCM)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or cloud controller manager specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}
	csiToken, err := extractToken(secret, HcloudTokenCSI)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or container storage interface driver specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}
	mcmToken, err := extractToken(secret, HcloudTokenMCM)
	if err != nil && hcloudErr != nil {
		return nil, fmt.Errorf("Need either common or machine controller manager specific Hcloud account credentials: %s, %s", hcloudErr, err)
	}

	return &Credentials{
		commonToken: commonToken,
		ccmToken:    ccmToken,
		csiToken:    csiToken,
		mcmToken:    mcmToken,
	}, nil
}
