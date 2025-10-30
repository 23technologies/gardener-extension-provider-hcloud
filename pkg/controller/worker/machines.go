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
	"path/filepath"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	mcmv1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/23technologies/gardener-extension-provider-hcloud/charts"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
)

// MachineClassKind yields the name of the machine class.
func (w *workerDelegate) MachineClassKind() string {
	return "MachineClass"
}

// MachineClass yields a newly initialized MachineClass object.
func (w *workerDelegate) MachineClass() client.Object {
	return &mcmv1alpha1.MachineClass{}
}

// MachineClassList yields a newly initialized MachineClassList object.
func (w *workerDelegate) MachineClassList() client.ObjectList {
	return &mcmv1alpha1.MachineClassList{}
}

// DeployMachineClasses generates and creates the HCloud specific machine classes.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	return w.seedChartApplier.ApplyFromEmbeddedFS(ctx, charts.InternalChart, filepath.Join(charts.InternalChartsPath, "machineclass"), w.worker.Namespace, "machineclass", kubernetes.Values(map[string]interface{}{"machineClasses": w.machineClasses}))
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

// getSecretData returns the secret referenced by the WorkerDelegate instance's spec.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) getSecretData(ctx context.Context) (*corev1.Secret, error) {
	return extensionscontroller.GetSecretByReference(ctx, w.client, &w.worker.Spec.SecretRef)
}

// generateMachineClassSecretData returns the machine class relevant secret values.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) generateMachineClassSecretData(ctx context.Context) (map[string][]byte, error) {
	secret, err := w.getSecretData(ctx)
	if err != nil {
		return nil, err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return nil, err
	}

	return map[string][]byte{
		hcloud.HcloudToken: []byte(credentials.MCM().Token),
	}, nil
}

// generateMachineConfig generates the machine config of the WorkerDelegate instance's spec.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineDeployments = worker.MachineDeployments{}
		machineClasses     []map[string]interface{}
		// machineImages      []apis.MachineImage
	)

	machineClassSecretData, err := w.generateMachineClassSecretData(ctx)
	if err != nil {
		return err
	}

	infraStatus, err := transcoder.DecodeInfrastructureStatusFromWorker(w.worker)
	if err != nil {
		return err
	}

	workerStatus, err := transcoder.DecodeWorkerStatusFromWorker(w.worker)
	if err != nil {
		return err
	}

	sshFingerprint := infraStatus.SSHFingerprint

	if "" == sshFingerprint {
		sshFingerprint, err = apis.GetSSHFingerprint(w.worker.Spec.SSHPublicKey)
		if err != nil {
			return err
		}
	}

	if len(w.worker.Spec.Pools) == 0 {
		return fmt.Errorf("missing pool")
	}

	for _, pool := range w.worker.Spec.Pools {
		zoneLen := int32(len(pool.Zones)) // #nosec: G115 - We validate if num pool zones exceeds max_int32.

		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster, nil, nil, nil)
		if err != nil {
			return err
		}

		imageName, err := w.findMachineImageName(ctx, pool.MachineImage.Name, pool.MachineImage.Version)
		if err != nil {
			return err
		}

		values, err := w.extractMachineValues(pool.MachineType)
		if err != nil {
			return fmt.Errorf("extracting machine values failed: %w", err)
		}

		userData, err := worker.FetchUserData(ctx, w.client, w.worker.Namespace, pool)
		if err != nil {
			return err
		}

		for zoneIndex, zone := range pool.Zones {
			zoneIdx := int32(zoneIndex) // #nosec: G115 - We validate if num pool zones exceeds max_int32.
			secretMap := map[string]interface{}{
				"userData": string(userData),
			}

			for key, value := range machineClassSecretData {
				secretMap[key] = value
			}

			machineClassSpec := map[string]interface{}{
				"cluster":        w.worker.Namespace,
				"zone":           zone,
				"imageName":      string(imageName),
				"sshFingerprint": sshFingerprint,
				"machineType":    string(pool.MachineType),
				"networkName":    fmt.Sprintf("%s-workers", w.worker.Namespace),
				"tags": map[string]string{
					"mcm.gardener.cloud/cluster": w.worker.Namespace,
					"mcm.gardener.cloud/role":    "node",
				},
				"credentialsSecretRef": map[string]interface{}{
					"name":      w.worker.Spec.SecretRef.Name,
					"namespace": w.worker.Spec.SecretRef.Namespace,
				},
				"secret": secretMap,
			}

			placementGroupName := fmt.Sprintf("%s-%s", w.worker.Namespace, pool.Name)
			if placementGroupID, ok := workerStatus.PlacementGroupIDs[placementGroupName]; ok {
				machineClassSpec["placementGroupID"] = placementGroupID
			}

			if "" != infraStatus.FloatingPoolName {
				machineClassSpec["floatingPoolName"] = infraStatus.FloatingPoolName
			}

			if values.MachineTypeOptions != nil {
				if len(values.MachineTypeOptions.ExtraConfig) > 0 {
					machineClassSpec["extraConfig"] = values.MachineTypeOptions.ExtraConfig
				}
			}

			deploymentName := fmt.Sprintf("%s-%s-%s", w.worker.Namespace, pool.Name, zone)
			className := fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)

			updateConfiguration := machinev1alpha1.UpdateConfiguration{
				MaxUnavailable: ptr.To(worker.DistributePositiveIntOrPercent(zoneIdx, pool.MaxUnavailable, zoneLen, pool.Minimum)),
				MaxSurge:       ptr.To(worker.DistributePositiveIntOrPercent(zoneIdx, pool.MaxSurge, zoneLen, pool.Maximum)),
			}

			machineDeploymentStrategy := machinev1alpha1.MachineDeploymentStrategy{
				Type: machinev1alpha1.RollingUpdateMachineDeploymentStrategyType,
				RollingUpdate: &machinev1alpha1.RollingUpdateMachineDeployment{
					UpdateConfiguration: updateConfiguration,
				},
			}

			if gardencorev1beta1helper.IsUpdateStrategyInPlace(pool.UpdateStrategy) {
				machineDeploymentStrategy = machinev1alpha1.MachineDeploymentStrategy{
					Type: machinev1alpha1.InPlaceUpdateMachineDeploymentStrategyType,
					InPlaceUpdate: &machinev1alpha1.InPlaceUpdateMachineDeployment{
						UpdateConfiguration: updateConfiguration,
						OrchestrationType:   machinev1alpha1.OrchestrationTypeAuto,
					},
				}

				if gardencorev1beta1helper.IsUpdateStrategyManualInPlace(pool.UpdateStrategy) {
					machineDeploymentStrategy.InPlaceUpdate.OrchestrationType = machinev1alpha1.OrchestrationTypeManual
				}
			}

			machineDeployments = append(machineDeployments, worker.MachineDeployment{
				Name:                 deploymentName,
				ClassName:            className,
				SecretName:           className,
				Minimum:              pool.Minimum,
				Maximum:              pool.Maximum,
				Strategy:             machineDeploymentStrategy,
				Labels:               pool.Labels,
				Annotations:          pool.Annotations,
				Taints:               pool.Taints,
				MachineConfiguration: genericworkeractuator.ReadMachineConfiguration(pool),
			})

			machineClassSpec["name"] = className

			machineClasses = append(machineClasses, machineClassSpec)
		}

	}
	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses

	return nil
}

type machineValues struct {
	MachineTypeOptions *apis.MachineTypeOptions
}

// extractMachineValues extracts the relevant machine values from the cloud profile spec.
//
// PARAMETERS
// ctx context.Context Execution context
func (w *workerDelegate) extractMachineValues(machineTypeName string) (*machineValues, error) {
	var machineType *corev1beta1.MachineType
	for _, mt := range w.cluster.CloudProfile.Spec.MachineTypes {
		if mt.Name == machineTypeName {
			machineType = &mt
			break
		}
	}
	if machineType == nil {
		err := fmt.Errorf("machine type %s not found in cloud profile spec", machineTypeName)
		return nil, err
	}

	values := &machineValues{}

	cloudProfileConfig, err := transcoder.DecodeConfigFromCloudProfile(w.cluster.CloudProfile)
	if err != nil {
		return nil, err
	}

	for _, mt := range cloudProfileConfig.MachineTypeOptions {
		if mt.Name == machineTypeName {
			values.MachineTypeOptions = &mt
			break
		}
	}

	return values, nil
}
