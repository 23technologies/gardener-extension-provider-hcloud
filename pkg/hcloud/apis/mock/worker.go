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
	"regexp"
	"strconv"
	"strings"

	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	TestWorkerInfrastructureProviderStatus = `{
		"apiVersion": "hcloud.provider.extensions.gardener.cloud/v1alpha1",
		"kind": "InfrastructureStatus",
		"floatingPoolName": "MY-FLOATING-POOL"
	}`
	TestWorkerMachineImageName = "ubuntu"
	TestWorkerMachineImageVersion = "20.04"
	TestWorkerMachineType = "cx11"
	TestWorkerName = "hcloud"
	TestWorkerPoolName = "hcloud-pool-1"
	TestWorkerSecretName = "secret"
	TestWorkerUserData = "IyEvYmluL2Jhc2gKCmVjaG8gImhlbGxvIHdvcmxkIgo="
)

// NewWorker generates a new provider specification for testing purposes.
func NewWorker() *v1alpha1.Worker {
	return &v1alpha1.Worker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestWorkerName,
			Namespace: TestNamespace,
		},
		Spec: v1alpha1.WorkerSpec{
			SecretRef: corev1.SecretReference{
				Name:      TestWorkerSecretName,
				Namespace: TestNamespace,
			},
			Region: TestRegion,
			InfrastructureProviderStatus: &runtime.RawExtension{
				Raw: []byte(TestWorkerInfrastructureProviderStatus),
			},
			Pools: []v1alpha1.WorkerPool{
				{
					Name:           TestWorkerPoolName,
					Minimum:        5,
					Maximum:        10,
					MaxSurge:       intstr.FromInt(3),
					MaxUnavailable: intstr.FromInt(2),
					MachineType:    TestWorkerMachineType,
					MachineImage: v1alpha1.MachineImage{
						Name:    TestWorkerMachineImageName,
						Version: TestWorkerMachineImageVersion,
					},
					UserData: []byte(TestWorkerUserData),
					Zones: []string{
						TestZone,
					},
				},
			},
			SSHPublicKey: []byte(TestSSHPublicKey),
		},
	}
}

// ManipulateWorker changes given provider specification.
//
// PARAMETERS
// Worker *v1alpha1.Worker      Worker specification
// data    map[string]interface{} Members to change
func ManipulateWorker(worker *v1alpha1.Worker, data map[string]interface{}) *v1alpha1.Worker {
	reSpecPools := regexp.MustCompile(`^Spec\.Pools\.(\d+)\.`)

	for key, value := range data {
		if (strings.Index(key, "ObjectMeta") == 0) {
			manipulateStruct(&worker.ObjectMeta, key[11:], value)
		} else if (reSpecPools.MatchString(key)) {
			keyData := strings.SplitN(key, ".", 4)
			index, _ := strconv.Atoi(keyData[2])

			manipulateStruct(&worker.Spec.Pools[index], keyData[3], value)
		} else if (strings.Index(key, "Spec.Pools.") == 0) {
			manipulateStruct(&worker.Spec, key[7:], value)
		} else if (strings.Index(key, "Spec") == 0) {
			manipulateStruct(&worker.Spec, key[7:], value)
		} else {
			manipulateStruct(&worker, key, value)
		}
	}

	return worker
}

// SetupImagesEndpointOnMux configures a "/images" endpoint on the mux given.
//
// PARAMETERS
// mux *http.ServeMux Mux to add handler to
func SetupImagesEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/images", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")

		res.WriteHeader(http.StatusOK)

		res.Write([]byte(`
{
	"images": []
}
		`))
	})
}
