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

	hcloud "github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/worker/ensurer"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
)

// PreReconcileHook is a hook called at the beginning of the worker reconciliation flow.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) PreReconcileHook(ctx context.Context) error {
	test, _, _ := w.hclient.ServerType.List(ctx, hcloud.ServerTypeListOpts{})
	srvTypeIdToName := make(map[int]string, len(test))

	for _, srvType := range(test) {
		srvTypeIdToName[srvType.ID] = srvType.Name
	}

	for _, pool := range(w.worker.Spec.Pools) {
		// currently there is only one zone per region on hetzner.
		zone := pool.Zones[0]
		dc, _, _ := w.hclient.Datacenter.Get(ctx, zone)

		machineTypeAvailabe := false
		for _, curServerType := range(dc.ServerTypes.Available) {
			if pool.MachineType == srvTypeIdToName[curServerType.ID] {
				machineTypeAvailabe = true
				break
			}
		}
		if machineTypeAvailabe == false {
			return fmt.Errorf("Machine Type %s is currently not availbe in %s", pool.MachineType, dc.Name)
		}
	}

	placementGroupIDs, err := ensurer.EnsurePlacementGroups(ctx, w.hclient, w.worker)
	if err != nil {
		return err
	}

	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return fmt.Errorf("unable to decode the worker provider status: %w", err)
	}

	workerStatus.PlacementGroupIDs = placementGroupIDs

	updateErr := w.updateProviderStatus(ctx, workerStatus)
	if updateErr != nil {
		return fmt.Errorf("%s: %w", err.Error(), updateErr)
	}

	return nil
}

// PostReconcileHook is a hook called at the end of the worker reconciliation flow.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) PostReconcileHook(_ context.Context) error {
	return nil
}

// PreDeleteHook is a hook called at the beginning of the worker deletion flow.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) PreDeleteHook(_ context.Context) error {
	return nil
}

// PostDeleteHook is a hook called at the end of the worker deletion flow.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) PostDeleteHook(ctx context.Context) error {
	deleteAllPlacementGroups := w.worker.DeletionTimestamp != nil

	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return err
	}

	for _, worker := range w.worker.Spec.Pools {
		// if there is no placementgroup in the workerstatus for current pool,
		// mark it for deletion
		deletePlacementGroup := deleteAllPlacementGroups

		name := fmt.Sprintf("%s-%s", w.worker.Namespace, worker.Name)
		_, ok := workerStatus.PlacementGroupIDs[name]
		if !ok {
			deletePlacementGroup = true
		}

		if deletePlacementGroup {
			err := ensurer.EnsurePlacementGroupDeleted(ctx, w.hclient, workerStatus.PlacementGroupIDs[name])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeployMachineDependencies should deploy dependencies for the worker node machines.
// DEPRECATED, use PreReconcileHook and PostReconcileHook instead.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) DeployMachineDependencies(ctx context.Context) error {
	return nil
}

// CleanupMachineDependencies should clean up dependencies previously deployed for the worker node machines.
// DEPRECATED, use PreDeleteHook and PostDeleteHook instead.
//
// PARAMETERS
// _ context.Context Execution context
func (w *workerDelegate) CleanupMachineDependencies(ctx context.Context) error {
	return nil
}
