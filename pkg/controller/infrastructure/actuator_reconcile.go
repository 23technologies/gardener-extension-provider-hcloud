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

// Package infrastructure contains functions used at the infrastructure controller
package infrastructure

import (
	"context"
	"strconv"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/infrastructure/ensurer"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/controller"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// reconcile reconciles the infrastructure config.
//
// PARAMETERS
// ctx     context.Context                    Execution context
// infra   *extensionsv1alpha1.Infrastructure Infrastructure struct
// cluster *extensionscontroller.Cluster      Cluster struct
func (a *actuator) reconcile(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster) error {
	actuatorConfig, err := a.getActuatorConfig(ctx, infra, cluster)
	if err != nil {
		return err
	}

	cpConfig, err := transcoder.DecodeControlPlaneConfigFromControllerCluster(cluster)
	if err != nil {
		return err
	}

	infraConfig, err := transcoder.DecodeInfrastructureConfigFromInfrastructure(infra)
	if err != nil {
		return err
	}

	client := apis.GetClientForToken(string(actuatorConfig.token))

	oldProviderStatus, err := transcoder.DecodeInfrastructureStatus(infra.Status.GetProviderStatus())
	if err != nil {
		return err
	}
	oldFingerprint := oldProviderStatus.SSHFingerprint
	newFingerprint, err := apis.GetSSHFingerprint(infra.Spec.SSHPublicKey)
	if nil != err {
		return err
	}

	if oldFingerprint != newFingerprint {
		sshKey, _, err := client.SSHKey.GetByFingerprint(ctx, oldFingerprint)
		if nil != err {
			return err
		} else if sshKey != nil {
			_, err := client.SSHKey.Delete(ctx, sshKey)
			if nil != err {
				return err
			}
		}
	}

	sshFingerprint, err := ensurer.EnsureSSHPublicKey(ctx, client, infra.Spec.SSHPublicKey)
	if err != nil {
		return err
	}

	workerNetworkID, err := ensurer.EnsureNetworks(ctx, client, infra.Namespace, cpConfig.Zone, actuatorConfig.infraConfig.Networks)
	if err != nil {
		return err
	}

	infraStatus := &v1alpha1.InfrastructureStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "InfrastructureStatus",
		},
		SSHFingerprint: sshFingerprint,
	}

	if "" != infraConfig.FloatingPoolName {
		infraStatus.FloatingPoolName = infraConfig.FloatingPoolName
	}

	if workerNetworkID > -1 {
		infraStatus.NetworkIDs = &v1alpha1.InfrastructureConfigNetworkIDs{
			Workers: strconv.Itoa(workerNetworkID),
		}
	}

	return a.updateProviderStatus(ctx, infra, infraStatus)
}

// reconcileOnErrorCleanup cleans up a failed reconcile request
//
// PARAMETERS
// ctx     context.Context                    Execution context
// infra   *extensionsv1alpha1.Infrastructure Infrastructure struct
// cluster *extensionscontroller.Cluster      Cluster struct
// err     error                              Error encountered
func (a *actuator) reconcileOnErrorCleanup(ctx context.Context, infra *extensionsv1alpha1.Infrastructure, cluster *extensionscontroller.Cluster, err error) {
	actuatorConfig, _ := a.getActuatorConfig(ctx, infra, cluster)
	resultData := ctx.Value(controller.CtxWrapDataKey("MethodData")).(*controller.InfrastructureReconcileMethodData)

	if nil != actuatorConfig {
		client := apis.GetClientForToken(string(actuatorConfig.token))

		if resultData.NetworkID != 0 {
			networkIDs := &apis.InfrastructureConfigNetworkIDs{
				Workers: strconv.Itoa(resultData.NetworkID),
			}

			ensurer.EnsureNetworksDeleted(ctx, client, infra.Namespace, networkIDs)
		}

		if resultData.SSHKeyID != 0 {
			sshKeyID := strconv.Itoa(resultData.SSHKeyID)
			ensurer.EnsureSSHPublicKeyDeleted(ctx, client, sshKeyID)
		}
	}
}
