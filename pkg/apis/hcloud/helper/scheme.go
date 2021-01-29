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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/install"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/validation"
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

func GetCloudProfileConfigFromProfile(profile *gardencorev1beta1.CloudProfile) (*hcloud.CloudProfileConfig, error) {
	var cloudProfileConfig *hcloud.CloudProfileConfig
	if profile.Spec.ProviderConfig != nil && profile.Spec.ProviderConfig.Raw != nil {
		cloudProfileConfig = &hcloud.CloudProfileConfig{}
		if _, _, err := decoder.Decode(profile.Spec.ProviderConfig.Raw, nil, cloudProfileConfig); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of cloudProfile")
		}
		// TODO validate cloud profile on admission instead
		if errs := validation.ValidateCloudProfileConfig(&profile.Spec, cloudProfileConfig); len(errs) > 0 {
			return nil, errors.Wrap(errs.ToAggregate(), "validation of providerConfig failed")
		}
	}
	return cloudProfileConfig, nil
}

func GetCloudProfileConfig(cluster *controller.Cluster) (*hcloud.CloudProfileConfig, error) {
	if cluster == nil {
		return nil, nil
	}
	if cluster.CloudProfile == nil {
		return nil, fmt.Errorf("missing cluster cloud profile")
	}
	cloudProfileConfig, err := GetCloudProfileConfigFromProfile(cluster.CloudProfile)
	if err != nil {
		return nil, errors.Wrapf(err, "shoot '%s'", cluster.ObjectMeta.Name)
	}
	return cloudProfileConfig, nil
}

// func GetControlPlaneConfig(cluster *controller.Cluster) (*hcloud.ControlPlaneConfig, error) {
// 	cpConfig := &hcloud.ControlPlaneConfig{}
// 	if cluster.Shoot.Spec.Provider.ControlPlaneConfig != nil {
// 		if _, _, err := decoder.Decode(cluster.Shoot.Spec.Provider.ControlPlaneConfig.Raw, nil, cpConfig); err != nil {
// 			return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", cluster.ObjectMeta.Name)
// 		}
// 	}
// 	return cpConfig, nil
// }

func GetInfrastructureStatus(name string, extension *runtime.RawExtension) (*hcloud.InfrastructureStatus, error) {
	if extension == nil || extension.Raw == nil {
		return nil, nil
	}
	infraStatus := &hcloud.InfrastructureStatus{}
	if _, _, err := decoder.Decode(extension.Raw, nil, infraStatus); err != nil {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", name)
	}
	return infraStatus, nil
	return nil, nil
}

// InfrastructureConfigFromInfrastructure extracts the InfrastructureConfig from the
// ProviderConfig section of the given Infrastructure.
func GetInfrastructureConfig(cluster *controller.Cluster) (*hcloud.InfrastructureConfig, error) {
	config := &hcloud.InfrastructureConfig{}
	if source := cluster.Shoot.Spec.Provider.InfrastructureConfig; source != nil && source.Raw != nil {
		if _, _, err := decoder.Decode(source.Raw, nil, config); err != nil {
			return nil, err
		}
		return config, nil
	}
	return config, nil
}

// func DecodeControlPlaneConfig(cp *runtime.RawExtension, fldPath *field.Path) (*hcloud.ControlPlaneConfig, error) {
// 	controlPlaneConfig := &hcloud.ControlPlaneConfig{}
// 	if err := util.Decode(decoder, cp.Raw, controlPlaneConfig); err != nil {
// 		return nil, field.Invalid(fldPath, string(cp.Raw), "cannot be decoded")
// 	}

// 	return controlPlaneConfig, nil
// }

func DecodeInfrastructureConfig(infra *runtime.RawExtension, fldPath *field.Path) (*hcloud.InfrastructureConfig, error) {
	infraConfig := &hcloud.InfrastructureConfig{}
	if err := util.Decode(decoder, infra.Raw, infraConfig); err != nil {
		return nil, field.Invalid(fldPath, string(infra.Raw), "cannot be decoded")
	}

	return infraConfig, nil
}

func DecodeCloudProfileConfig(config *runtime.RawExtension, fldPath *field.Path) (*hcloud.CloudProfileConfig, error) {
	cloudProfileConfig := &hcloud.CloudProfileConfig{}
	if err := util.Decode(decoder, config.Raw, cloudProfileConfig); err != nil {
		return nil, field.Invalid(fldPath, string(config.Raw), "cannot be decoded")
	}

	return cloudProfileConfig, nil
}
