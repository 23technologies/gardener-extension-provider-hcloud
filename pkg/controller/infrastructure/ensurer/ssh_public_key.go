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

package ensurer

import (
	"context"
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func EnsureSSHPublicKey(ctx context.Context, client *hcloud.Client, publicKey []byte) error {
	fingerprint, err := transcoder.DecodeSSHFingerprintFromPublicKey(publicKey)
	if nil != err {
		return err
	}

	labels := map[string]string{ "hcloud.provider.extensions.gardener.cloud/role": "infrastructure-ssh-v1" }

	sshKey, _, err := client.SSHKey.GetByFingerprint(ctx, fingerprint)
	if nil != err {
		return err
	} else if sshKey == nil {
		opts := hcloud.SSHKeyCreateOpts{
			Name: fmt.Sprintf("infrastructure-ssh-%s", fingerprint),
			PublicKey: string(publicKey),
			Labels: labels,
		}

		_, _, err := client.SSHKey.Create(ctx, opts)
		if nil != err {
			return err
		}
	}

	return nil
}
