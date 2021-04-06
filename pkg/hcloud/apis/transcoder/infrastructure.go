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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func DecodeInfrastructureConfig(infra *runtime.RawExtension, fldPath *field.Path) (*apis.InfrastructureConfig, error) {
	infraConfig := &apis.InfrastructureConfig{}
	if _, _, err := decoder.Decode(infra.Raw, nil, infraConfig); err != nil {
		return nil, field.Invalid(fldPath, string(infra.Raw), "cannot be decoded")
	}

	return infraConfig, nil
}
