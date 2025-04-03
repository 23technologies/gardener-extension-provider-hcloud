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

// Package controlplane contains functions used to provide /controlplane
package controlplane

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/original/components/kubelet"
	oscutils "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
)

var (
	logger = log.Log.WithName("hcloud-controlplane-webhook")
)

// AddToManager creates a webhook and adds it to the manager.
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	logger.Info("Adding webhook to manager")
	fciCodec := oscutils.NewFileContentInlineCodec()
	return controlplane.New(mgr, controlplane.Args{
		Kind:     controlplane.KindShoot,
		Provider: hcloud.Type,
		Types: []extensionswebhook.Type{
			{Obj: &appsv1.Deployment{}},
			{Obj: &vpaautoscalingv1.VerticalPodAutoscaler{}},
			{Obj: &extensionsv1alpha1.OperatingSystemConfig{}},
		},
		ObjectSelector: &metav1.LabelSelector{MatchLabels: map[string]string{v1beta1constants.LabelExtensionProviderMutatedByControlplaneWebhook: "true"}},
		Mutator: genericmutator.NewMutator(
			mgr,
			NewEnsurer(logger),
			oscutils.NewUnitSerializer(),
			kubelet.NewConfigCodec(fciCodec),
			fciCodec,
			logger,
		),
	})
}
