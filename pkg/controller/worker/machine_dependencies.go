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

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	hetzner "github.com/hetznercloud/hcloud-go/hcloud"
)

// DeployMachineDependencies should deploy dependencies for the worker node machines.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) DeployMachineDependencies(ctx context.Context) error {
	hclient := w.hclient

	placementGroupIds := map[string]int{}
	labels := map[string]string{"hcloud.provider.extensions.gardener.cloud/role": "placement-group-v1"}
	for _, worker := range w.worker.Spec.Pools {
		if worker.ProviderConfig == nil {
			continue
		}

		name := fmt.Sprintf("%s-%s", w.worker.Namespace, worker.Name)

		workerConfig, err := transcoder.DecodeWorkerConfigFromRawExtension(worker.ProviderConfig)
		if err != nil {
			return err
		}

		if workerConfig.PlacementGroupType == "" {
			continue
		}

		placementGroup, _, err := hclient.PlacementGroup.GetByName(ctx, name)
		if nil != err {
			return err
		} else if placementGroup == nil {
			opts := hetzner.PlacementGroupCreateOpts{
				Name:   name,
				Labels: labels,
				Type:   hetzner.PlacementGroupTypeSpread,
			}

			placementGroupResult, _, err := hclient.PlacementGroup.Create(ctx, opts)
			if nil != err {
				return err
			}

			placementGroup = placementGroupResult.PlacementGroup
		}

		placementGroupIds[name] = placementGroup.ID
	}

	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return fmt.Errorf("unable to decode the worker provider status: %w", err)
	}

	w.updateMachineDependenciesStatus(ctx, workerStatus, placementGroupIds, nil)

	return nil
}

// CleanupMachineDependencies should clean up dependencies previously deployed for the worker node machines.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) CleanupMachineDependencies(ctx context.Context) error {

	hclient := w.hclient

	deleteAllPlacementGroups := w.worker.DeletionTimestamp != nil
	deleteCurrentPlacementGroup := false

	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return err
	}

	for _, worker := range w.worker.Spec.Pools {
		// if there is no placementgroup in the workerstatus for current pool,
		// mark it for deletion
		name := fmt.Sprintf("%s-%s", w.worker.Namespace, worker.Name)
		_, ok := workerStatus.PlacementGroupIDs[name]
		if !ok {
			deleteCurrentPlacementGroup = true
		}

		if deleteAllPlacementGroups || deleteCurrentPlacementGroup {
			placementGroup, _, err := hclient.PlacementGroup.GetByName(ctx, name)
			if err != nil {
				return err
			} else if placementGroup != nil {
				hclient.PlacementGroup.Delete(ctx, placementGroup)
			}
		}
	}
	return nil
}
