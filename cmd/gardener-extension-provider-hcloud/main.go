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
	"os"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/cmd/controller"
	"github.com/gardener/gardener/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// main is the executable entry point.
func main() {
	log.SetLogger(logger.ZapLogger(false))
	cmdDefinition := controller.NewControllerManagerCommand(signals.SetupSignalHandler())

	if err := cmdDefinition.Execute(); err != nil {
		log.Log.Error(err, "Error executing command")
		os.Exit(1)
	}
}
