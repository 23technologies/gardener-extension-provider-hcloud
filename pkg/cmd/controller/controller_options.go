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

// Package controller provides Kubernetes controller configuration structures used for command execution
package controller

import (
	hcloudcontrolplane "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/controlplane"
	hcloudhealthcheck "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/healthcheck"
	hcloudinfrastructure "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/infrastructure"
	hcloudworker "github.com/23technologies/gardener-extension-provider-hcloud/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/extensions/pkg/controller/worker"
)

// controllerSwitchOptions are the cmd.SwitchOptions for the provider controllers.
func controllerSwitchOptions() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(controlplane.ControllerName, hcloudcontrolplane.AddToManager),
		cmd.Switch(infrastructure.ControllerName, hcloudinfrastructure.AddToManager),
		cmd.Switch(worker.ControllerName, hcloudworker.AddToManager),
		cmd.Switch(healthcheck.ControllerName, hcloudhealthcheck.AddToManager),
		cmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
