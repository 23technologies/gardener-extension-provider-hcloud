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
	cidrvalidation "github.com/gardener/gardener/pkg/utils/validation/cidr"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
)

// ValidateInfrastructureConfig validates a InfrastructureConfig object.
func ValidateInfrastructureConfig(infra *api.InfrastructureConfig, nodesCIDR *string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	networkingPath := field.NewPath("networking")
	var nodes cidrvalidation.CIDR
	if nodesCIDR != nil {
		nodes = cidrvalidation.NewCIDR(*nodesCIDR, networkingPath.Child("nodes"))
	}

	networksPath := fldPath.Child("networks")
	if len(infra.Networks.Workers) == 0 {
		allErrs = append(allErrs, field.Required(networksPath.Child("workers"), "must specify the network range for the worker network"))
	}

	var workerCIDR cidrvalidation.CIDR
	if infra.Networks.Workers != "" {
		workerCIDR = cidrvalidation.NewCIDR(infra.Networks.Workers, networksPath.Child("workers"))
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRParse(workerCIDR)...)
		allErrs = append(allErrs, cidrvalidation.ValidateCIDRIsCanonical(networksPath.Child("workers"), infra.Networks.Workers)...)
	}

	if nodes != nil {
		allErrs = append(allErrs, nodes.ValidateSubset(workerCIDR)...)
	}

	return allErrs
}

// ValidateInfrastructureConfigUpdate validates a InfrastructureConfig object.
func ValidateInfrastructureConfigUpdate(oldConfig, newConfig *api.InfrastructureConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	newNetworks := newConfig.DeepCopy().Networks
	oldNetworks := oldConfig.DeepCopy().Networks

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(newNetworks, oldNetworks, fldPath.Child("networks"))...)
	return allErrs
}
