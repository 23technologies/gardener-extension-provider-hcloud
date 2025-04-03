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

// Package ensurer provides functions used to ensure infrastructure changes to be applied
package ensurer

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/controller"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/transcoder"
)

// EnsureSSHPublicKey verifies that the SSH public key resource requested is available.
//
// PARAMETERS
// ctx       context.Context  Execution context
// client    *hcloud.Client   HCloud client
// publicKey []byte           SSH public key
func EnsureSSHPublicKey(ctx context.Context, client *hcloud.Client, cluster *extensionscontroller.Cluster, infra *extensionsv1alpha1.Infrastructure) (string, error) {
	publicKey := infra.Spec.SSHPublicKey

	if len(publicKey) == 0 {
		return "", fmt.Errorf("SSH public key given is empty")
	}

	oldProviderStatus, err := transcoder.DecodeInfrastructureStatus(infra.Status.GetProviderStatus())
	if nil != err {
		return "", err
	}

	oldFingerprint := oldProviderStatus.SSHFingerprint

	fingerprint, err := api.GetSSHFingerprint(publicKey)
	if nil != err {
		return "", err
	}

	if oldFingerprint != fingerprint {
		err := EnsureSSHPublicKeyDeleted(ctx, client, oldFingerprint)
		if nil != err {
			return "", err
		}
	}

	labels := map[string]string{
		"cluster.gardener.cloud/id":                      string(cluster.Shoot.GetUID()),
		"cluster.gardener.cloud/name":                    cluster.Shoot.Name,
		"hcloud.provider.extensions.gardener.cloud/role": "infrastructure-ssh-v1",
	}

	sshKey, _, err := client.SSHKey.GetByFingerprint(ctx, fingerprint)
	if nil != err {
		return "", err
	} else if sshKey == nil {
		opts := hcloud.SSHKeyCreateOpts{
			Name:      fmt.Sprintf("infrastructure-ssh-%s", fingerprint),
			PublicKey: string(publicKey),
			Labels:    labels,
		}

		sshKey, _, err := client.SSHKey.Create(ctx, opts)
		if nil != err {
			return "", err
		}

		resultData := ctx.Value(controller.CtxWrapDataKey("MethodData")).(*controller.InfrastructureReconcileMethodData)
		resultData.SSHKeyID = sshKey.ID
	}

	return fingerprint, nil
}

// EnsureSSHPublicKeyDeleted removes any previously created SSH public key resource identified by the given fingerprint.
//
// PARAMETERS
// ctx         context.Context  Execution context
// client      *hcloud.Client   HCloud client
// fingerprint string           SSH fingerprint
func EnsureSSHPublicKeyDeleted(ctx context.Context, client *hcloud.Client, fingerprint string) error {
	if "" != fingerprint {
		sshKey, _, err := client.SSHKey.GetByFingerprint(ctx, fingerprint)
		if nil != err {
			return err
		} else if sshKey != nil {
			_, err := client.SSHKey.Delete(ctx, sshKey)
			if nil != err {
				return err
			}
		}
	}

	return nil
}
