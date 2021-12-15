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
	"errors"
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/validation"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	errorhelpers "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// DecodeInfrastructureConfig extracts the InfrastructureConfig from the
// given RawExtension.
func DecodeInfrastructureConfig(infra *runtime.RawExtension) (*apis.InfrastructureConfig, error) {
	infraConfig, err := DecodeInfrastructureConfigWithDecoder(decoder, infra)
	if err != nil {
		return nil, err
	}

	return infraConfig, nil
}

// DecodeInfrastructureConfigWithDecoder extracts the InfrastructureConfig from
// the given RawExtension with the given decoder.
func DecodeInfrastructureConfigWithDecoder(decoder runtime.Decoder, infra *runtime.RawExtension) (*apis.InfrastructureConfig, error) {
	infraConfig := &apis.InfrastructureConfig{}

	if infra == nil || infra.Raw == nil {
		return nil, &MissingProviderConfig{}
	}

	if _, _, err := decoder.Decode(infra.Raw, nil, infraConfig); err != nil {
		return nil, errorhelpers.Wrapf(err, "could not decode providerConfig")
	}

	return infraConfig, nil
}

// DecodeInfrastructureConfigFromCluster extracts the InfrastructureConfig from the
// ProviderConfig section of the given Infrastructure.
func DecodeInfrastructureConfigFromCluster(cluster *controller.Cluster) (*apis.InfrastructureConfig, error) {
	infraConfig, err := DecodeInfrastructureConfig(cluster.Shoot.Spec.Provider.InfrastructureConfig)
	if err != nil {
		return nil, err
	}

	return infraConfig, nil
}

// DecodeInfrastructureConfigFromInfrastructure extracts the
// InfrastructureConfig from the ProviderConfig section of the given
// Infrastructure.
func DecodeInfrastructureConfigFromInfrastructure(infra *v1alpha1.Infrastructure) (*apis.InfrastructureConfig, error) {
	infraConfig, err := DecodeInfrastructureConfig(infra.Spec.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if errs := validation.ValidateInfrastructureConfigSpec(infraConfig); len(errs) > 0 {
		return nil, fmt.Errorf("Error while validating ProviderSpec %v", errs)
	}

	return infraConfig, nil
}

// DecodeInfrastructureStatus extracts the InfrastructureStatus from the
// given RawExtension.
func DecodeInfrastructureStatus(infra *runtime.RawExtension) (*apis.InfrastructureStatus, error) {
	infraStatus := &apis.InfrastructureStatus{}

	if infra == nil || infra.Raw == nil {
		return nil, errors.New("Missing infrastructure status")
	}

	if _, _, err := decoder.Decode(infra.Raw, nil, infraStatus); err != nil {
		return nil, errorhelpers.Wrapf(err, "could not decode infrastructureStatus")
	}

	return infraStatus, nil
}

func DecodeInfrastructureStatusFromInfrastructure(infra *v1alpha1.Infrastructure) (*apis.InfrastructureStatus, error) {
	infraStatus, err := DecodeInfrastructureStatus(infra.Status.ProviderStatus)
	if err != nil {
		return nil, err
	}

	return infraStatus, nil
}
