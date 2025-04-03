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
	"github.com/gardener/gardener/pkg/apis/core"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateNetworking validates the network settings of a Shoot.
func ValidateNetworking(networking *core.Networking, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if networking.Nodes == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("nodes"), "a nodes CIDR must be provided for hcloud shoots"))
	}

	if core.IsIPv6SingleStack(networking.IPFamilies) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ipFamilies"), networking.IPFamilies, "IPv6 single-stack networking is not supported"))
	}

	if len(networking.IPFamilies) > 1 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ipFamilies"), networking.IPFamilies, "dual-stack networking is not supported"))
	}

	return allErrs
}
