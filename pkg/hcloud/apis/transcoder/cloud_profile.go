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

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/validation"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	errorhelpers "github.com/pkg/errors"
)

func DecodeCloudProfileConfigFromControllerCluster(cluster *controller.Cluster) (*apis.CloudProfileConfig, error) {
	if cluster == nil || cluster.CloudProfile == nil {
		return nil, errors.New("Missing cluster cloud profile")
	}

	cloudProfileConfig, err := DecodeConfigFromCloudProfile(cluster.CloudProfile)
	if err != nil {
		return nil, err
	}
	return cloudProfileConfig, nil
}

func DecodeConfigFromCloudProfile(profile *v1beta1.CloudProfile) (*apis.CloudProfileConfig, error) {
	cloudProfileConfig := &apis.CloudProfileConfig{}

	if profile.Spec.ProviderConfig == nil || profile.Spec.ProviderConfig.Raw == nil {
		return nil, errors.New("Missing cloud profile")
	}

	if _, _, err := decoder.Decode(profile.Spec.ProviderConfig.Raw, nil, cloudProfileConfig); err != nil {
		return nil, errorhelpers.Wrapf(err, "could not decode providerConfig")
	}

	if errs := validation.ValidateCloudProfileConfig(&profile.Spec, cloudProfileConfig); len(errs) > 0 {
		return nil, errorhelpers.Wrap(errs.ToAggregate(), "validation of providerConfig failed")
	}

	return cloudProfileConfig, nil
}
