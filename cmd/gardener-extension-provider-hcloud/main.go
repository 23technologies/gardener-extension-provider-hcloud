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

// Package main provides the application's entry point
package main

import (
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/cmd/options"
	"github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/log"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	runtimelog.SetLogger(log.ZapLogger(false))
	cmdDefinition := options.NewControllerManagerCommand(controller.SetupSignalHandlerContext())

	if err := cmdDefinition.Execute(); err != nil {
		cmd.LogErrAndExit(err, "error executing the main controller command")
	}
}
