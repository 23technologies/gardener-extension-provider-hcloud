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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestControlPlaneInfrastructureProviderStatus = `{
		"apiVersion": "hcloud.provider.extensions.gardener.cloud/v1alpha1",
		"kind": "InfrastructureStatus",
		"networkIDs": {"workers": "42"}
	}`
	TestControlPlaneName           = "xyz"
	TestControlPlaneProviderConfig = `{
		"apiVersion": "hcloud.provider.extensions.gardener.cloud/v1alpha1",
		"kind": "ControlPlaneConfig"
	}`
	TestControlPlaneSecretName = "cloudprovider"
)

// NewControlPlane generates a new provider specification for testing purposes.
func NewControlPlane() *v1alpha1.ControlPlane {
	return &v1alpha1.ControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions.gardener.cloud",
			Kind:       "ControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestControlPlaneName,
			Namespace: TestNamespace,
		},
		Spec: v1alpha1.ControlPlaneSpec{
			DefaultSpec: v1alpha1.DefaultSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: []byte(TestControlPlaneProviderConfig),
				},
			},
			SecretRef: corev1.SecretReference{
				Name:      TestControlPlaneSecretName,
				Namespace: TestNamespace,
			},
			InfrastructureProviderStatus: &runtime.RawExtension{
				Raw: []byte(TestControlPlaneInfrastructureProviderStatus),
			},
			Region: TestRegion,
		},
	}
}

// ManipulateControlPlane changes given provider specification.
//
// PARAMETERS
// cp   *v1alpha1.ControlPlane ControlPlane specification
// data map[string]interface{} Members to change
func ManipulateControlPlane(cp *v1alpha1.ControlPlane, data map[string]interface{}) *v1alpha1.ControlPlane {
	for key, value := range data {
		if strings.Index(key, "ObjectMeta") == 0 {
			manipulateStruct(&cp.ObjectMeta, key[11:], value)
		} else if strings.Index(key, "Spec") == 0 {
			manipulateStruct(&cp.Spec, key[7:], value)
		} else if strings.Index(key, "TypeMeta") == 0 {
			manipulateStruct(&cp.TypeMeta, key[9:], value)
		} else {
			manipulateStruct(&cp, key, value)
		}
	}

	return cp
}
