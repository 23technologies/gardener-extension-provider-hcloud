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

// Package healthcheck contains functions used for cluster validation
package healthcheck

import (
	"context"
	"time"

	extensionconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	genericcontrolplaneactuator "github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/worker"
	extensionspredicate "github.com/gardener/gardener/extensions/pkg/predicate"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
)

var (
	defaultSyncPeriod = time.Second * 30
	// DefaultAddOptions are the default DefaultAddArgs for AddToManager.
	DefaultAddOptions = healthcheck.DefaultAddArgs{
		HealthCheckConfig: extensionconfig.HealthCheckConfig{SyncPeriod: metav1.Duration{Duration: defaultSyncPeriod}},
	}
)

// RegisterHealthChecks registers health checks for each extension resource
// HealthChecks are grouped by extension (e.g worker), extension.type (e.g hcloud) and  Health Check Type (e.g SystemComponentsHealthy)
//
// PARAMETERS
// mgr  manager.Manager            Health check controller manager instance
// opts healthcheck.DefaultAddArgs Options to add
func RegisterHealthChecks(ctx context.Context, mgr manager.Manager, opts healthcheck.DefaultAddArgs) error {
	err := healthcheck.DefaultRegistration(
		ctx,
		hcloud.Type,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.ControlPlaneResource),
		func() client.ObjectList { return &extensionsv1alpha1.ControlPlaneList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.ControlPlane{} },
		mgr,
		opts,
		[]predicate.Predicate{extensionspredicate.HasPurpose(extensionsv1alpha1.Normal)},
		[]healthcheck.ConditionTypeToHealthCheck{
			{
				ConditionType: string(gardencorev1beta1.ShootControlPlaneHealthy),
				HealthCheck:   general.NewSeedDeploymentHealthChecker(hcloud.CloudControllerManagerName),
			},
			{
				ConditionType: string(gardencorev1beta1.ShootSystemComponentsHealthy),
				HealthCheck:   general.CheckManagedResource(genericcontrolplaneactuator.ControlPlaneShootChartResourceName),
			},
			{
				ConditionType: string(gardencorev1beta1.ShootSystemComponentsHealthy),
				HealthCheck:   general.CheckManagedResource(genericcontrolplaneactuator.StorageClassesChartResourceName),
			},
		},
		sets.Set[gardencorev1beta1.ConditionType]{},
	)
	if err != nil {
		return err
	}

	return healthcheck.DefaultRegistration(
		ctx,
		hcloud.Type,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.WorkerResource),
		func() client.ObjectList { return &extensionsv1alpha1.WorkerList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.Worker{} },
		mgr,
		opts,
		nil,
		[]healthcheck.ConditionTypeToHealthCheck{{
			ConditionType: string(gardencorev1beta1.ShootEveryNodeReady),
			HealthCheck:   worker.NewNodesChecker(),
		}},
		sets.New(gardencorev1beta1.ShootControlPlaneHealthy),
	)
}

// AddToManager adds a controller with the default Options.
//
// PARAMETERS
// mgr manager.Manager Health check controller manager instance
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return RegisterHealthChecks(ctx, mgr, DefaultAddOptions)
}
