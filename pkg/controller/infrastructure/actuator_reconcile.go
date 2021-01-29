/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package infrastructure

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	apishcloud "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	apishelper "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/helper"
)

type preparedReconcile struct {
	cloudProfileConfig *apishcloud.CloudProfileConfig
	infraConfig        *apishcloud.InfrastructureConfig
	region             *apishcloud.RegionSpec
}

func (a *actuator) prepareReconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) (*preparedReconcile, error) {
	cloudProfileConfig, err := apishelper.GetCloudProfileConfig(cluster)
	if err != nil {
		return nil, err
	}

	infraConfig, err := apishelper.GetInfrastructureConfig(cluster)
	if err != nil {
		return nil, err
	}

	region := apishelper.FindRegion(infra.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile", infra.Spec.Region)
	}

	prepared := &preparedReconcile{
		cloudProfileConfig: cloudProfileConfig,
		infraConfig:        infraConfig,
		region:             region,
	}
	return prepared, nil
}

func (a *actuator) reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {

	return nil
}
