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

// Package apis is the main package for provider specific APIs
package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneConfig contains configuration settings for the control plane.
type ControlPlaneConfig struct {
	metav1.TypeMeta

	// CloudControllerManager contains configuration settings for the cloud-controller-manager.
	CloudControllerManager *CloudControllerManagerConfig
	// LoadBalancerClasses lists the load balancer classes to be used.
	LoadBalancerClasses []CPLoadBalancerClass
	// LoadBalancerSize can override the default of the NSX-T load balancer size ("SMALL", "MEDIUM", or "LARGE") defined in the cloud profile.
	LoadBalancerSize *string
}

// CloudControllerManagerConfig contains configuration settings for the cloud-controller-manager.
type CloudControllerManagerConfig struct {
	// FeatureGates contains information about enabled feature gates.
	FeatureGates map[string]bool
}

// CPLoadBalancerClass provides the name of a load balancer
type CPLoadBalancerClass struct {
	Name string
	// IPPoolName is the name of the NSX-T IP pool.
	IPPoolName *string
	// TCPAppProfileName is the profile name of the load balaner profile for TCP
	TCPAppProfileName *string
	// UDPAppProfileName is the profile name of the load balaner profile for UDP
	UDPAppProfileName *string
}
