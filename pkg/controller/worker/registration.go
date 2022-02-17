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
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	machinescheme "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/scheme"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	// UseTokenRequestor specifies whether the token requestor shall be used for the control plane components.
	UseTokenRequestor bool
	// UseProjectedTokenMount specifies whether the projected token mount shall be used for the
	// control plane components.
	UseProjectedTokenMount bool
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
//
// PARAMETERS
// mgr  manager.Manager Worker controller manager instance
// opts AddOptions      Options to add
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) error {
	scheme := mgr.GetScheme()
	if err := apiextensionsscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := machinescheme.AddToScheme(scheme); err != nil {
		return err
	}

	return worker.Add(mgr, worker.AddArgs{
		Actuator:          NewActuator(opts.UseTokenRequestor, opts.UseProjectedTokenMount),
		ControllerOptions: opts.Controller,
		Predicates:        worker.DefaultPredicates(opts.IgnoreOperationAnnotation),
		Type:              hcloud.Type,
	})
}

// AddToManager adds a controller with the default Options.
//
// PARAMETERS
// mgr manager.Manager Worker controller manager instance
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}
