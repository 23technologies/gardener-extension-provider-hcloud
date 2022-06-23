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
	"strconv"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/controller"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// ensurePlacementGroupDeleted removes any previously created placement group identified by the given fingerprint.
//
// PARAMETERS
// ctx              context.Context Execution context
// client           *hcloud.Client  HCloud client
// placementGroupID string          Placement group ID
func ensurePlacementGroupDeleted(ctx context.Context, client *hcloud.Client, placementGroupID string) error {
	id, err := strconv.Atoi(placementGroupID)
	if nil != err {
		return err
	}

	placementGroup, _, err := client.PlacementGroup.GetByID(ctx, id)
	if nil != err {
		return err
	} else if placementGroup != nil {
		_, err := client.PlacementGroup.Delete(ctx, placementGroup)
		if nil != err {
			return err
		}
	}

	return nil
}

// EnsurePlacementGroups verifies that the placement groups requested are available.
//
// PARAMETERS
// ctx                       context.Context      Execution context
// client                    *hcloud.Client       HCloud client
// namespace                 string               Shoot namespace
// workers                   []corev1beta1.Worker Worker specifications
// placementGroupQuantityMap map[string]int       List of placement group quantities requested
func EnsurePlacementGroups(ctx context.Context, client *hcloud.Client, namespace string, workers []corev1beta1.Worker, placementGroupQuantityMap map[string]int) (map[string][]string, error) {
	placementGroupIDs := map[string][]string{ }
	labels := map[string]string{ "hcloud.provider.extensions.gardener.cloud/role": "placement-group-v1" }

	for _, worker := range workers {
		placementGroupQuantity, ok := placementGroupQuantityMap[worker.Name]

		if !ok {
			placementGroupQuantity, ok = placementGroupQuantityMap["*"]

			if !ok {
				placementGroupQuantity = 1
			}
		}

		if placementGroupQuantity < 1 {
			continue
		}

		placementGroupIDs[worker.Name] = []string{}

		for i := 1; i <= placementGroupQuantity; i++ {
			name := fmt.Sprintf("%s-%s-%d", namespace, worker.Name, i)
			var placementGroupID string

			placementGroup, _, err := client.PlacementGroup.GetByName(ctx, name)
			if nil != err {
				return placementGroupIDs, err
			} else if placementGroup != nil {
				placementGroupID = strconv.Itoa(placementGroup.ID)
			} else {
				opts := hcloud.PlacementGroupCreateOpts{
					Name: name,
					Labels: labels,
					Type: hcloud.PlacementGroupTypeSpread,
				}

				placementGroupResult, _, err := client.PlacementGroup.Create(ctx, opts)
				if nil != err {
					return placementGroupIDs, err
				}

				placementGroupID = strconv.Itoa(placementGroupResult.PlacementGroup.ID)

				resultData := ctx.Value(controller.CtxWrapDataKey("MethodData")).(*controller.InfrastructureReconcileMethodData)
				resultData.PlacementGroupIDs = append(resultData.PlacementGroupIDs, placementGroupID)
			}

			placementGroupIDs[worker.Name] = append(placementGroupIDs[worker.Name], placementGroupID)
		}
	}

	return placementGroupIDs, nil
}

// EnsurePlacementGroupsDeleted removes any previously created placement group identified by the given fingerprint.
//
// PARAMETERS
// ctx               context.Context Execution context
// client            *hcloud.Client  HCloud client
// placementGroupIDs []string        Placement group IDs
func EnsurePlacementGroupsDeleted(ctx context.Context, client *hcloud.Client, placementGroupIDs []string) error {
	var err error

	for _, id := range placementGroupIDs {
		if "" == id {
			continue
		}

		err = ensurePlacementGroupDeleted(ctx, client, id)
		if nil != err {
			return err
		}
	}

	return err
}
