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
	"errors"
	"fmt"
	"reflect"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/validation"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewShootValidator returns a new instance of a shoot validator.
func NewShootValidator() extensionswebhook.Validator {
	return &shoot{}
}

type shoot struct {
	client         client.Client
	decoder        runtime.Decoder
	lenientDecoder runtime.Decoder
}

// InjectClient injects the given client into the validator.
func (s *shoot) InjectClient(client client.Client) error {
	s.client = client
	return nil
}

// Validate validates the given shoot object.
func (s *shoot) Validate(ctx context.Context, new, old client.Object) error {
	shoot, ok := new.(*core.Shoot)
	if !ok {
		return fmt.Errorf("wrong object type %T", new)
	}

	if old != nil {
		oldShoot, ok := old.(*core.Shoot)
		if !ok {
			return fmt.Errorf("wrong object type %T for old object", old)
		}
		return s.validateShootUpdate(ctx, oldShoot, shoot)
	}

	return s.validateShootCreation(ctx, shoot)
}

func (s *shoot) validateShoot(_ context.Context, shoot *core.Shoot) error {
	// Network validation
	if errList := validation.ValidateShootNetworking(*shoot.Spec.Networking); len(errList) != 0 {
		return errList.ToAggregate()
	}

	// Provider validation
	fldPath := field.NewPath("spec", "provider")

	infraConfig, err := transcoder.DecodeInfrastructureConfig(shoot.Spec.Provider.InfrastructureConfig)
	if err != nil {
		return field.InternalError(fldPath.Child("infrastructureConfig"), err)
	}

	if errList := validation.ValidateInfrastructureConfig(infraConfig, shoot.Spec.Networking.Nodes, shoot.Spec.Networking.Pods, shoot.Spec.Networking.Services); len(errList) != 0 {
		return errList.ToAggregate()
	}

	// ControlPlaneConfig
	if shoot.Spec.Provider.ControlPlaneConfig != nil {
		if _, err := transcoder.DecodeControlPlaneConfigWithDecoder(s.decoder, shoot.Spec.Provider.ControlPlaneConfig); err != nil {
			return err
		}
	}

	// WorkerConfig and Shoot workers
	if errList := validation.ValidateWorkers(shoot.Spec.Provider.Workers, fldPath.Child("workers")); len(errList) != 0 {
		return errList.ToAggregate()
	}

	return nil
}

func (s *shoot) validateShootUpdate(ctx context.Context, oldShoot, shoot *core.Shoot) error {
	var (
		fldPath            = field.NewPath("spec", "provider")
		infraConfigFldPath = fldPath.Child("infrastructureConfig")
	)

	infraConfig, err := transcoder.DecodeInfrastructureConfig(shoot.Spec.Provider.InfrastructureConfig)
	if err != nil {
		return field.InternalError(infraConfigFldPath, err)
	}

	if oldShoot.Spec.Provider.InfrastructureConfig == nil {
		return field.InternalError(infraConfigFldPath, errors.New("InfrastructureConfig is not available on old shoot"))
	}

	oldInfraConfig, err := transcoder.DecodeInfrastructureConfig(oldShoot.Spec.Provider.InfrastructureConfig)
	if err != nil {
		return field.InternalError(infraConfigFldPath, err)
	}

	if !reflect.DeepEqual(oldInfraConfig, infraConfig) {
		if errList := validation.ValidateInfrastructureConfigUpdate(oldInfraConfig, infraConfig); len(errList) != 0 {
			return errList.ToAggregate()
		}
	}

	if err := s.validateAgainstCloudProfile(ctx, shoot, oldInfraConfig, infraConfig, infraConfigFldPath); err != nil {
		return err
	}

	if errList := validation.ValidateWorkersUpdate(oldShoot.Spec.Provider.Workers, shoot.Spec.Provider.Workers, fldPath.Child("workers")); len(errList) != 0 {
		return errList.ToAggregate()
	}

	return s.validateShoot(ctx, shoot)
}

func (s *shoot) validateShootCreation(ctx context.Context, shoot *core.Shoot) error {
	fldPath := field.NewPath("spec", "provider", "infrastructureConfig")

	infraConfig, err := transcoder.DecodeInfrastructureConfig(shoot.Spec.Provider.InfrastructureConfig)
	if err != nil {
		return field.InternalError(fldPath, err)
	}

	if err := s.validateAgainstCloudProfile(ctx, shoot, nil, infraConfig, fldPath); err != nil {
		return err
	}

	return s.validateShoot(ctx, shoot)
}

func (s *shoot) validateAgainstCloudProfile(ctx context.Context, shoot *core.Shoot, oldInfraConfig, infraConfig *apis.InfrastructureConfig, fldPath *field.Path) error {
	cloudProfile := &gardencorev1beta1.CloudProfile{}
	if err := s.client.Get(ctx, kutil.Key(shoot.Spec.CloudProfileName), cloudProfile); err != nil {
		return err
	}

	if errList := validation.ValidateInfrastructureConfigAgainstCloudProfile(oldInfraConfig, infraConfig, shoot, cloudProfile, fldPath); len(errList) != 0 {
		return errList.ToAggregate()
	}

	return nil
}
