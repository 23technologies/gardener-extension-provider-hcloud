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

	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	mockkubernetes "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	"github.com/golang/mock/gomock"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/onsi/ginkgo"
)

const (
	TestNamespace = "test-namespace"
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

		Server: server,
		Mux:    mux,
		HcloudClient: hcloudClient,
	}
}
