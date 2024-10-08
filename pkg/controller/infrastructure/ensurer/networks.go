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
	"net"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/controller"
)

// EnsureNetworks verifies the network resources requested are available.
//
// PARAMETERS
// ctx       context.Context                    Execution context
// client    *hcloud.Client                     HCloud client
// namespace string                             Shoot namespace
// zone      string                             Shoot zone
// networks  *apis.InfrastructureConfigNetworks Networks struct
func EnsureNetworks(ctx context.Context, client *hcloud.Client, namespace, zone string, networks *apis.InfrastructureConfigNetworks) (int64, error) {
	workersConfiguration := networks.WorkersConfiguration

	if nil == workersConfiguration && "" != networks.Workers {
		workersConfiguration = &apis.InfrastructureConfigNetwork{
			Cidr: networks.Workers,
		}
	}

	if nil != workersConfiguration {
		if "" == workersConfiguration.Zone {
			locationName := apis.GetRegionFromZone(zone)

			locations, err := client.Location.All(ctx)
			if nil != err {
				return -1, err
			}

			for _, location := range locations {
				if locationName == location.Name {
					workersConfiguration.Zone = location.NetworkZone
					break
				}
			}

			if "" == workersConfiguration.Zone {
				return -1, fmt.Errorf("Failed to find matching location for zone %q", zone)
			}
		}

		name := fmt.Sprintf("%s-workers", namespace)

		network, _, err := client.Network.GetByName(ctx, name)
		if nil != err {
			return -1, err
		} else if network == nil {
			_, ipRange, _ := net.ParseCIDR(workersConfiguration.Cidr)

			labels := map[string]string{"hcloud.provider.extensions.gardener.cloud/role": "workers-network-v1"}

			opts := hcloud.NetworkCreateOpts{
				Name:    name,
				IPRange: ipRange,
				Subnets: []hcloud.NetworkSubnet{
					hcloud.NetworkSubnet{
						Type:        hcloud.NetworkSubnetTypeCloud,
						IPRange:     ipRange,
						NetworkZone: workersConfiguration.Zone,
					}},
				Labels: labels,
			}

			network, _, err = client.Network.Create(ctx, opts)
			if nil != err {
				return -1, err
			}

			resultData := ctx.Value(controller.CtxWrapDataKey("MethodData")).(*controller.InfrastructureReconcileMethodData)
			resultData.NetworkID = network.ID
		}

		return network.ID, nil
	}

	return -1, nil
}

// EnsureNetworksDeleted removes any previously created network resources.
//
// PARAMETERS
// ctx       context.Context                      Execution context
// client    *hcloud.Client                       HCloud client
// namespace string                               Shoot namespace
// networks  *apis.InfrastructureConfigNetworkIDs Network IDs struct
func EnsureNetworksDeleted(ctx context.Context, client *hcloud.Client, namespace string, networks *apis.InfrastructureConfigNetworkIDs) error {
	if networks != nil && "" != networks.Workers {
		name := fmt.Sprintf("%s-workers", namespace)

		network, _, err := client.Network.GetByName(ctx, name)
		if nil != err {
			return err
		} else if network != nil {
			_, err := client.Network.Delete(ctx, network)
			if nil != err {
				return err
			}
		}
	}

	return nil
}
