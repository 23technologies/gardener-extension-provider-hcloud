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

// Package mock provides all methods required to simulate a HCloud provider environment
package mock

import (
	"net/http"
	"net/http/httptest"

	mockkubernetes "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	mockclient "github.com/gardener/gardener/third_party/mock/controller-runtime/client"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/onsi/ginkgo/v2"
	gomock "go.uber.org/mock/gomock"
)

const (
	TestFloatingPoolName = "MY-FLOATING-POOL"
	TestNamespace        = "test-namespace"
	TestRegion           = "hel1"
	TestSSHFingerprint   = "b0:aa:73:08:9e:4f:6b:d1:3f:12:eb:66:78:61:63:08"
	TestSSHPublicKey     = "ecdsa-sha2-nistp384 AAAAE2VjZHNhLXNoYTItbmlzdHAzODQAAAAIbmlzdHAzODQAAABhBJ9S5cCzfygWEEVR+h3yDE83xKiTlc7S3pC3IadoYu/HAmjGPNRQZWLPCfZe5K3PjOGgXghmBY22voYl7bSVjy+8nZRPuVBuFDZJ9xKLPBImQcovQ1bMn8vXno4fvAF4KQ=="
	TestZone             = "hel1-dc2"
)

// MockTestEnv represents the test environment for testing HCloud API calls
type MockTestEnv struct {
	ChartApplier   *mockkubernetes.MockChartApplier
	Client         *mockclient.MockClient
	MockController *gomock.Controller
	StatusWriter   *mockclient.MockStatusWriter

	Server       *httptest.Server
	Mux          *http.ServeMux
	HcloudClient *hcloud.Client
}

// Teardown shuts down the test environment
func (env *MockTestEnv) Teardown() {
	env.MockController.Finish()

	env.ChartApplier = nil
	env.Client = nil
	env.MockController = nil
	env.StatusWriter = nil

	env.Server.Close()

	env.Server = nil
	env.Mux = nil
	env.HcloudClient = nil
}

// NewMockTestEnv generates a new, unconfigured test environment for testing purposes.
func NewMockTestEnv() MockTestEnv {
	ctrl := gomock.NewController(ginkgo.GinkgoT())

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	hcloudClient := hcloud.NewClient(
		hcloud.WithEndpoint(server.URL),
		hcloud.WithHTTPClient(server.Client()),
	)

	return MockTestEnv{
		ChartApplier:   mockkubernetes.NewMockChartApplier(ctrl),
		Client:         mockclient.NewMockClient(ctrl),
		MockController: ctrl,
		StatusWriter:   mockclient.NewMockStatusWriter(ctrl),

		Server:       server,
		Mux:          mux,
		HcloudClient: hcloudClient,
	}
}
