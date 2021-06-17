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

// Package validation contains functions to validate controller specifications
package validation

import (
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateControlPlaneConfig validates a ControlPlaneConfig object.
func ValidateControlPlaneConfig(controlPlaneConfig *apis.ControlPlaneConfig, allowedZones, workerZones sets.String, version string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(controlPlaneConfig.Zone) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("zone"), "must provide the name of a zone in this region"))
	}

	if !workerZones.Has(controlPlaneConfig.Zone) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("zone"), controlPlaneConfig.Zone, "must be part of at least one worker zone"))
	}

	return allErrs
}
