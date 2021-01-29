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

package app

import (
	"context"
	"fmt"
	"os"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/util"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"

	hcloudcmd "github.com/23technologies/gardener-extension-provider-hcloud/pkg/cmd"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"

	// hcloudcontrolplane "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/controlplane"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/healthcheck"
	hcloudinfrastructure "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/infrastructure"

	hcloudworker "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/worker"
	// hcloudcontrolplaneexposure "github.com/23technologies/gardener-extension-provider-hcloud/pkg/webhook/controlplaneexposure"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewControllerManagerCommand creates a new command for running a HCloud provider controller.
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		restOpts = &controllercmd.RESTOptions{}
		mgrOpts  = &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(hcloud.Name),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			WebhookServerPort:       443,
		}
		configFileOpts = &hcloudcmd.ConfigOptions{}

		// options for the infrastructure controller
		infraCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		reconcileOpts = &controllercmd.ReconcilerOptions{}

		// options for the health care controller
		healthCareCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the control plane controller
		controlPlaneCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the worker controller
		workerCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		workerReconcileOpts = &worker.Options{
			DeployCRDs: true,
		}
		workerCtrlOptsUnprefixed = controllercmd.NewOptionAggregator(workerCtrlOpts, workerReconcileOpts)

		// options for the webhook server
		webhookServerOptions = &webhookcmd.ServerOptions{
			Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
		}

		controllerSwitches = hcloudcmd.ControllerSwitchOptions()
		webhookSwitches    = hcloudcmd.WebhookSwitchOptions()
		webhookOptions     = webhookcmd.NewAddToManagerOptions(hcloud.Name, webhookServerOptions, webhookSwitches)

		aggOption = controllercmd.NewOptionAggregator(
			restOpts,
			mgrOpts,
			controllercmd.PrefixOption("controlplane-", controlPlaneCtrlOpts),
			controllercmd.PrefixOption("infrastructure-", infraCtrlOpts),
			controllercmd.PrefixOption("worker-", &workerCtrlOptsUnprefixed),
			controllercmd.PrefixOption("healthcheck-", healthCareCtrlOpts),
			controllerSwitches,
			configFileOpts,
			reconcileOpts,
			webhookOptions,
		)
	)

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", hcloud.Name),

		Run: func(cmd *cobra.Command, args []string) {
			if err := aggOption.Complete(); err != nil {
				controllercmd.LogErrAndExit(err, "Error completing options")
			}

			util.ApplyClientConnectionConfigurationToRESTConfig(configFileOpts.Completed().Config.ClientConnection, restOpts.Completed().Config)

			if workerReconcileOpts.Completed().DeployCRDs {
				if err := worker.ApplyMachineResourcesForConfig(ctx, restOpts.Completed().Config); err != nil {
					controllercmd.LogErrAndExit(err, "Error ensuring the machine CRDs")
				}
			}

			mgr, err := manager.New(restOpts.Completed().Config, mgrOpts.Completed().Options())
			if err != nil {
				controllercmd.LogErrAndExit(err, "Could not instantiate manager")
			}

			scheme := mgr.GetScheme()
			if err := controller.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			// if err := hcloudinstall.AddToScheme(scheme); err != nil {
			// controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			// }
			if err := druidv1alpha1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := machinev1alpha1.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}
			if err := autoscalingv1beta2.AddToScheme(scheme); err != nil {
				controllercmd.LogErrAndExit(err, "Could not update manager scheme")
			}

			// add common meta types to schema for controller-runtime to use v1.ListOptions
			metav1.AddToGroupVersion(scheme, machinev1alpha1.SchemeGroupVersion)

			// configFileOpts.Completed().ApplyETCDStorage(&hcloudcontrolplaneexposure.DefaultAddOptions.ETCDStorage)
			// configFileOpts.Completed().ApplyGardenId(&hcloudcontrolplane.DefaultAddOptions.GardenId)
			configFileOpts.Completed().ApplyGardenId(&hcloudinfrastructure.DefaultAddOptions.GardenId)
			configFileOpts.Completed().ApplyHealthCheckConfig(&healthcheck.DefaultAddOptions.HealthCheckConfig)
			healthCareCtrlOpts.Completed().Apply(&healthcheck.DefaultAddOptions.Controller)
			// controlPlaneCtrlOpts.Completed().Apply(&hcloudcontrolplane.DefaultAddOptions.Controller)
			infraCtrlOpts.Completed().Apply(&hcloudinfrastructure.DefaultAddOptions.Controller)
			reconcileOpts.Completed().Apply(&hcloudinfrastructure.DefaultAddOptions.IgnoreOperationAnnotation)
			// reconcileOpts.Completed().Apply(&hcloudcontrolplane.DefaultAddOptions.IgnoreOperationAnnotation)
			reconcileOpts.Completed().Apply(&hcloudworker.DefaultAddOptions.IgnoreOperationAnnotation)
			workerCtrlOpts.Completed().Apply(&hcloudworker.DefaultAddOptions.Controller)

			if _, _, err := webhookOptions.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add webhooks to manager")
			}

			if err := controllerSwitches.Completed().AddToManager(mgr); err != nil {
				controllercmd.LogErrAndExit(err, "Could not add controllers to manager")
			}

			if err := mgr.Start(ctx.Done()); err != nil {
				controllercmd.LogErrAndExit(err, "Error running manager")
			}
		},
	}

	aggOption.AddFlags(cmd.Flags())

	return cmd
}
