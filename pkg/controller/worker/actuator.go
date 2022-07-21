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

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/controller"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	"github.com/gardener/gardener/extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardener "github.com/gardener/gardener/pkg/client/kubernetes"
	hcloudclient "github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type delegateFactory struct {
	logger logr.Logger
	common.RESTConfigContext
}

// NewActuator creates a new Actuator that updates the status of the handled WorkerPoolConfigs.
func NewActuator() worker.Actuator {
	delegateFactory := &delegateFactory{
		logger: log.Log.WithName("worker-actuator"),
	}

	return genericactuator.NewActuator(
		log.Log.WithName("hcloud-worker-actuator"),
		delegateFactory,
		hcloud.MachineControllerManagerName,
		mcmChart,
		mcmShootChart,
		controller.ImageVector(),
		extensionscontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot),
	)
}

// WorkerDelegate returns the WorkerDelegate instance for the given worker and cluster struct.
//
// PARAMETERS
// ctx     context.Context               Execution context
// worker  *extensionsv1alpha1.Worker    Worker struct
// cluster *extensionscontroller.Cluster Cluster struct
func (d *delegateFactory) WorkerDelegate(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) (genericactuator.WorkerDelegate, error) {
	clientset, err := kubernetes.NewForConfig(d.RESTConfig())
	if err != nil {
		return nil, err
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	seedChartApplier, err := gardener.NewChartApplierForConfig(d.RESTConfig())
	if err != nil {
		return nil, err
	}

	return NewWorkerDelegate(
		d.ClientContext,
		seedChartApplier,
		serverVersion.GitVersion,

		worker,
		cluster,
	)
}

type workerDelegate struct {
	common.ClientContext

	seedChartApplier gardener.ChartApplier
	serverVersion    string

	cloudProfileConfig *apis.CloudProfileConfig
	cluster            *extensionscontroller.Cluster
	worker             *extensionsv1alpha1.Worker

	machineClasses     []map[string]interface{}
	machineDeployments worker.MachineDeployments
	machineImages      []apis.MachineImage

	hclient *hcloudclient.Client
}

// NewWorkerDelegate creates a new context for a worker reconciliation.
//
// PARAMETERS
// clientContext    common.ClientContext          Client context
// seedChartApplier gardener.ChartApplier         Chart applier instance
// serverVersion    string                        Kubernetes version
// worker           *extensionsv1alpha1.Worker    Worker struct
// cluster          *extensionscontroller.Cluster Cluster struct
func NewWorkerDelegate(
	clientContext common.ClientContext,

	seedChartApplier gardener.ChartApplier,
	serverVersion string,

	worker *extensionsv1alpha1.Worker,
	cluster *extensionscontroller.Cluster,
) (genericactuator.WorkerDelegate, error) {
	cloudProfileConfig, err := controller.GetCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	secret, err := extensionscontroller.GetSecretByReference(context.Background(), clientContext.Client(), &worker.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return nil, err
	}

	token := credentials.CCM().Token
	client := apis.GetClientForToken(string(token))

	return &workerDelegate{
		ClientContext: clientContext,

		seedChartApplier: seedChartApplier,
		serverVersion:    serverVersion,

		cloudProfileConfig: cloudProfileConfig,
		cluster:            cluster,
		worker:             worker,
		hclient:            client,
	}, nil
}

// updateProviderStatus updates the worker provider status.
//
// PARAMETERS
// ctx         context.Context     Execution context
// workerStatus *apis.WorkerStatus Worker status to be applied
func (w *workerDelegate) updateProviderStatus(ctx context.Context, workerStatus *apis.WorkerStatus) error {
	var workerStatusV1alpha1 = &v1alpha1.WorkerStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "WorkerStatus",
		},
	}

	err := w.Scheme().Convert(workerStatus, workerStatusV1alpha1, nil)
	if nil != err {
		return err
	}

	patch := client.MergeFrom(w.worker.DeepCopy())
	w.worker.Status.ProviderStatus = &runtime.RawExtension{Object: workerStatusV1alpha1}
	return w.Client().Status().Patch(ctx, w.worker, patch)
}
