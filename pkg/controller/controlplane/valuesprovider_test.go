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

// Package controlplane contains functions used at the controlplane controller
package controlplane

import (
	"context"
	"errors"
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/mock"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	fakesecretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager/fake"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

var (
	mockTestEnv        mock.MockTestEnv
	vp                 genericactuator.ValuesProvider
	fakeSecretsManager secretsmanager.Interface
)

var _ = BeforeSuite(func() {
	mockTestEnv = mock.NewMockTestEnv()

	apis.SetClientForToken("dummy-token", mockTestEnv.HcloudClient)
	mock.SetupImagesEndpointOnMux(mockTestEnv.Mux)

	vp = NewValuesProvider(logger, "garden")
	inject.ClientInto(mockTestEnv.Client, vp)

	fakeClient := fakeclient.NewClientBuilder().Build()
	fakeSecretsManager = fakesecretsmanager.New(fakeClient, mock.TestNamespace)

	By("creating secrets managed outside of this package for whose secretsmanager.Get() will be called")
	Expect(fakeClient.Create(context.TODO(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "cloud-controller-manager-server", Namespace: mock.TestNamespace}})).To(Succeed())
})

var _ = AfterSuite(func() {
	mockTestEnv.Teardown()
})

var _ = Describe("ValuesProvider", func() {
	Describe("#GetControlPlaneChartValues", func() {
		type setup struct {
		}

		type action struct {
			cp         *v1alpha1.ControlPlane
			cluster    *v1alpha1.Cluster
			scaledDown bool
		}

		type expect struct {
			errToHaveOccurred bool
			err               error
			comparator        func(mapValues map[string]interface{}) error
		}

		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("##table",
			func(data *data) {
				ctx := context.TODO()

				mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestControlPlaneSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret) error {
					secret.Data = map[string][]byte{
						"hcloudToken": []byte("dummy-token"),
					}

					return nil
				}).AnyTimes()

				decodedCluster, err := mock.DecodeCluster(data.action.cluster)
				Expect(err).NotTo(HaveOccurred())

				values, err := vp.GetControlPlaneChartValues(ctx, data.action.cp, decodedCluster, fakeSecretsManager, map[string]string{}, data.action.scaledDown)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(data.expect.err))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(values).Should(WithTransform(data.expect.comparator, Succeed()))
				}
			},

			Entry("should successfully return control plane chart values", &data{
				setup: setup{},
				action: action{
					mock.NewControlPlane(),
					mock.NewCluster(),
					false,
				},
				expect: expect{
					errToHaveOccurred: false,
					comparator: func(mapValues map[string]interface{}) error {
						mapValue, ok := mapValues["global"].(map[string]interface{})
						if !ok {
							return errors.New("global is missing")
						}

						value, ok := mapValue["genericTokenKubeconfigSecretName"]
						if !ok || value != mock.TestClusterGenericTokenKubeconfigSecretName {
							return fmt.Errorf("%q is invalid for global.genericTokenKubeconfigSecretName", value)
						}

						mapValue, ok = mapValues["cloud-controller-manager"].(map[string]interface{})
						if !ok {
							return errors.New("cloud-controller-manager is missing")
						}

						value, ok = mapValue["podRegion"]
						if !ok || value != mock.TestRegion {
							return fmt.Errorf("%q is invalid for cloud-controller-manager.podRegion", value)
						}

						mapValue, ok = mapValues["hcloud-csi-controller"].(map[string]interface{})
						if !ok {
							return errors.New("hcloud-csi-controller is missing")
						}

						value, ok = mapValue["csiRegion"]
						if !ok || value != mock.TestRegion {
							return fmt.Errorf("%q is invalid for hcloud-csi-controller.csiRegion", value)
						}

						return nil
					},
				},
			}),
		)
	})
})
