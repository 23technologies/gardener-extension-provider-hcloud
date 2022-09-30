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
	"fmt"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config/loader"
	extensionconfig "github.com/gardener/gardener/extensions/pkg/apis/config"
	"github.com/spf13/pflag"
)

// ConfigOptions are command line options that can be set for config.ControllerConfiguration.
type ConfigOptions struct {
	// Kubeconfig is the path to a kubeconfig.
	ConfigFilePath string

	config *Config
}

// Config is a completed controller configuration.
type Config struct {
	// Config is the controller configuration.
	Config *config.ControllerConfiguration
}

// buildConfig loads the controller configuration from the configured file.
func (c *ConfigOptions) buildConfig() (*config.ControllerConfiguration, error) {
	if len(c.ConfigFilePath) == 0 {
		return nil, fmt.Errorf("config file path not set")
	}
	return loader.LoadFromFile(c.ConfigFilePath)
}

// Complete implements RESTCompleter.Complete.
func (c *ConfigOptions) Complete() error {
	config, err := c.buildConfig()
	if err != nil {
		return err
	}

	c.config = &Config{config}
	return nil
}

// Completed returns the completed Config. Only call this if `Complete` was successful.
func (c *ConfigOptions) Completed() *Config {
	return c.config
}

// AddFlags implements Flagger.AddFlags.
//
// PARAMETERS
// fs *pflag.FlagSet Flags to recognize
func (c *ConfigOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.ConfigFilePath, "config-file", "", "path to the controller manager configuration file")
}

// Apply sets the values of this Config in the given config.ControllerConfiguration.
//
// PARAMETERS
// cfg *config.ControllerConfiguration Pointer to the configuration to set
func (c *Config) Apply(cfg *config.ControllerConfiguration) {
	*cfg = *c.Config
}

// ApplyETCDStorage sets the given etcd storage configuration to that of this Config.
//
// PARAMETERS
// etcdStorage *config.ETCDStorage Pointer to the etcd storage configuration to set
func (c *Config) ApplyETCDStorage(etcdStorage *config.ETCDStorage) {
	*etcdStorage = *c.Config.ETCD.Storage
}

// ApplyGardenId sets the gardenId.
//
// PARAMETERS
// gardenId *string Pointer to the gardenId to set
func (c *Config) ApplyGardenId(gardenId *string) {
	*gardenId = c.Config.GardenId
}

// ApplyHealthProbeBindAddress sets the healthProbeBindAddress.
//
// PARAMETERS
// healthProbeBindAddress *string Pointer to the healthProbeBindAddress to set
func (c *Config) ApplyHealthProbeBindAddress(healthProbeBindAddress *string) {
	*healthProbeBindAddress = c.Config.HealthProbeBindAddress
}

// ApplyMetricsBindAddress sets the metricsBindAddress.
//
// PARAMETERS
// metricsBindAddress *string Pointer to the metricsBindAddress to set
func (c *Config) ApplyMetricsBindAddress(metricsBindAddress *string) {
	*metricsBindAddress = c.Config.MetricsBindAddress
}

// Options initializes empty config.ControllerConfiguration, applies the set values and returns it.
func (c *Config) Options() config.ControllerConfiguration {
	var cfg config.ControllerConfiguration
	c.Apply(&cfg)
	return cfg
}

// ApplyHealthCheckConfig applies the HealthCheckConfig to the config
//
// PARAMETERS
// config *healthcheckconfig.HealthCheckConfig Pointer to the HealthCheckConfig to set
func (c *Config) ApplyHealthCheckConfig(config *extensionconfig.HealthCheckConfig) {
	if c.Config.HealthCheckConfig != nil {
		*config = *c.Config.HealthCheckConfig
	}
}
