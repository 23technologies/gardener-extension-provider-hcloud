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
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// EnsurePlacementGroup verifies that the placement group requested is available.
//
// PARAMETERS
// ctx       context.Context  Execution context
// client    *hcloud.Client   HCloud client
// namespace string           Shoot namespace
// zone      string           Shoot zone
func EnsurePlacementGroup(ctx context.Context, client *hcloud.Client, namespace string) (int, error) {
	labels := map[string]string{ "hcloud.provider.extensions.gardener.cloud/role": "placement-group-v1" }

	placementGroup, _, err := client.PlacementGroup.GetByName(ctx, namespace)
	if nil != err {
		return -1, err
	} else if placementGroup == nil {
		opts := hcloud.PlacementGroupCreateOpts{
			Name: namespace,
			Labels: labels,
			Type: hcloud.PlacementGroupTypeSpread,
		}

		placementGroupResult, _, err := client.PlacementGroup.Create(ctx, opts)
		if nil != err {
			return -1, err
		}

		placementGroup = placementGroupResult.PlacementGroup
	}

	return placementGroup.ID, nil
}

// EnsurePlacementGroupDeleted removes any previously created placement group identified by the given fingerprint.
//
// PARAMETERS
// ctx         context.Context  Execution context
// client      *hcloud.Client   HCloud client
// fingerprint string           SSH fingerprint
func EnsurePlacementGroupDeleted(ctx context.Context, client *hcloud.Client, placementGroupID string) error {
	if "" != placementGroupID {
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
	}

	return nil
}
