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

	})
})
