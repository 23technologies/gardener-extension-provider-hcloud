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

// Package loader contains functions used for reading hcloud.provider.extensions.config.gardener.cloud
package loader

import (
	"io/ioutil"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/config/install"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
)

var (
	codec  runtime.Codec
	scheme *runtime.Scheme
)

// init is called by Go once.
func init() {
	scheme = runtime.NewScheme()
	install.Install(scheme)
	yamlSerializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	codec = versioning.NewDefaultingCodecForScheme(
		scheme,
		yamlSerializer,
		yamlSerializer,
		schema.GroupVersion{Version: "v1alpha1"},
		runtime.InternalGroupVersioner,
	)
}

// LoadFromFile takes a filename and de-serializes the contents into ControllerConfiguration object.
//
// PARAMETERS
// filename string File path and name to load
func LoadFromFile(filename string) (*config.ControllerConfiguration, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Load(bytes)
}

// Load takes a byte slice and de-serializes the contents into ControllerConfiguration object.
// Encapsulates de-serialization without assuming the source is a file.
//
// PARAMETERS
// data []byte Data to decode and interprete
func Load(data []byte) (*config.ControllerConfiguration, error) {
	cfg := &config.ControllerConfiguration{}

	if len(data) == 0 {
		return cfg, nil
	}

	decoded, _, err := codec.Decode(data, &schema.GroupVersionKind{Version: "v1alpha1", Kind: "Config"}, cfg)
	if err != nil {
		return nil, err
	}

	return decoded.(*config.ControllerConfiguration), nil
}
