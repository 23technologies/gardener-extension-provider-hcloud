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

// Package ensurer provides functions used to ensure worker changes to be applied
package ensurer

import (
	"context"
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// EnsurePlacementGroups verifies that the placement groups requested are available.
//
// PARAMETERS
// ctx          context.Context  Execution context
// client       *hcloud.Client   HCloud client
// workerConfig *v1alpha1.Worker Worker config
func EnsurePlacementGroups(ctx context.Context, client *hcloud.Client, workerConfig *v1alpha1.Worker) (map[string]int, error) {
	placementGroupIDs := map[string]int{}

	labels := map[string]string{"hcloud.provider.extensions.gardener.cloud/role": "placement-group-v1"}

	for _, worker := range workerConfig.Spec.Pools {
		if worker.ProviderConfig == nil {
			continue
		}

		workerProviderConfig, err := transcoder.DecodeWorkerConfigFromRawExtension(worker.ProviderConfig)
		if err != nil {
			return placementGroupIDs, err
		}

		if workerProviderConfig.PlacementGroupType == "" {
			continue
		}

		name := fmt.Sprintf("%s-%s", workerConfig.Namespace, worker.Name)

		placementGroup, _, err := client.PlacementGroup.GetByName(ctx, name)
		if nil != err {
			return placementGroupIDs, err
		} else if placementGroup == nil {
			opts := hcloud.PlacementGroupCreateOpts{
				Name:   name,
				Labels: labels,
				Type:   hcloud.PlacementGroupTypeSpread,
			}

			placementGroupResult, _, err := client.PlacementGroup.Create(ctx, opts)
			if nil != err {
				return placementGroupIDs, err
			}

			placementGroup = placementGroupResult.PlacementGroup
		}

		placementGroupIDs[name] = placementGroup.ID
	}

	return placementGroupIDs, nil
}

// EnsurePlacementGroupDeleted removes any previously created placement group identified by the given fingerprint.
//
// PARAMETERS
// ctx         context.Context  Execution context
// client      *hcloud.Client   HCloud client
// fingerprint string           SSH fingerprint
func EnsurePlacementGroupDeleted(ctx context.Context, client *hcloud.Client, placementGroupID int) error {
	if placementGroupID != 0 {
		placementGroup, _, err := client.PlacementGroup.GetByID(ctx, placementGroupID)
		if nil != err {
			return err
		} else if placementGroup != nil {
			_, err := client.PlacementGroup.Delete(ctx, placementGroup)
			if nil != err {
				return err
			}
		}
	}

	return nil
}
