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

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardener "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/go-logr/logr"
	hcloudclient "github.com/hetznercloud/hcloud-go/v2/hcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/controller"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/v1alpha1"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
)

type delegateFactory struct {
	logger       logr.Logger
	seedClient   client.Client
	restConfig   *rest.Config
	scheme       *runtime.Scheme
	gardenReader client.Reader
}

// NewActuator creates a new Actuator that updates the status of the handled WorkerPoolConfigs.
func NewActuator(mgr manager.Manager, gardenCluster cluster.Cluster) (worker.Actuator, error) {
	delegateFactory := &delegateFactory{
		logger:     log.Log.WithName("worker-actuator"),
		seedClient: mgr.GetClient(),
		restConfig: mgr.GetConfig(),
		scheme:     mgr.GetScheme(),
	}

	return genericactuator.NewActuator(
		mgr,
		gardenCluster,
		delegateFactory,
		nil), nil
}

// WorkerDelegate returns the WorkerDelegate instance for the given worker and cluster struct.
//
// PARAMETERS
// ctx     context.Context               Execution context
// worker  *extensionsv1alpha1.Worker    Worker struct
// cluster *extensionscontroller.Cluster Cluster struct
func (d *delegateFactory) WorkerDelegate(ctx context.Context, worker *extensionsv1alpha1.Worker, cluster *extensionscontroller.Cluster) (genericactuator.WorkerDelegate, error) {
	clientset, err := kubernetes.NewForConfig(d.restConfig)
	if err != nil {
		return nil, err
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	seedChartApplier, err := gardener.NewChartApplierForConfig(d.restConfig)
	if err != nil {
		return nil, err
	}

	return NewWorkerDelegate(
		d.seedClient,
		d.scheme,
		seedChartApplier,
		serverVersion.GitVersion,

		worker,
		cluster,
	)
}

type workerDelegate struct {
	client  client.Client
	decoder runtime.Decoder
	scheme  *runtime.Scheme

	seedChartApplier gardener.ChartApplier
	serverVersion    string

	cloudProfileConfig *api.CloudProfileConfig
	cluster            *extensionscontroller.Cluster
	worker             *extensionsv1alpha1.Worker

	machineClasses     []map[string]interface{}
	machineDeployments worker.MachineDeployments
	machineImages      []api.MachineImage

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
	client client.Client,
	scheme *runtime.Scheme,

	seedChartApplier gardener.ChartApplier,
	serverVersion string,

	worker *extensionsv1alpha1.Worker,
	cluster *extensionscontroller.Cluster,
) (genericactuator.WorkerDelegate, error) {
	cloudProfileConfig, err := controller.GetCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	secret, err := extensionscontroller.GetSecretByReference(context.Background(), client, &worker.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return nil, err
	}

	token := credentials.CCM().Token
	hclient := api.GetClientForToken(string(token))

	return &workerDelegate{
		client:  client,
		scheme:  scheme,
		decoder: serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDecoder(),

		seedChartApplier: seedChartApplier,
		serverVersion:    serverVersion,

		cloudProfileConfig: cloudProfileConfig,
		cluster:            cluster,
		worker:             worker,
		hclient:            hclient,
	}, nil
}

// updateProviderStatus updates the worker provider status.
//
// PARAMETERS
// ctx         context.Context     Execution context
// workerStatus *apis.WorkerStatus Worker status to be applied
func (w *workerDelegate) updateProviderStatus(ctx context.Context, workerStatus *api.WorkerStatus) error {
	var workerStatusV1alpha1 = &v1alpha1.WorkerStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "WorkerStatus",
		},
	}

	err := w.scheme.Convert(workerStatus, workerStatusV1alpha1, nil)
	if nil != err {
		return err
	}

	patch := client.MergeFrom(w.worker.DeepCopy())
	w.worker.Status.ProviderStatus = &runtime.RawExtension{Object: workerStatusV1alpha1}
	return w.client.Status().Patch(ctx, w.worker, patch)
}
