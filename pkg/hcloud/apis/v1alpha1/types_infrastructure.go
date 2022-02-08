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

// Package v1alpha1 provides hcloud.provider.extensions.gardener.cloud/v1alpha1
package v1alpha1

import (
	"github.com/hetznercloud/hcloud-go/hcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta `json:",inline"`

	// FloatingPoolName contains the FloatingPoolName name in which LoadBalancer FIPs should be created.
	// +optional
	FloatingPoolName string `json:"floatingPoolName,omitempty"`
	// Networks is the HCloud specific network configuration
	// +optional
	Networks *InfrastructureConfigNetworks `json:"networks,omitempty"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type InfrastructureConfigNetworks struct {
	// WorkersNetwork is a struct of a worker subnet (private) configuration to create (used for the VMs).
	WorkersConfiguration *InfrastructureConfigNetwork `json:"workersConfiguration"`
	// Workers is a CIDRs of a worker subnet (private) to create (used for the VMs).
	Workers              string                       `json:"workers,omitempty"`
}

// InfrastructureConfig holds information about the Kubernetes and infrastructure network.
type InfrastructureConfigNetwork struct {
	// Workers is a CIDRs of a worker subnet (private) to create (used for the VMs).
	Cidr string             `json:"cidr"`
	// Workers is a CIDRs of a worker subnet (private) to create (used for the VMs).
	Zone hcloud.NetworkZone `json:"zone,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta `json:",inline"`
	// SSHFingerprint contains the SSH fingerprint.
	SSHFingerprint string `json:"sshFingerprint"`

	// PlacementGroupIDs contains the placement group IDs.
	PlacementGroupIDs []string `json:"placementGroupIDs,omitempty"`
	// PlacementGroupID contains the placement group ID.
	PlacementGroupID string `json:"placementGroupID,omitempty"`
	// FloatingPoolName contains the FloatingPoolName name in which LoadBalancer FIPs should be created.
	// +optional
	FloatingPoolName string `json:"floatingPoolName,omitempty"`
	// Networks is the HCloud specific network configuration
	// +optional
	NetworkIDs *InfrastructureConfigNetworkIDs `json:"networkIDs,omitempty"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type InfrastructureConfigNetworkIDs struct {
	// Workers is the HCloud network ID created.
	Workers string `json:"workers"`
}
