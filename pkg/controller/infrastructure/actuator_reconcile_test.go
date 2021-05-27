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

// Package infrastructure contains functions used at the infrastructure controller
package infrastructure

import (
	"context"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/mock"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/extensions"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

var _ = Describe("ActuatorReconcile", func() {
	var actuator    infrastructure.Actuator
	var cluster     *extensions.Cluster
	var ctx         context.Context
	var mockTestEnv mock.MockTestEnv

	var _ = BeforeSuite(func() {
		ctx = context.TODO()
		mockTestEnv = mock.NewMockTestEnv()

		apis.SetClientForToken("dummy-token", mockTestEnv.HcloudClient)
		mock.SetupNetworksEndpointOnMux(mockTestEnv.Mux)
		mock.SetupSshKeysEndpointOnMux(mockTestEnv.Mux)

		newCluster, err := mock.DecodeCluster(mock.NewCluster())
		Expect(err).NotTo(HaveOccurred())
		cluster = newCluster

		actuator = NewActuator("garden")
		inject.ClientInto(mockTestEnv.Client, actuator)
	})

	var _ = AfterSuite(func() {
		mockTestEnv.Teardown()
	})

	Describe("#Reconcile", func() {
		It("should successfully reconcile", func() {
			mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestInfrastructureSecretName), gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret) error {
				secret.Data = map[string][]byte{
					"hcloudToken": []byte("dummy-token"),
				}

				return nil
			})

			mockTestEnv.Client.EXPECT().Get(ctx, kutil.Key(mock.TestNamespace, mock.TestInfrastructureName), gomock.AssignableToTypeOf(&v1alpha1.Infrastructure{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, infrastructure *v1alpha1.Infrastructure) error {
				return nil
			})

			mockTestEnv.Client.EXPECT().Status().Return(mockTestEnv.Client)
			mockTestEnv.Client.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&v1alpha1.Infrastructure{}), gomock.Any()).Times(1)

			err := actuator.Reconcile(ctx, mock.NewInfrastructure(), cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
