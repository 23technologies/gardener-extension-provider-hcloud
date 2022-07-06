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

// Package transcoder is used for API related object transformations
package transcoder

import (
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func DecodeInfrastructureStatusFromWorker(worker *v1alpha1.Worker) (*apis.InfrastructureStatus, error) {
	infraStatus, err := DecodeInfrastructureStatus(worker.Spec.InfrastructureProviderStatus)
	if err != nil {
		return nil, err
	}

	return infraStatus, nil
}

func DecodeWorkerStatus(status *runtime.RawExtension) (*apis.WorkerStatus, error) {
	providerStatus := &apis.WorkerStatus{}

	if status == nil {
		return providerStatus, nil
	}

	if _, _, err := decoder.Decode(status.Raw, nil, providerStatus); err != nil {
		return nil, fmt.Errorf("could not decode workerStatus: %w")
	}

	return providerStatus, nil
}

func DecodeWorkerStatusFromWorker(worker *v1alpha1.Worker) (*apis.WorkerStatus, error) {
	providerStatus, err := DecodeWorkerStatus(worker.Status.ProviderStatus)
	if err != nil {
		return nil, err
	}

	return providerStatus, nil
}

// WorkerConfigFromRawExtension extracts the provider specific configuration for a worker pool.
func DecodeWorkerConfigFromRawExtension(raw *runtime.RawExtension) (*apis.WorkerConfig, error) {
	poolConfig := &apis.WorkerConfig{}

	if raw != nil {
		marshalled, err := raw.MarshalJSON()
		if err != nil {
			return nil, err
		}

		if _, _, err := decoder.Decode(marshalled, nil, poolConfig); err != nil {
			return nil, err
		}
	}

	return poolConfig, nil
}
