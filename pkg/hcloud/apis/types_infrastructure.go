// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta
	// Networks contains optional existing network infrastructure to use.
	// If not defined, NSX-T Tier-1 gateway and load balancer are created for the shoot cluster.
	// Networks *Networks
}

// Networks contains existing NSX-T network infrastructure to use.
type Networks struct {
	// Tier1GatewayPath is the path of the existing NSX-T Tier-1 Gateway to use.
	Tier1GatewayPath string
	// LoadBalancerServicePath is the path of the existing NSX-T load balancer service assigned to the Tier-1 Gateway
	LoadBalancerServicePath string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta

	CreationStarted *bool
}
