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

// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta
// +groupName=hcloud.provider.extensions.gardener.cloud

//go:generate ../../../../hack/update-codegen.sh

// Package v1alpha1 contains the HCloud provider API resources.
package v1alpha1 // import "github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
