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

// Package mock provides all methods required to simulate a HCloud provider environment
package mock

import (
	"strings"

	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/extensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestClusterCloudProfile = `{
		"apiVersion": "core.gardener.cloud/v1alpha1",
		"kind": "CloudProfile",
		"spec": {
			"regions": [{"name": "hel1", "zones": [{"name": "hel1-dc2"}]}],
			"machineTypes": [{"name": "cx11"}],
			"providerConfig": {
				"apiVersion": "hcloud.provider.extensions.gardener.cloud/v1alpha1",
				"kind": "CloudProfileConfig",
				"regions": [{"name": "hel1"}],
				"machineImages": [{"name": "ubuntu", "versions": [{"version": "20.04"}]}],
				"machineTypes": [{"name": "cx11"}]
			}
		}
	}`
	TestClusterName = "xyz"
	TestClusterSeed = `{
		"apiVersion": "core.gardener.cloud/v1alpha1",
		"kind": "Seed"
	}`
	TestClusterShoot = `{
		"apiVersion": "core.gardener.cloud/v1alpha1",
		"kind": "Shoot",
		"spec": {
			"kubernetes": {"version": "1.13.4"},
			"cloud": {"hcloud": {"test": "foo"}},
			"region": "hel1",
			"status": {
				"lastOperation": {"state": "Succeeded"}
			}
		}
	}`
)

// DecodeCluster returns a decoded cluster structure.
//
// PARAMETERS
// cluster *v1alpha1.Cluster Cluster specification
func DecodeCluster(cluster *v1alpha1.Cluster) (*extensions.Cluster, error) {
	decoder := extensions.NewGardenDecoder()

	cloudProfile, err := extensions.CloudProfileFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}

	seed, err := extensions.SeedFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}

	shoot, err := extensions.ShootFromCluster(decoder, cluster)
	if err != nil {
		return nil, err
	}

	return &extensions.Cluster{cluster.ObjectMeta, cloudProfile, seed, shoot}, nil
}

// NewCluster generates a new provider specification for testing purposes.
func NewCluster() *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions.gardener.cloud",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestClusterName,
			Namespace: TestNamespace,
		},
		Spec: v1alpha1.ClusterSpec{
			CloudProfile: runtime.RawExtension{
				Raw: []byte(TestClusterCloudProfile),
			},
			Seed: runtime.RawExtension{
				Raw: []byte(TestClusterSeed),
			},
			Shoot: runtime.RawExtension{
				Raw: []byte(TestClusterShoot),
			},
		},
	}
}

// ManipulateCluster changes given provider specification.
//
// PARAMETERS
// cluster *v1alpha1.Cluster      Cluster specification
// data    map[string]interface{} Members to change
func ManipulateCluster(cluster *v1alpha1.Cluster, data map[string]interface{}) *v1alpha1.Cluster {
	for key, value := range data {
		if (strings.Index(key, "ObjectMeta") == 0) {
			manipulateStruct(&cluster.ObjectMeta, key[11:], value)
		} else if (strings.Index(key, "Spec") == 0) {
			manipulateStruct(&cluster.Spec, key[7:], value)
		} else if (strings.Index(key, "TypeMeta") == 0) {
			manipulateStruct(&cluster.TypeMeta, key[9:], value)
		} else {
			manipulateStruct(&cluster, key, value)
		}
	}

	return cluster
}
