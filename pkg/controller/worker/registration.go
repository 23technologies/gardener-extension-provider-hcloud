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

// Package worker contains functions used at the worker controller
package worker

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	machinescheme "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/scheme"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the HCloud worker controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
	// GardenCluster is the garden cluster object.
	GardenCluster  cluster.Cluster
	ExtensionClass extensionsv1alpha1.ExtensionClass
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
//
// PARAMETERS
// mgr  manager.Manager Worker controller manager instance
// opts AddOptions      Options to add
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	schemeBuilder := runtime.NewSchemeBuilder(
		apiextensionsscheme.AddToScheme,
		machinescheme.AddToScheme,
	)

	if err := schemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}

	actuator, err := NewActuator(mgr, opts.GardenCluster)
	if err != nil {
		return err
	}
	return worker.Add(ctx, mgr, worker.AddArgs{
		Actuator:          actuator,
		ControllerOptions: opts.Controller,
		Predicates:        worker.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              hcloud.Type,
		ExtensionClass:    opts.ExtensionClass,
	})
}

// AddToManager adds a controller with the default Options.
//
// PARAMETERS
// mgr manager.Manager Worker controller manager instance
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
