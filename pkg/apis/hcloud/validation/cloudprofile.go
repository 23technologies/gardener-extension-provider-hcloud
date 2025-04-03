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
	"regexp"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
)

var validLoadBalancerSizeValues = sets.NewString("SMALL", "MEDIUM", "LARGE")
var namePrefixPattern = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

// ValidateCloudProfileConfig validates a CloudProfileConfig object.
func ValidateCloudProfileConfig(profileSpec *gardencorev1beta1.CloudProfileSpec, profileConfig *hcloud.CloudProfileConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}
