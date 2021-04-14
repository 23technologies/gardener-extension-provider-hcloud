// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package infrastructure

import (
	"context"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type actuator struct {
	common.ChartRendererContext

	logger   logr.Logger
	gardenID string
}

// NewActuator creates a new Actuator that updates the status of the handled Infrastructure resources.
func NewActuator(gardenID string) infrastructure.Actuator {
	return &actuator{
		logger:   log.Log.WithName("infrastructure-actuator"),
		gardenID: gardenID,
	}
}

func (a *actuator) Reconcile(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return a.reconcile(ctx, config, cluster)
}

func (a *actuator) Delete(ctx context.Context, config *extensionsv1alpha1.Infrastructure, cluster *controller.Cluster) error {
	return a.delete(ctx, config, cluster)
}

func (a *actuator) updateProviderStatus(ctx context.Context, infra *extensionsv1alpha1.Infrastructure) error {
	infraConfig, err := transcoder.DecodeInfrastructureConfigFromInfrastructure(infra)
	if err != nil {
		return err
	}

	infraStatus := v1alpha1.InfrastructureStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "InfrastructureStatus",
		},
	}

	if "" != infraConfig.FloatingPoolName {
		infraStatus.FloatingPoolName = infraConfig.FloatingPoolName
	}

	return controller.TryUpdateStatus(ctx, retry.DefaultBackoff, a.Client(), infra, func() error {
		infra.Status.ProviderStatus = &runtime.RawExtension{
			Object: &infraStatus,
		}

		return nil
	})
}
