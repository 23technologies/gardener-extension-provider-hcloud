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
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerclient "github.com/gardener/gardener/pkg/client/kubernetes"
	mockkubernetes "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	mockclient "github.com/gardener/gardener/third_party/mock/controller-runtime/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/23technologies/gardener-extension-provider-hcloud/charts"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/mock"
	hcloudv1alpha1 "github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/v1alpha1"
)

// newWorkerDelegate creates a new context for a worker reconciliation.
func newWorkerDelegate(
	client *mockclient.MockClient,
	scheme *runtime.Scheme,
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

	workerDelegate, err := NewWorkerDelegate(client, scheme, seedChartApplier, serverVersion, worker, decodedCluster)
	if nil != err {
		return nil, err
	}

	return workerDelegate, nil
}

var (
	mockTestEnv mock.MockTestEnv
	scheme      *runtime.Scheme
)

var _ = BeforeSuite(func() {
	mockTestEnv = mock.NewMockTestEnv()

	apis.SetClientForToken("dummy-token", mockTestEnv.HcloudClient)
	mock.SetupImagesEndpointOnMux(mockTestEnv.Mux)

	scheme = runtime.NewScheme()
	_ = apis.AddToScheme(scheme)
	_ = hcloudv1alpha1.AddToScheme(scheme)

	mockTestEnv.Client.EXPECT().Get(gomock.Any(), kutil.Key(mock.TestNamespace, mock.TestWorkerSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret, _ ...k8sclient.GetOption) error {
		secret.Data = map[string][]byte{
			"hcloudToken": []byte("dummy-token"),
		}

		return nil
	}).AnyTimes()
})

var _ = AfterSuite(func() {
	mockTestEnv.Teardown()
})

var _ = Describe("Machines", func() {
	Describe("#DeployMachineClasses", func() {
		type setup struct {
		}

		type action struct {
			cluster *v1alpha1.Cluster
			worker  *v1alpha1.Worker
		}

		type expect struct {
			errToHaveOccurred bool
			err               error
			machineClasses    []map[string]interface{}
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		machineClassName := fmt.Sprintf("%s-%s-%s-%s", mock.TestNamespace, mock.TestWorkerPoolName, mock.TestZone, "2ef7b")

		DescribeTable("##table",
			func(data *data) {
				chartApplier := mockkubernetes.NewMockChartApplier(mockTestEnv.MockController)
				ctx := context.TODO()

				mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestWorkerSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret, _ ...k8sclient.GetOption) error {
					secret.Data = map[string][]byte{
						"hcloudToken": []byte("dummy-token"),
					}

					return nil
				}).AnyTimes()

				chartApplier.EXPECT().ApplyFromEmbeddedFS(
					ctx,
					charts.InternalChart,
					filepath.Join(charts.InternalChartsPath, "machineclass"),
					mock.TestNamespace,
					"machineclass",
					gardenerclient.Values(
						map[string]interface{}{"machineClasses": data.expect.machineClasses},
					),
				).AnyTimes()

				workerDelegate, err := newWorkerDelegate(mockTestEnv.Client, scheme, chartApplier, "", data.action.worker, data.action.cluster)
				Expect(err).NotTo(HaveOccurred())

				err = workerDelegate.DeployMachineClasses(ctx)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.err))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},

			Entry("should successfully deploy machine classes", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.NewWorker(),
				},
				expect: expect{
					errToHaveOccurred: false,
					machineClasses: []map[string]interface{}{
						{
							"name": machineClassName,
							"credentialsSecretRef": map[string]interface{}{
								"name":      "secret",
								"namespace": "test-namespace"},
							"cluster":          mock.TestNamespace,
							"zone":             mock.TestZone,
							"imageName":        fmt.Sprintf("%s-%s", mock.TestWorkerMachineImageName, mock.TestWorkerMachineImageVersion),
							"sshFingerprint":   mock.TestSSHFingerprint,
							"machineType":      mock.TestWorkerMachineType,
							"floatingPoolName": mock.TestFloatingPoolName,
							"networkName":      fmt.Sprintf("%s-workers", mock.TestNamespace),
							"tags": map[string]string{
								"mcm.gardener.cloud/cluster": mock.TestNamespace,
								"mcm.gardener.cloud/role":    "node",
							},
							"secret": map[string]interface{}{
								"hcloudToken": []byte("dummy-token"),
								"userData":    mock.TestWorkerUserData,
							},
						},
					},
				},
			}),

			Entry("should not generate machine classes because of missing zones", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.ManipulateWorker(mock.NewWorker(), map[string]interface{}{"Spec.Pools.0.Zones": []string{}}),
				},
				expect: expect{
					errToHaveOccurred: false,
					machineClasses:    nil,
				},
			}),
			Entry("should fail because of invalid image name", &data{
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
					err:               errors.New("could not find machine image for test/1.0 neither in cloud profile nor in worker status"),
					errToHaveOccurred: true,
				},
			}),
		)
	})

	Describe("#GenerateMachineDeployments", func() {
		type setup struct {
		}

		type action struct {
			cluster *v1alpha1.Cluster
			worker  *v1alpha1.Worker
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

				mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestWorkerSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret, _ ...k8sclient.GetOption) error {
					secret.Data = map[string][]byte{
						"hcloudToken": []byte("dummy-token"),
					}

					return nil
				}).AnyTimes()

				workerDelegate, err := newWorkerDelegate(mockTestEnv.Client, scheme, nil, "", data.action.worker, data.action.cluster)
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
					errToHaveOccurred:          false,
					numberOfMachineDeployments: 1,
				},
			}),

			Entry("should not generate machine deployments because of missing zones", &data{
				setup: setup{},
				action: action{
					mock.NewCluster(),
					mock.ManipulateWorker(mock.NewWorker(), map[string]interface{}{"Spec.Pools.0.Zones": []string{}}),
				},
				expect: expect{
					errToHaveOccurred:          false,
					numberOfMachineDeployments: 0,
				},
			}),
			Entry("should fail because of invalid image name", &data{
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
					err:               errors.New("could not find machine image for test/1.0 neither in cloud profile nor in worker status"),
					errToHaveOccurred: true,
				},
			}),
		)
	})
})
