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
	"context"
	"errors"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller"
	webhookcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
)

// DecodeCloudProfileConfig extracts the CloudProfileConfig from the given
// RawExtension.
func DecodeCloudProfileConfig(profile *runtime.RawExtension) (*apis.CloudProfileConfig, error) {
	cpConfig, err := DecodeCloudProfileConfigWithDecoder(decoder, profile)
	if err != nil {
		return nil, err
	}

	return cpConfig, nil
}

// DecodeCloudProfileConfigWithDecoder extracts the CloudProfileConfig from the
// given RawExtension with the given decoder.
func DecodeCloudProfileConfigWithDecoder(decoder runtime.Decoder, profile *runtime.RawExtension) (*apis.CloudProfileConfig, error) {
	cpConfig := &apis.CloudProfileConfig{}

	if profile == nil || profile.Raw == nil {
		return nil, &MissingProviderConfig{}
	}

	if _, _, err := decoder.Decode(profile.Raw, nil, cpConfig); err != nil {
		return nil, fmt.Errorf("could not decode cpConfig: %w", err)
	}

	return cpConfig, nil
}

// DecodeCloudProfileConfigFromControllerCluster extracts the
// CloudProfileConfig from the ProviderConfig section of the given Cluster.
func DecodeCloudProfileConfigFromControllerCluster(cluster *controller.Cluster) (*apis.CloudProfileConfig, error) {
	if cluster == nil || cluster.CloudProfile == nil {
		return nil, errors.New("Missing cluster cloud profile")
	}

	cpConfig, err := DecodeConfigFromCloudProfile(cluster.CloudProfile)
	if err != nil {
		return nil, err
	}

	return cpConfig, nil
}

// DecodeCloudProfileConfigFromGardenContext extracts the CloudProfileConfig
// from the ProviderConfig section of the given GardenContext.
func DecodeCloudProfileConfigFromGardenContext(ctx context.Context, webhookcontext webhookcontext.GardenContext) (*apis.CloudProfileConfig, error) {
	cluster, err := webhookcontext.GetCluster(ctx)
	if err != nil {
		return nil, err
	}

	cpConfig, err := DecodeConfigFromCloudProfile(cluster.CloudProfile)
	if err != nil {
		return nil, err
	}

	return cpConfig, nil
}

// DecodeConfigFromCloudProfile extracts the CloudProfileConfig from the
// ProviderConfig section of the given CloudProfile.
func DecodeConfigFromCloudProfile(profile *v1beta1.CloudProfile) (*apis.CloudProfileConfig, error) {
	if profile == nil {
		return nil, errors.New("Missing cloud profile")
	}

	cpConfig, err := DecodeCloudProfileConfig(profile.Spec.ProviderConfig)
	if err != nil {
		return nil, err
	}

	return cpConfig, nil
}

// DecodeMachineImageNameFromCloudProfile takes a list of machine images, and the desired image name and version. It tries
// to find the image with the given name and version in the desired cloud profile. If it cannot be found then an error
// is returned.
func DecodeMachineImageNameFromCloudProfile(cpConfig *apis.CloudProfileConfig, imageName, imageVersion string) (string, error) {
	if cpConfig != nil {
		for _, machineImage := range cpConfig.MachineImages {
			if machineImage.Name != imageName {
				continue
			}
			for _, version := range machineImage.Versions {
				if imageVersion == version.Version {
					imageNameFound := version.ImageName
					if "" == imageNameFound {
						imageNameFound = fmt.Sprintf("%s-%s", machineImage.Name, version.Version)
					}

					return imageNameFound, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Could not find an image for name %q in version %q", imageName, imageVersion)
}
