// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"github.com/gardener/gardener/extensions/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
)

func decodeControlPlaneConfig(decoder runtime.Decoder, cp *runtime.RawExtension) (*hcloud.ControlPlaneConfig, error) {
	controlPlaneConfig := &hcloud.ControlPlaneConfig{}
	if err := util.Decode(decoder, cp.Raw, controlPlaneConfig); err != nil {
		return nil, err
	}

	return controlPlaneConfig, nil
}

func decodeInfrastructureConfig(decoder runtime.Decoder, infra *runtime.RawExtension) (*hcloud.InfrastructureConfig, error) {
	infraConfig := &hcloud.InfrastructureConfig{}
	if err := util.Decode(decoder, infra.Raw, infraConfig); err != nil {
		return nil, err
	}

	return infraConfig, nil
}

func decodeCloudProfileConfig(decoder runtime.Decoder, config *runtime.RawExtension) (*hcloud.CloudProfileConfig, error) {
	cloudProfileConfig := &hcloud.CloudProfileConfig{}
	if err := util.Decode(decoder, config.Raw, cloudProfileConfig); err != nil {
		return nil, err
	}

	return cloudProfileConfig, nil
}
