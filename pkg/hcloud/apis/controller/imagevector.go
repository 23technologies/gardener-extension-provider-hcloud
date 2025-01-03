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

// Package controller provides functions to access controller specifications
package controller

import (
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/23technologies/gardener-extension-provider-hcloud/charts"
)

var imageVector imagevector.ImageVector

func init() {
	newImageVector, err := imagevector.Read([]byte(charts.ImagesYAML))
	runtime.Must(err)

	newImageVector, err = imagevector.WithEnvOverride(newImageVector, imagevector.OverrideEnv)
	runtime.Must(err)

	imageVector = newImageVector
}

// ImageVector is the image vector that contains all the needed images.
func ImageVector() imagevector.ImageVector {
	return imageVector
}
