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

// Package apis is the main package for HCloud specific APIs
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta `json:",inline"`
	// FloatingPoolName contains the FloatingPoolName name in which LoadBalancer FIPs should be created.
	FloatingPoolName string `json:"floatingPoolName"`
	// Networks is the OpenStack specific network configuration
	Networks *Networks `json:"networks"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta `json:",inline"`

	// FloatingPoolName contains the FloatingPoolName name in which LoadBalancer FIPs should be created.
	// +optional
	FloatingPoolName string `json:"floatingPoolName,omitempty"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// Workers is a CIDRs of a worker subnet (private) to create (used for the VMs).
	Workers string `json:"workers"`
}
