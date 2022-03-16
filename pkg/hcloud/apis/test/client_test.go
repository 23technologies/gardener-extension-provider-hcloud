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

package test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/mock"
	"github.com/hetznercloud/hcloud-go/hcloud"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	client        *hcloud.Client
	ctx           context.Context
	mockTestEnv   mock.MockTestEnv
)

var _ = BeforeSuite(func() {
	ctx = context.TODO()
	mockTestEnv = mock.NewMockTestEnv()

	mock.SetupTestTokenEndpointOnMux(mockTestEnv.Mux)

})

var _ = AfterSuite(func() {
	mockTestEnv.Teardown()
})

var _ = Describe("Api", func() {
	Describe("#GetClientForToken", func() {
		It("should return StatusOK", func() {

			hcloudClient := hcloud.NewClient(
				hcloud.WithEndpoint(mockTestEnv.Server.URL),
				hcloud.WithHTTPClient(mockTestEnv.Server.Client()),
				hcloud.WithToken("dummy-token"),
			)

			apis.SetClientForToken("dummy-token", hcloudClient)
			client = apis.GetClientForToken("dummy-token")

			req, _ := client.NewRequest(ctx, "GET", fmt.Sprintf("/testtokenendpoint") , nil)
			resp, _ := client.Do(req, &req.Body)

			Expect( resp.StatusCode ).To(Equal(http.StatusOK))

		})

		It("should return StatusOK", func() {

			hcloudClient := hcloud.NewClient(
				hcloud.WithEndpoint(mockTestEnv.Server.URL),
				hcloud.WithHTTPClient(mockTestEnv.Server.Client()),
				hcloud.WithToken("dummy-token"),
			)

			apis.SetClientForToken("dummy-token", hcloudClient)
			client = apis.GetClientForToken("dummy-token\n")

			req, _ := client.NewRequest(ctx, "GET", fmt.Sprintf("/testtokenendpoint") , nil)
			resp, _ := client.Do(req, &req.Body)

			Expect( resp.StatusCode ).To(Equal(http.StatusOK))

		})

		It("should return StatusForbidden", func() {

			hcloudClient := hcloud.NewClient(
				hcloud.WithEndpoint(mockTestEnv.Server.URL),
				hcloud.WithHTTPClient(mockTestEnv.Server.Client()),
				hcloud.WithToken("bogo-token"),
			)

			apis.SetClientForToken("bogo-token", hcloudClient)
			client = apis.GetClientForToken("bogo-token")

			req, _ := client.NewRequest(ctx, "GET", fmt.Sprintf("/testtokenendpoint") , nil)
			resp, _ := client.Do(req, &req.Body)

			Expect( resp.StatusCode ).To(Equal(http.StatusForbidden))

		})
	})
})
