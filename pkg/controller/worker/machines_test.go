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

// Package worker contains functions used at the worker controller
package worker

import (
	"context"
	"errors"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/mock"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/extensions/pkg/controller"
	gardenerclient "github.com/gardener/gardener/pkg/client/kubernetes"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/golang/mock/gomock"
	mcmv1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// newWorkerDelegate creates a new context for a worker reconciliation.
func newWorkerDelegate(
	client *mockclient.MockClient,

	clientContext common.ClientContext,

	seedChartApplier gardenerclient.ChartApplier,
	serverVersion string,

	worker *v1alpha1.Worker,
	cluster *v1alpha1.Cluster,
) (genericactuator.WorkerDelegate, error) {
	var decodedCluster *controller.Cluster

	if nil != cluster {
		newDecodedCluster, err := mock.DecodeCluster(cluster)
		if nil != err {
			return nil, err
		}

		decodedCluster = newDecodedCluster
	}

	workerDelegate, err := NewWorkerDelegate(clientContext, seedChartApplier, serverVersion, worker, decodedCluster)
	if nil != err {
		return nil, err
	}

	inject.ClientInto(client, workerDelegate)
	return workerDelegate, nil
}

var _ = Describe("Machines", func() {
	var mockTestEnv    mock.MockTestEnv

	var _ = BeforeSuite(func() {
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.HcloudClient)
		mock.SetupImagesEndpointOnMux(mockTestEnv.Mux)
	})

	var _ = AfterSuite(func() {
		mockTestEnv.Teardown()
	})

	Describe("#MachineClass", func() {
		It("should return the correct kind of the machine class", func() {
			workerDelegate, err := newWorkerDelegate(mockTestEnv.Client, common.NewClientContext(nil, nil, nil), nil, "", nil, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(workerDelegate.MachineClass()).To(Equal(&mcmv1alpha1.MachineClass{}))
		})
	})

	Describe("#MachineClassKind", func() {
		It("should return the correct kind of the machine class", func() {
			workerDelegate, err := newWorkerDelegate(mockTestEnv.Client, common.NewClientContext(nil, nil, nil), nil, "", nil, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(workerDelegate.MachineClassKind()).To(Equal("MachineClass"))
		})
	})

	Describe("#GenerateMachineDeployments", func() {
		type setup struct {
		}

		type action struct {
			cluster *v1alpha1.Cluster
			worker *v1alpha1.Worker
		}

		type expect struct {
			errToHaveOccurred          bool
			err                        error
			numberOfMachineDeployments int
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.TODO()

				mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestWorkerSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret) error {
					secret.Data = map[string][]byte{
						"hcloudToken": []byte("dummy-token"),
					}

					return nil
				}).AnyTimes()

				workerDelegate, err := newWorkerDelegate(mockTestEnv.Client, common.NewClientContext(nil, nil, nil), nil, "", data.action.worker, data.action.cluster)
				Expect(err).NotTo(HaveOccurred())

				result, err := workerDelegate.GenerateMachineDeployments(ctx)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.err))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).Should(HaveLen(data.expect.numberOfMachineDeployments))
				}
			},

			Entry("should successfully generate machine deployments", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.NewWorker(),
				},
				expect: expect{
					errToHaveOccurred: false,
					numberOfMachineDeployments: 1,
				},
			}),

			Entry("should not generate machine deployments because of missing zones", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.ManipulateWorker(mock.NewWorker(), map[string]interface{}{ "Spec.Pools.0.Zones": []string{} }),
				},
				expect: expect{
					errToHaveOccurred: false,
					numberOfMachineDeployments: 0,
				},
			}),

			Entry("should not generate machine deployments because of missing zones", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.ManipulateWorker(
						mock.NewWorker(),
						map[string]interface{}{
							"Spec.Pools.0.MachineImage": v1alpha1.MachineImage{
								Name:    "test",
								Version: "1.0",
							},
						},
					),
				},
				expect: expect{
					err: errors.New("could not find machine image for test/1.0 neither in cloud profile nor in worker status"),
					errToHaveOccurred: true,
				},
			}),
		)
	})
})
