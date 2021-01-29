/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package worker

import (
	"context"
	"fmt"

	apishcloud "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	apishcloudhelper "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/helper"
)

// GetMachineImages returns the used machine images for the `Worker` resource.
func (w *workerDelegate) UpdateMachineImagesStatus(ctx context.Context) error {
	if w.machineImages == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	// Decode the current worker provider status.
	workerStatus, err := w.decodeWorkerProviderStatus()
	if err != nil {
		return err
	}

	workerStatus.MachineImages = w.machineImages
	return w.updateWorkerProviderStatus(ctx, workerStatus)
}

func errorMachineImageNotFound(name, version string) error {
	return fmt.Errorf("could not find machine image for %s/%s neither in componentconfig nor in worker status", name, version)
}

func appendMachineImage(machineImages []apishcloud.MachineImage, machineImage apishcloud.MachineImage) []apishcloud.MachineImage {
	if _, err := apishcloudhelper.FindMachineImage(machineImages, machineImage.Name, machineImage.Version); err != nil {
		return append(machineImages, machineImage)
	}
	return machineImages
}
