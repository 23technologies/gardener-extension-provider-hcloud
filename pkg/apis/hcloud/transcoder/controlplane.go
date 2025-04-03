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

// Package transcoder is used for API related object transformations
package transcoder

import (
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
)

// DecodeControlPlaneConfig extracts the ControlPlaneConfig from the
// given RawExtension.
func DecodeControlPlaneConfig(cp *runtime.RawExtension) (*api.ControlPlaneConfig, error) {
	controlPlaneConfig, err := DecodeControlPlaneConfigWithDecoder(decoder, cp)
	if err != nil {
		return nil, err
	}

	return controlPlaneConfig, nil
}

// DecodeControlPlaneConfigWithDecoder extracts the ControlPlaneConfig from the
// given RawExtension with the given decoder.
func DecodeControlPlaneConfigWithDecoder(decoder runtime.Decoder, cp *runtime.RawExtension) (*api.ControlPlaneConfig, error) {
	controlPlaneConfig := &api.ControlPlaneConfig{}

	if cp == nil || cp.Raw == nil {
		return nil, &MissingProviderConfig{}
	}

	if _, _, err := decoder.Decode(cp.Raw, nil, controlPlaneConfig); err != nil {
		return nil, fmt.Errorf("could not decode controlPlaneConfig: %w", err)
	}

	return controlPlaneConfig, nil
}

// DecodeControlPlaneConfigFromControllerCluster extracts the
// ControlPlaneConfig from the ProviderConfig section of the given Cluster.
func DecodeControlPlaneConfigFromControllerCluster(cluster *controller.Cluster) (*api.ControlPlaneConfig, error) {
	controlPlaneConfig, err := DecodeControlPlaneConfig(cluster.Shoot.Spec.Provider.ControlPlaneConfig)
	if err != nil {
		return nil, err
	}

	return controlPlaneConfig, nil
}

// DecodeInfrastructureStatusFromControlPlane extracts the InfrastructureStatus
// from the ProviderStatus section of the given ControlPlane.
func DecodeInfrastructureStatusFromControlPlane(controlPlane *v1alpha1.ControlPlane) (*api.InfrastructureStatus, error) {
	infraStatus, err := DecodeInfrastructureStatus(controlPlane.Spec.InfrastructureProviderStatus)
	if err != nil {
		return nil, err
	}

	return infraStatus, nil
}
