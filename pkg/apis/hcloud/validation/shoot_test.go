// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/gardener/gardener/pkg/apis/core"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

var _ = Describe("Shoot validation", func() {
	Describe("#ValidateNetworking", func() {
		networkingPath := field.NewPath("spec", "networking")

		It("should return no error because nodes CIDR was provided", func() {
			networking := &core.Networking{
				Nodes: ptr.To("1.2.3.4/5"),
			}

			errorList := ValidateNetworking(networking, networkingPath)

			Expect(errorList).To(BeEmpty())
		})

		It("should return an error because no nodes CIDR was provided", func() {
			networking := &core.Networking{}

			errorList := ValidateNetworking(networking, networkingPath)

			Expect(errorList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("spec.networking.nodes"),
				})),
			))
		})
	})
	Describe("#validateWorkerConfig", func() {
		var (
			nilPath *field.Path
			workers []core.Worker
		)

		BeforeEach(func() {
			workers = []core.Worker{
				{
					Name: "worker1",
					Volume: &core.Volume{
						Type:       ptr.To("Volume"),
						VolumeSize: "30G",
					},
					Minimum: 1,
					Maximum: 2,
					Zones:   []string{"1", "2"},
				},
				{
					Name: "worker2",
					Volume: &core.Volume{
						Type:       ptr.To("Volume"),
						VolumeSize: "20G",
					},
					Minimum: 1,
					Maximum: 2,
					Zones:   []string{"1", "2"},
				},
			}
		})

		Describe("#ValidateWorkers", func() {
			It("should pass because workers are configured correctly", func() {
				errorList := ValidateWorkers(workers, nil, nilPath)

				Expect(errorList).To(BeEmpty())
			})

			It("should forbid because worker does not specify a zone", func() {
				workers[0].Zones = nil

				errorList := ValidateWorkers(workers, nil, nilPath)

				Expect(errorList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("[0].zones"),
					})),
				))
			})

			Describe("#ValidateWorkersUpdate", func() {
				It("should pass because workers are unchanged", func() {
					newWorkers := copyWorkers(workers)
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(BeEmpty())
				})

				It("should allow adding workers", func() {
					newWorkers := append(workers[:0:0], workers...)
					workers = workers[:1]
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(BeEmpty())
				})

				It("should allow adding a zone to a worker", func() {
					newWorkers := copyWorkers(workers)
					newWorkers[0].Zones = append(newWorkers[0].Zones, "another-zone")
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(BeEmpty())
				})

				It("should forbid removing a zone from a worker", func() {
					newWorkers := copyWorkers(workers)
					newWorkers[1].Zones = newWorkers[1].Zones[1:]
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("[1].zones"),
						})),
					))
				})

				It("should forbid changing the zone order", func() {
					newWorkers := copyWorkers(workers)
					newWorkers[0].Zones[0] = workers[0].Zones[1]
					newWorkers[0].Zones[1] = workers[0].Zones[0]
					newWorkers[1].Zones[0] = workers[1].Zones[1]
					newWorkers[1].Zones[1] = workers[1].Zones[0]
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("[0].zones"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("[1].zones"),
						})),
					))
				})

				It("should forbid adding a zone while changing an existing one", func() {
					newWorkers := copyWorkers(workers)
					newWorkers = append(newWorkers, core.Worker{Name: "worker3", Zones: []string{"zone1"}})
					newWorkers[1].Zones[0] = workers[1].Zones[1]
					errorList := ValidateWorkersUpdate(workers, newWorkers, nilPath)

					Expect(errorList).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("[1].zones"),
						})),
					))
				})
			})

		})
	})
}) // Closes Describe("Shoot validation", ...)

func copyWorkers(workers []core.Worker) []core.Worker {
	cp := append(workers[:0:0], workers...)
	for i := range cp {
		cp[i].Zones = append(workers[i].Zones[:0:0], workers[i].Zones...)
	}
	return cp
}
