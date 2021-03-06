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
	"strings"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestInfrastructureName = "abc"
	TestInfrastructureProviderConfig = `{
		"apiVersion": "hcloud.provider.extensions.gardener.cloud/v1alpha1",
		"kind": "InfrastructureConfig",
		"floatingPoolName": "MY-FLOATING-POOL",
		"networks": {"workers": "10.250.0.0/19"}
	}`
	TestInfrastructureSecretName = "cloudprovider"
	TestInfrastructureWorkersNetworkName = "test-namespace-workers"
)

// NewInfrastructure generates a new provider specification for testing purposes.
func NewInfrastructure() *v1alpha1.Infrastructure {
	return &v1alpha1.Infrastructure{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions.gardener.cloud",
			Kind:       "Infrastructure",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestInfrastructureName,
			Namespace: TestNamespace,
		},
		Spec: v1alpha1.InfrastructureSpec{
			Region: TestRegion,
			SecretRef: corev1.SecretReference{
				Name: TestInfrastructureSecretName,
				Namespace: TestNamespace,
			},
			DefaultSpec: v1alpha1.DefaultSpec{
				ProviderConfig: &runtime.RawExtension{
					Raw: []byte(TestInfrastructureProviderConfig),
				},
			},
			SSHPublicKey: []byte(TestSSHPublicKey),
		},
	}
}

// NewInfrastructureConfigSpec generates a new infrastructure config specification for testing purposes.
func NewInfrastructureConfigSpec() *apis.InfrastructureConfig {
	return &apis.InfrastructureConfig{
		FloatingPoolName: TestFloatingPoolName,
		Networks: &apis.Networks{
			Workers: TestInfrastructureWorkersNetworkName,
		},
	}
}

// ManipulateInfrastructure changes given provider specification.
//
// PARAMETERS
// infrastructure *extensions.Infrastructure Infrastructure specification
// data           map[string]interface{}     Members to change
func ManipulateInfrastructure(infrastructure *v1alpha1.Infrastructure, data map[string]interface{}) *v1alpha1.Infrastructure {
	for key, value := range data {
		if (strings.Index(key, "ObjectMeta") == 0) {
			manipulateStruct(&infrastructure.ObjectMeta, key[11:], value)
		} else if (strings.Index(key, "Spec") == 0) {
			manipulateStruct(&infrastructure.Spec, key[7:], value)
		} else if (strings.Index(key, "TypeMeta") == 0) {
			manipulateStruct(&infrastructure.TypeMeta, key[9:], value)
		} else {
			manipulateStruct(&infrastructure, key, value)
		}
	}

	return infrastructure
}

// SetupNetworksEndpointOnMux configures a "/ssh_keys" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupNetworksEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/networks", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		queryParams := req.URL.Query()

		res.Write([]byte(`
{
	"networks": [
		`))

		if (queryParams.Get("name") == TestInfrastructureWorkersNetworkName) {
			res.Write([]byte(`
{
	"id": 42,
	"name": "Simulated network",
	"range": "127.0.0.0/8",
	"subnets": [],
	"routes": [],
	"servers": [],
	"load_balancers": [],
	"labels": {},
	"created": "2016-01-30T23:50:00+00:00"
}
			`))
		}

		res.Write([]byte(`
	]
}
		`))
	})
}

// SetupSshKeysEndpointOnMux configures a "/ssh_keys" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupSshKeysEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/ssh_keys", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		queryParams := req.URL.Query()

		res.Write([]byte(`
{
	"ssh_keys": [
		`))

		if (queryParams.Get("fingerprint") == TestSSHFingerprint) {
			res.Write([]byte(`
{
	"id": 42,
	"name": "Simulated ssh key",
	"fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff",
	"public_key": "ssh-rsa invalid",
	"labels": {},
	"created": "2016-01-30T23:50:00+00:00"
}
			`))
		}

		res.Write([]byte(`
	]
}
		`))
	})
}
