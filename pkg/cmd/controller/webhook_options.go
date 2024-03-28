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
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	webhook "github.com/gardener/gardener/extensions/pkg/webhook/controlplane"

	hcloudwebhook "github.com/23technologies/gardener-extension-provider-hcloud/pkg/webhook/controlplane"
)

// webhookSwitchOptions are the webhookcmd.SwitchOptions for the provider webhooks.
func webhookSwitchOptions() *webhookcmd.SwitchOptions {
	return webhookcmd.NewSwitchOptions(
		webhookcmd.Switch(webhook.WebhookName, hcloudwebhook.AddToManager),
	)
}
