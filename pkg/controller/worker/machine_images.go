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

// Package worker contains functions used at the worker controller
package worker

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	hcloudclient "github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
)

// findMachineImageName returns the image name for the given name and version values.
//
// PARAMETERS
// ctx     context.Context Execution context
// name    string          Machine image name
// version string          Machine image version
func (w *workerDelegate) findMachineImageName(ctx context.Context, name, version string, architecture *string) (string, error) {
	var arch hcloudclient.Architecture
	if architecture != nil {
		if *architecture == "arm" {
			arch = hcloudclient.ArchitectureARM
		} else {
			arch = hcloudclient.ArchitectureX86
		}
	} else {
		arch = hcloudclient.ArchitectureX86
	}

	machineImage, err := transcoder.DecodeMachineImageNameFromCloudProfile(w.cloudProfileConfig, name, version)
	if err == nil {
		return machineImage, nil
	}

	secret, err := w.getSecretData(ctx)
	if err != nil {
		return "", err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return "", err
	}

	client := apis.GetClientForToken(string(credentials.MCM().Token))

	opts := hcloudclient.ImageListOpts{
		Type:   []hcloudclient.ImageType{"system"},
		Status: []hcloudclient.ImageStatus{"available"},
	}

	images, _, err := client.Image.List(ctx, opts)
	if err != nil {
		return "", err
	}

	for _, image := range images {
		if image.OSFlavor != name || image.OSVersion != version || image.Architecture != arch {
			continue
		}

		return image.Name, nil
	}

	archStr := string(arch)
	return "", worker.ErrorMachineImageNotFound(name, version, archStr)
}

// UpdateMachineImagesStatus adds machineImages to the `WorkerStatus` resource.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) UpdateMachineImagesStatus(ctx context.Context) error {
	if w.machineImages == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	// Decode the current worker provider status.
	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return fmt.Errorf("unable to decode the worker provider status: %w", err)
	}

	workerStatus.MachineImages = w.machineImages

	if err := w.updateProviderStatus(ctx, workerStatus); err != nil {
		return fmt.Errorf("unable to update worker provider status: %w", err)
	}

	return nil
}
