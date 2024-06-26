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

// Package install provides functions used for registration of hcloud.provider.extensions.config.gardener.cloud
package install

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config/v1alpha1"
)

var (
	schemeBuilder = runtime.NewSchemeBuilder(
		v1alpha1.AddToScheme,
		config.AddToScheme,
		setVersionPriority,
	)

	// AddToScheme adds all APIs to the scheme.
	AddToScheme = schemeBuilder.AddToScheme
)

// setVersionPriority is used to set priority of the scheme to the latest one.
//
// PARAMETERS
// scheme *runtime.Scheme Kubernetes scheme to set version in.
func setVersionPriority(scheme *runtime.Scheme) error {
	return scheme.SetVersionPriority(v1alpha1.SchemeGroupVersion)
}

// Install installs all APIs in the scheme.
//
// PARAMETERS
// scheme *runtime.Scheme Kubernetes scheme to install into.
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(AddToScheme(scheme))
}
