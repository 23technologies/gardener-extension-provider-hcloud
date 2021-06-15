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

// Package infrastructure contains functions used at the infrastructure controller
package infrastructure

import (
	"context"
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type actuator struct {
	common.ChartRendererContext

	logger   logr.Logger
	gardenID string
}

type actuatorConfig struct {
	cloudProfileConfig *apis.CloudProfileConfig
	infraConfig        *apis.InfrastructureConfig
	region             *apis.RegionSpec
	token              string
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(gardenID string) infrastructure.Actuator {
	return &actuator{
		logger:   log.Log.WithName("infrastructure-actuator"),
		gardenID: gardenID,
	}
}

func (a *actuator) getActuatorConfig(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) (*actuatorConfig, error) {
	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	infraConfig, err := transcoder.DecodeInfrastructureConfigFromInfrastructure(infra)
	if err != nil {
		return nil, err
	}

	regionSpec := apis.FindRegionSpecForGardenerRegion(infra.Spec.Region, cloudProfileConfig)
	if regionSpec == nil {
		return nil, fmt.Errorf("Region %q not found in cloud profile", infra.Spec.Region)
	}

	secret, err := controller.GetSecretByReference(ctx, a.Client(), &infra.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return nil, err
	}

	token := credentials.HcloudCCM().Token

	config := &actuatorConfig{
		cloudProfileConfig: cloudProfileConfig,
		infraConfig:        infraConfig,
		region:             regionSpec,
		token:              token,
	}

	return config, nil
}

// Restore implements infrastructure.Actuator.Delete
func (a *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return a.delete(ctx, config, cluster)
}

// Restore implements infrastructure.Actuator.Migrate
func (a *actuator) Migrate(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return nil
}

// Restore implements infrastructure.Actuator.Reconcile
func (a *actuator) Reconcile(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return a.reconcile(ctx, config, cluster)
}

// Restore implements infrastructure.Actuator.Restore
func (a *actuator) Restore(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return nil
}

func (a *actuator) updateProviderStatus(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, infraStatus *v1alpha1.InfrastructureStatus) error {
	if nil == infraStatus {
		return nil
	}

	return controller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client(), infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{
			Object: infraStatus,
		}

		return nil
	})
}
