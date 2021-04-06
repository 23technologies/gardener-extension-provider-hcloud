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

// Package transcoder is used for API related object transformations
package transcoder

import (
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func DecodeControlPlaneConfig(cp *runtime.RawExtension, fldPath *field.Path) (*apis.ControlPlaneConfig, error) {
	controlPlaneConfig := &apis.ControlPlaneConfig{}
	if _, _, err := decoder.Decode(cp.Raw, nil, controlPlaneConfig); err != nil {
		return nil, field.Invalid(fldPath, string(cp.Raw), "cannot be decoded")
	}

	return controlPlaneConfig, nil
}

func DecodeControlPlaneConfigFromControllerCluster(cluster *controller.Cluster) (*apis.ControlPlaneConfig, error) {
	controlPlaneConfig := &apis.ControlPlaneConfig{}
	if cluster.Shoot.Spec.Provider.ControlPlaneConfig != nil {
		if _, _, err := decoder.Decode(cluster.Shoot.Spec.Provider.ControlPlaneConfig.Raw, nil, controlPlaneConfig); err != nil {
			return nil, errors.Wrapf(err, "could not decode providerConfig of controlplane '%s'", cluster.ObjectMeta.Name)
		}
	}

	return controlPlaneConfig, nil
}
