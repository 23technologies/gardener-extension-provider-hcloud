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
	"fmt"
	"github.com/23technologies/gardener-extension-provider-hcloud/charts"
	"path/filepath"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/chart"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var (
	mcmChart = &chart.Chart{
		Name:       hcloud.MachineControllerManagerName,
		EmbeddedFS: charts.InternalChart,
		Path:       filepath.Join(charts.InternalChartsPath, hcloud.MachineControllerManagerName, "seed"),
		Images:     []string{hcloud.MachineControllerManagerImageName, hcloud.MCMProviderHcloudImageName},
		Objects: []*chart.Object{
			{Type: &appsv1.Deployment{}, Name: hcloud.MachineControllerManagerName},
			{Type: &corev1.Service{}, Name: hcloud.MachineControllerManagerName},
			{Type: &corev1.ServiceAccount{}, Name: hcloud.MachineControllerManagerName},
			{Type: &corev1.Secret{}, Name: hcloud.MachineControllerManagerName},
			{Type: extensionscontroller.GetVerticalPodAutoscalerObject(), Name: hcloud.MachineControllerManagerVpaName},
			{Type: &corev1.ConfigMap{}, Name: hcloud.MachineControllerManagerMonitoringConfigName},
		},
	}

	mcmShootChart = &chart.Chart{
		Name:       hcloud.MachineControllerManagerName,
		EmbeddedFS: charts.InternalChart,
		Path:       filepath.Join(charts.InternalChartsPath, hcloud.MachineControllerManagerName, "shoot"),
		Objects: []*chart.Object{
			{Type: &rbacv1.ClusterRole{}, Name: fmt.Sprintf("extensions.gardener.cloud:%s:%s", hcloud.Name, hcloud.MachineControllerManagerName)},
			{Type: &rbacv1.ClusterRoleBinding{}, Name: fmt.Sprintf("extensions.gardener.cloud:%s:%s", hcloud.Name, hcloud.MachineControllerManagerName)},
		},
	}
)

// GetMachineControllerManagerChartValues returns chart values relevant for the MCM instance.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) GetMachineControllerManagerChartValues(ctx context.Context) (map[string]interface{}, error) {
	namespace := &corev1.Namespace{}
	if err := w.client.Get(ctx, kutil.Key(w.worker.Namespace), namespace); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"providerName": hcloud.Name,
		"namespace": map[string]interface{}{
			"uid": namespace.UID,
		},
		"podLabels": map[string]interface{}{
			v1beta1constants.LabelPodMaintenanceRestart: "true",
		},
	}, nil
}

// GetMachineControllerManagerShootChartValues returns chart values relevant for the MCM shoot instance.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) GetMachineControllerManagerShootChartValues(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"providerName": hcloud.Name,
	}, nil
}

// GetMachineControllerManagerCloudCredentials should return the IaaS credentials
// with the secret keys used by the machine-controller-manager.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) GetMachineControllerManagerCloudCredentials(ctx context.Context) (map[string][]byte, error) {
	return w.generateMachineClassSecretData(ctx)
}
