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

// Package controller provides Kubernetes controller configuration structures used for command execution
package controller

import (
	"context"
	"fmt"
	"os"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	hcloudcontrolplane "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/controlplane"
	hcloudhealthcheck "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/healthcheck"
	hcloudinfrastructure "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/infrastructure"
	hcloudworker "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/worker"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	hcloudapisinstall "github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/install"
)

// NewControllerManagerCommand creates a new command for running a HCloud provider controller.
//
// PARAMETERS
// ctx context.Context Execution context
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	generalOpts := &cmd.GeneralOptions{}
	restOpts := &cmd.RESTOptions{}

	mgrOpts := &cmd.ManagerOptions{
		LeaderElection:          true,
		LeaderElectionID:        cmd.LeaderElectionNameID(hcloud.Name),
		LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		WebhookServerPort:       443,
	}

	configFileOpts := &ConfigOptions{}

	// options for the infrastructure controller
	infraCtrlOpts := &cmd.ControllerOptions{
		MaxConcurrentReconciles: 5,
	}
	reconcileOpts := &cmd.ReconcilerOptions{}

	// options for the health care controller
	healthCareCtrlOpts := &cmd.ControllerOptions{
		MaxConcurrentReconciles: 5,
	}

	// options for the heartbeat controller
	heartbeatCtrlOpts := &heartbeatcmd.Options{
		ExtensionName:        hcloud.Name,
		RenewIntervalSeconds: 30,
		Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
	}

	// options for the control plane controller
	controlPlaneCtrlOpts := &cmd.ControllerOptions{
		MaxConcurrentReconciles: 5,
	}

	// options for the worker controller
	workerCtrlOpts := &cmd.ControllerOptions{
		MaxConcurrentReconciles: 5,
	}

	// options for the webhook server
	webhookServerOptions := &webhookcmd.ServerOptions{
		Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
	}

	controllerSwitches := controllerSwitchOptions()
	webhookSwitches := webhookSwitchOptions()

	webhookOptions := webhookcmd.NewAddToManagerOptions(hcloud.Name,
		genericactuator.ShootWebhooksResourceName,
		genericactuator.ShootWebhookNamespaceSelector(hcloud.Type),
		webhookServerOptions,
		webhookSwitches,
	)

	aggOption := cmd.NewOptionAggregator(
		generalOpts,
		restOpts,
		mgrOpts,
		cmd.PrefixOption("controlplane-", controlPlaneCtrlOpts),
		cmd.PrefixOption("infrastructure-", infraCtrlOpts),
		cmd.PrefixOption("worker-", workerCtrlOpts),
		cmd.PrefixOption("healthcheck-", healthCareCtrlOpts),
		cmd.PrefixOption("heartbeat-", heartbeatCtrlOpts),
		controllerSwitches,
		configFileOpts,
		reconcileOpts,
		webhookOptions,
	)

	cmdDefinition := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", hcloud.Name),

		PreRun: func(cmdDefinition *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
		},

		RunE: func(cmdDefinition *cobra.Command, args []string) error {
			if err := aggOption.Complete(); err != nil {
				return fmt.Errorf("Error completing options: %w", err)
			}
			if err := heartbeatCtrlOpts.Validate(); err != nil {
				return err
			}

			util.ApplyClientConnectionConfigurationToRESTConfig(configFileOpts.Completed().Config.ClientConnection, restOpts.Completed().Config)

			mgrOptions := mgrOpts.Completed().Options()

			configFileOpts.Completed().ApplyMetricsBindAddress(&mgrOptions.Metrics.BindAddress)

			mgr, err := manager.New(restOpts.Completed().Config, mgrOptions)
			if err != nil {
				return fmt.Errorf("Could not instantiate manager: %w", err)
			}

			scheme := mgr.GetScheme()
			if err := controller.AddToScheme(scheme); err != nil {
				return fmt.Errorf("Could not update manager scheme: %w", err)
			}
			if err := hcloudapisinstall.AddToScheme(scheme); err != nil {
				return fmt.Errorf("Could not update manager scheme: %w", err)
			}
			if err := druidv1alpha1.AddToScheme(scheme); err != nil {
				return fmt.Errorf("Could not update manager scheme: %w", err)
			}
			if err := machinev1alpha1.AddToScheme(scheme); err != nil {
				return fmt.Errorf("Could not update manager scheme: %w", err)
			}
			if err := autoscalingv1beta2.AddToScheme(scheme); err != nil {
				return fmt.Errorf("Could not update manager scheme: %w", err)
			}

			// add common meta types to schema for controller-runtime to use v1.ListOptions
			metav1.AddToGroupVersion(scheme, machinev1alpha1.SchemeGroupVersion)

			log := mgr.GetLogger()
			gardenCluster, err := getGardenCluster(log)
			log.Info("Adding garden cluster to manager")
			if err := mgr.Add(gardenCluster); err != nil {
				return fmt.Errorf("failed adding garden cluster to manager: %w", err)
			}
			if err != nil {
				return err
			}
			log.Info("Adding controllers to manager")

			configFileOpts.Completed().ApplyGardenId(&hcloudcontrolplane.DefaultAddOptions.GardenId)
			configFileOpts.Completed().ApplyGardenId(&hcloudinfrastructure.DefaultAddOptions.GardenId)
			configFileOpts.Completed().ApplyHealthCheckConfig(&hcloudhealthcheck.DefaultAddOptions.HealthCheckConfig)
			healthCareCtrlOpts.Completed().Apply(&hcloudhealthcheck.DefaultAddOptions.Controller)
			heartbeatCtrlOpts.Completed().Apply(&heartbeat.DefaultAddOptions)
			controlPlaneCtrlOpts.Completed().Apply(&hcloudcontrolplane.DefaultAddOptions.Controller)
			infraCtrlOpts.Completed().Apply(&hcloudinfrastructure.DefaultAddOptions.Controller)
			reconcileOpts.Completed().Apply(&hcloudinfrastructure.DefaultAddOptions.IgnoreOperationAnnotation)
			reconcileOpts.Completed().Apply(&hcloudcontrolplane.DefaultAddOptions.IgnoreOperationAnnotation)
			reconcileOpts.Completed().Apply(&hcloudworker.DefaultAddOptions.IgnoreOperationAnnotation)
			workerCtrlOpts.Completed().Apply(&hcloudworker.DefaultAddOptions.Controller)

			hcloudworker.DefaultAddOptions.GardenCluster = gardenCluster

			if _, err := webhookOptions.Completed().AddToManager(ctx, mgr, nil); err != nil {
				return fmt.Errorf("Could not add webhooks to manager: %w", err)
			}

			hcloudcontrolplane.DefaultAddOptions.WebhookServerNamespace = webhookOptions.Server.Namespace

			if err := controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("Could not add controllers to manager: %w", err)
			}

			if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
				return err
			}

			if err := mgr.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
				return fmt.Errorf("could not add readycheck for informers: %w", err)
			}

			if err := mgr.AddReadyzCheck("webhook-server", mgr.GetWebhookServer().StartedChecker()); err != nil {
				return fmt.Errorf("could not add readycheck of webhook to manager: %w", err)
			}

			if err := mgr.Start(ctx); err != nil {
				return fmt.Errorf("Error running manager: %w", err)
			}

			return nil
		},
	}

	cmdFlags := cmdDefinition.Flags()
	aggOption.AddFlags(cmdFlags)
	verflag.AddFlags(cmdFlags)

	return cmdDefinition
}

func getGardenCluster(log logr.Logger) (cluster.Cluster, error) {
	log.Info("Getting rest config for garden")
	gardenRESTConfig, err := kubernetes.RESTConfigFromKubeconfigFile(os.Getenv("GARDEN_KUBECONFIG"), kubernetes.AuthTokenFile)
	if err != nil {
		return nil, err
	}

	log.Info("Setting up cluster object for garden")
	gardenCluster, err := cluster.New(gardenRESTConfig, func(opts *cluster.Options) {
		opts.Scheme = kubernetes.GardenScheme
		opts.Logger = log
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating garden cluster object: %w", err)
	}

	return gardenCluster, nil
}
