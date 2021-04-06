/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package helper

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/install"
)

var (
	// Scheme is a scheme with the types relevant for HCloud actuators.
	Scheme *runtime.Scheme

	decoder runtime.Decoder
)

func init() {
	Scheme = runtime.NewScheme()
	utilruntime.Must(install.AddToScheme(Scheme))

	decoder = serializer.NewCodecFactory(Scheme).UniversalDecoder()
}

func GetInfrastructureStatus(name string, extension *runtime.RawExtension) (*apis.InfrastructureStatus, error) {
	if extension == nil || extension.Raw == nil {
		return nil, nil
	}
	infraStatus := &apis.InfrastructureStatus{}
	if _, _, err := decoder.Decode(extension.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", name)
	}
	return infraStatus, nil
	return nil, nil
}

// InfrastructureConfigFromInfrastructure extracts the InfrastructureConfig from the
// ProviderConfig section of the given Infrastructure.
func GetInfrastructureConfig(cluster *controller.Cluster) (*apis.InfrastructureConfig, error) {
	config := &apis.InfrastructureConfig{}
	if source := cluster.Shoot.Spec.Provider.InfrastructureConfig; source != nil && source.Raw != nil {
		if _, _, err := decoder.Decode(source.Raw, nil, config); err != nil {
			return nil, err
		}
		return config, nil
	}
	return config, nil
}

func DecodeControlPlaneConfig(cp *runtime.RawExtension, fldPath *field.Path) (*apis.ControlPlaneConfig, error) {
	controlPlaneConfig := &apis.ControlPlaneConfig{}
	if err := util.Decode(decoder, cp.Raw, controlPlaneConfig); err != nil {
		return nil, field.Invalid(fldPath, string(cp.Raw), "cannot be decoded")
	}

	return controlPlaneConfig, nil
}

func DecodeInfrastructureConfig(infra *runtime.RawExtension, fldPath *field.Path) (*apis.InfrastructureConfig, error) {
	infraConfig := &apis.InfrastructureConfig{}
	if err := util.Decode(decoder, infra.Raw, infraConfig); err != nil {
		return nil, field.Invalid(fldPath, string(infra.Raw), "cannot be decoded")
	}

	return infraConfig, nil
}

func DecodeCloudProfileConfig(config *runtime.RawExtension, fldPath *field.Path) (*apis.CloudProfileConfig, error) {
	cloudProfileConfig := &apis.CloudProfileConfig{}
	if err := util.Decode(decoder, config.Raw, cloudProfileConfig); err != nil {
		return nil, field.Invalid(fldPath, string(config.Raw), "cannot be decoded")
	}

	return cloudProfileConfig, nil
}
