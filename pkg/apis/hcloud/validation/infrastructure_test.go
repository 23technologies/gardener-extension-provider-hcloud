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
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
)

var _ = Describe("InfrastructureConfig validation", func() {
	var (
		nilPath *field.Path

		infrastructureConfig *api.InfrastructureConfig

		nodes       = "10.250.0.0/16"
		invalidCIDR = "invalid-cidr"
	)

	BeforeEach(func() {
		infrastructureConfig = &api.InfrastructureConfig{
			Networks: &api.Networks{
				Workers: "10.250.0.0/16",
			},
		}
	})

	Context("CIDR", func() {
		It("should forbid empty workers CIDR", func() {
			infrastructureConfig.Networks.Workers = ""

			errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, nilPath)

			Expect(errorList).To(ConsistOfFields(Fields{
				"Type":   Equal(field.ErrorTypeRequired),
				"Field":  Equal("networks.workers"),
				"Detail": Equal("must specify the network range for the worker network"),
			}))
		})

		It("should forbid invalid workers CIDR", func() {
			infrastructureConfig.Networks.Workers = invalidCIDR

			errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, nilPath)

			Expect(errorList).To(ConsistOfFields(Fields{
				"Type":   Equal(field.ErrorTypeInvalid),
				"Field":  Equal("networks.workers"),
				"Detail": Equal("invalid CIDR address: invalid-cidr"),
			}))
		})

		It("should forbid workers CIDR which are not in Nodes CIDR", func() {
			infrastructureConfig.Networks.Workers = "1.1.1.1/32"

			errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, nilPath)

			Expect(errorList).To(ConsistOfFields(Fields{
				"Type":   Equal(field.ErrorTypeInvalid),
				"Field":  Equal("networks.workers"),
				"Detail": Equal(`must be a subset of "networking.nodes" ("10.250.0.0/16")`),
			}))
		})

		It("should forbid non canonical CIDRs", func() {
			nodeCIDR := "10.250.0.3/16"

			infrastructureConfig.Networks.Workers = "10.250.3.8/24"

			errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodeCIDR, nilPath)
			Expect(errorList).To(HaveLen(1))

			Expect(errorList).To(ConsistOfFields(Fields{
				"Type":   Equal(field.ErrorTypeInvalid),
				"Field":  Equal("networks.workers"),
				"Detail": Equal("must be valid canonical CIDR"),
			}))
		})

	})
})
