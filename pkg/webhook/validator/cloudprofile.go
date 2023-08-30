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

package validator

import (
	"context"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewCloudProfileValidator returns a new instance of a cloud profile validator.
func NewCloudProfileValidator() extensionswebhook.Validator {
	return &cloudProfile{}
}

type cloudProfile struct {
	decoder runtime.Decoder
}

// Validate validates the given cloud profile objects.
func (cp *cloudProfile) Validate(_ context.Context, new, _ client.Object) error {
	cloudProfile, ok := new.(*core.CloudProfile)
	if !ok {
		return fmt.Errorf("wrong object type %T", new)
	}

	for _, region := range cloudProfile.Spec.Regions {
		if len(region.Zones) > 1 {
			return fmt.Errorf("This version of the hcloud extension does not support multiple zones per region. Consider implementing this feature.")
		}
	}

	return nil

}
