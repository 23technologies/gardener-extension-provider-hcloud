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

// Package v1alpha1 provides hcloud.provider.extensions.gardener.cloud/v1alpha1
package v1alpha1

import (
	"unsafe"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
)

// Add non-generated conversion functions
func addConversionFuncs(scheme *runtime.Scheme) error {
	if err := scheme.AddConversionFunc(
		(*InfrastructureStatus)(nil),
		(*apis.InfrastructureStatus)(nil),
		func(in, out interface{}, scope conversion.Scope) error {
			return Convert_v1alpha1_InfrastructureStatus_To_apis_InfrastructureStatus(in.(*InfrastructureStatus), out.(*apis.InfrastructureStatus), scope)
		},
	); err != nil {
		return err
	}

	if err := scheme.AddConversionFunc(
		(*apis.InfrastructureStatus)(nil),
		(*InfrastructureStatus)(nil),
		func(in, out interface{}, scope conversion.Scope) error {
			return Convert_apis_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in.(*apis.InfrastructureStatus), out.(*InfrastructureStatus), scope)
		},
	); err != nil {
		return err
	}

	return nil
}

func Convert_v1alpha1_InfrastructureStatus_To_apis_InfrastructureStatus(in *InfrastructureStatus, out *apis.InfrastructureStatus, scope conversion.Scope) error {
	out.SSHFingerprint = in.SSHFingerprint

	if in.PlacementGroupIDs != nil {
		in, out := &in.PlacementGroupIDs, &out.PlacementGroupIDs
		*out = make(map[string][]string, len(*in))
		for key, val := range *in {
			(*out)[key] = []string{ val }
		}
	} else if in.PlacementGroupID != "" {
		out.PlacementGroupIDs = map[string][]string{ "worker": []string{ in.PlacementGroupID } }
	} else {
		out.PlacementGroupIDs = nil
	}

	out.FloatingPoolName = in.FloatingPoolName
	out.NetworkIDs = (*apis.InfrastructureConfigNetworkIDs)(unsafe.Pointer(in.NetworkIDs))

	return nil
}

func Convert_apis_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *apis.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	out.SSHFingerprint = in.SSHFingerprint

	if in.PlacementGroupIDs != nil {
		in, out := &in.PlacementGroupIDs, &out.PlacementGroupIDs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			newVal := new(string)
			if err := runtime.Convert_Slice_string_To_string(&val, newVal, s); err != nil {
				return err
			}
			(*out)[key] = *newVal
		}
	} else {
		out.PlacementGroupIDs = nil
	}

	out.FloatingPoolName = in.FloatingPoolName
	out.NetworkIDs = (*InfrastructureConfigNetworkIDs)(unsafe.Pointer(in.NetworkIDs))

	return nil
}
