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

// Package controlplane contains functions used at the controlplane controller
package controlplane

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener/extensions/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	controllerapis "github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/controller"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}

	logger = log.Log.WithName("hcloud-controlplane-controller")
)

// AddOptions are options to apply when adding the HCloud controlplane controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
	// GardenId is the Gardener garden identity
	GardenId string
	// WebhookServerNamespace is the namespace in which the webhook server runs.
	WebhookServerNamespace string
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
//
// PARAMETERS
// mgr  manager.Manager Control plane controller manager instance
// opts AddOptions      Options to add
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	genericActuator, err := genericactuator.NewActuator(
		mgr,
		hcloud.Name,

		getSecretConfigs,
		getAccessSecrets,

		nil,
		nil,

		configChart,
		controlPlaneChart,
		controlPlaneShootChart,
		nil,
		storageClassChart,
		nil,
		NewValuesProvider(mgr, logger, opts.GardenId),
		extensionscontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot),
		controllerapis.ImageVector(),
		hcloud.CloudProviderConfig,
		nil,
		opts.WebhookServerNamespace,
	)

	if err != nil {
		return err
	}

	return controlplane.Add(ctx, mgr, controlplane.AddArgs{
		Actuator:          genericActuator,
		ControllerOptions: opts.Controller,
		Predicates:        controlplane.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              hcloud.Type,
	})
}

// AddToManager adds a controller with the default Options.
//
// PARAMETERS
// mgr manager.Manager Control plane controller manager instance
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
