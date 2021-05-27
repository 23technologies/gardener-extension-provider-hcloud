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
	"strings"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	mcmv1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}
	return w.seedChartApplier.Apply(ctx, filepath.Join(hcloud.InternalChartsPath, "machineclass"), w.worker.Namespace, "machineclass", kubernetes.Values(map[string]interface{}{"machineClasses": w.machineClasses}))
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

func (w *workerDelegate) generateMachineClassSecretData(ctx context.Context) (map[string][]byte, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, w.Client(), &w.worker.Spec.SecretRef)
	if err != nil {
		return nil, err
	}

	credentials, err := hcloud.ExtractCredentials(secret)
	if err != nil {
		return nil, err
	}

	region := apis.FindRegion(w.cluster.Shoot.Spec.Region, w.cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found", w.cluster.Shoot.Spec.Region)
	}

	return map[string][]byte{
		"hcloud.Token": []byte(credentials.HcloudMCM().Token),
	}, nil
}

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

	sshFingerprint, err := transcoder.DecodeSSHFingerprintFromPublicKey(w.worker.Spec.SSHPublicKey)
	if err != nil {
		return err
	}

	if len(w.worker.Spec.Pools) == 0 {
		return fmt.Errorf("missing pool")
	}

	for _, pool := range w.worker.Spec.Pools {
		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		values, err := w.extractMachineValues(pool.MachineType)
		if err != nil {
			return errors.Wrap(err, "extracting machine values failed")
		}

		machineClassSpec := map[string]interface{}{
			"cluster":        w.worker.Namespace,
			"datacenter":     string(w.worker.Spec.Region),
			"imageName":      strings.Join([]string{pool.MachineImage.Name, pool.MachineImage.Version}, "-"), //FIXME
			"sshFingerprint": sshFingerprint,
			"machineType":    string(pool.MachineType),
			"networkName":    fmt.Sprintf("%s-workers", w.worker.Namespace),
			"tags": map[string]string{
				"mcm.gardener.cloud/cluster": w.worker.Namespace,
				"mcm.gardener.cloud/role":    "node",
			},
			"secret": map[string]interface{}{
				"userData": string(pool.UserData),
			},
		}

		if "" != infraStatus.FloatingPoolName {
			machineClassSpec["floatingPoolName"] = infraStatus.FloatingPoolName
		}

		if values.MachineTypeOptions != nil {
			if len(values.MachineTypeOptions.ExtraConfig) > 0 {
				machineClassSpec["extraConfig"] = values.MachineTypeOptions.ExtraConfig
			}
		}

		var (
			deploymentName = fmt.Sprintf("%s-%s", w.worker.Namespace, pool.Name)
			className      = fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)
		)

		machineDeployments = append(machineDeployments, worker.MachineDeployment{
			Name:                 deploymentName,
			ClassName:            className,
			SecretName:           className,
			Minimum:              pool.Minimum,
			Maximum:              pool.Maximum,
			MaxSurge:             pool.MaxSurge,
			MaxUnavailable:       pool.MaxUnavailable,
			Labels:               pool.Labels,
			Annotations:          pool.Annotations,
			Taints:               pool.Taints,
			MachineConfiguration: genericworkeractuator.ReadMachineConfiguration(pool),
		})

		machineClassSpec["name"] = className
		secretMap := machineClassSpec["secret"].(map[string]interface{})
		for k, v := range machineClassSecretData {
			secretMap[k] = string(v)
		}

		machineClasses = append(machineClasses, machineClassSpec)

	}
	w.machineDeployments = machineDeployments
	w.machineClasses = machineClasses

	return nil
}

type machineValues struct {
	MachineTypeOptions *apis.MachineTypeOptions
}

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
