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

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/pkg/controllerutils"
	hcloudclient "github.com/hetznercloud/hcloud-go/hcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

// findMachineImageName returns the image name for the given name and version values.
//
// PARAMETERS
// ctx     context.Context Execution context
// name    string          Machine image name
// version string          Machine image version
func (w *workerDelegate) findMachineImageName(ctx context.Context, name, version string) (string, error) {
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
		Type: []hcloudclient.ImageType{"system"},
		Status: []hcloudclient.ImageStatus{"available"},
	}

	images, _, err := client.Image.List(ctx, opts)
	if nil != err {
		return "", err
	}

	for _, image := range images {
		if image.OSFlavor != name || image.OSVersion != version {
			continue
		}

		return image.Name, nil
	}

	return "", worker.ErrorMachineImageNotFound(name, version)
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

	var workerStatus *apis.WorkerStatus
	var workerStatusV1alpha1 *v1alpha1.WorkerStatus

	if w.worker.Status.ProviderStatus == nil {
		workerStatus = &apis.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
			MachineImages: w.machineImages,
		}
	} else {
		// Decode the current worker provider status.
		decodedWorkerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
		if err != nil {
			return err
		}

		workerStatus = decodedWorkerStatus
		workerStatus.MachineImages = w.machineImages

		workerStatusV1alpha1 = &v1alpha1.WorkerStatus{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeGroupVersion.String(),
				Kind:       "WorkerStatus",
			},
		}
	}

	workerStatusV1alpha1 = &v1alpha1.WorkerStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "WorkerStatus",
		},
	}

	if err := w.Scheme().Convert(workerStatus, workerStatusV1alpha1, nil); err != nil {
		return err
	}

	return controllerutils.TryUpdateStatus(ctx, retry.DefaultBackoff, w.Client(), w.worker, func() error {
		w.worker.Status.ProviderStatus = &runtime.RawExtension{Object: workerStatusV1alpha1}
		return nil
	})
}
