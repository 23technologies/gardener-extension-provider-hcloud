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
	"errors"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	errorhelpers "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func DecodeInfrastructureStatus(infra *runtime.RawExtension) (*apis.InfrastructureStatus, error) {
	infraStatus := &apis.InfrastructureStatus{}

	if infra == nil || infra.Raw == nil {
		return nil, errors.New("Missing infrastructure status")
	}

	if _, _, err := decoder.Decode(infra.Raw, nil, infraStatus); err != nil {
		return nil, errorhelpers.Wrapf(err, "could not decode infrastructureStatus")
	}

	return infraStatus, nil
}

func DecodeInfrastructureStatusFromWorker(worker *v1alpha1.Worker) (*apis.InfrastructureStatus, error) {
	infraStatus, err := DecodeInfrastructureStatus(worker.Spec.InfrastructureProviderStatus)
	if err != nil {
		return nil, err
	}

	return infraStatus, nil
}
