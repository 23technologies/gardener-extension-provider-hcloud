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

	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/pkg/extensions"
	mockclient "github.com/gardener/gardener/third_party/mock/controller-runtime/client"
	mockmanager "github.com/gardener/gardener/third_party/mock/controller-runtime/manager"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/mock"
	hcloudv1alpha1 "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/v1alpha1"
)

var (
	infraActuator infrastructure.Actuator
	cluster       *extensions.Cluster
	ctx           context.Context
	mockTestEnv   mock.MockTestEnv
	mgr           *mockmanager.MockManager
	scheme        *runtime.Scheme
	config        *rest.Config
	sw            *mockclient.MockStatusWriter
)

var _ = BeforeSuite(func() {
	ctx = context.TODO()
	mockTestEnv = mock.NewMockTestEnv()

	api.SetClientForToken("dummy-token", mockTestEnv.HcloudClient)
	mock.SetupLocationsEndpointOnMux(mockTestEnv.Mux)
	mock.SetupNetworksEndpointOnMux(mockTestEnv.Mux)
	mock.SetupPlacementGroupsEndpointOnMux(mockTestEnv.Mux)
	mock.SetupSshKeysEndpointOnMux(mockTestEnv.Mux)

	newCluster, err := mock.DecodeCluster(mock.NewCluster())
	Expect(err).NotTo(HaveOccurred())
	cluster = newCluster

	sw = mockclient.NewMockStatusWriter(mockTestEnv.MockController)

	mgr = mockmanager.NewMockManager(mockTestEnv.MockController)
	mgr.EXPECT().GetClient().Return(mockTestEnv.Client)

	scheme = runtime.NewScheme()
	_ = api.AddToScheme(scheme)
	_ = hcloudv1alpha1.AddToScheme(scheme)
	mgr.EXPECT().GetScheme().Return(scheme)
	mgr.EXPECT().GetConfig().Return(config)
	infraActuator = NewActuator(mgr, "garden")
})

var _ = AfterSuite(func() {
	mockTestEnv.Teardown()
})

var _ = Describe("ActuatorReconcile", func() {
	Describe("#Reconcile", func() {
		It("should successfully reconcile", func() {
			mockTestEnv.Client.EXPECT().Get(gomock.Any(), k8sclient.ObjectKey{Namespace: mock.TestNamespace, Name: mock.TestInfrastructureSecretName}, gomock.AssignableToTypeOf(&corev1.Secret{})).DoAndReturn(func(_ context.Context, _ k8sclient.ObjectKey, secret *corev1.Secret, _ ...k8sclient.GetOption) error {
				secret.Data = map[string][]byte{
					"hcloudToken": []byte("dummy-token"),
				}
				return nil
			}).AnyTimes() // Allow fetching the secret multiple times

			mockTestEnv.Client.EXPECT().Status().Return(sw).AnyTimes()
			sw.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil) // Expect status patch

			err := infraActuator.Reconcile(ctx, logr.Logger{}, mock.NewInfrastructure(), cluster)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
