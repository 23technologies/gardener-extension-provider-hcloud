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

// Package controller provides functions to access controller specifications
package controller

import (
	"github.com/gardener/gardener/extensions/pkg/controller"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
)

func GetCloudProfileConfigFromControllerCluster(cluster *controller.Cluster) (*apis.CloudProfileConfig, error) {
	if cluster == nil {
		return nil, nil
	}

	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	return cloudProfileConfig, nil
}
