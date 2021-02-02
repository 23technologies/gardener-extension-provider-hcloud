/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package hcloud

import (
	"path/filepath"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

const (
	// Name is the name of the Hcloud provider controller.
	Name = "provider-hcloud"

	// MachineControllerManagerImageName is the name of the MachineControllerManager image.
	MachineControllerManagerImageName = "machine-controller-manager"
	// MCMProviderHcloudImageName is the namne of the HCloud provider plugin image.
	MCMProviderHcloudImageName = "machine-controller-manager-provider-hcloud"
	// CloudControllerImageName is the name of the external HCloud CloudProvider image.
	CloudControllerImageName = "hcloud-cloud-controller-manager"

	// CSIAttacherImageName is the name of the CSI attacher image.
	CSIAttacherImageName = "csi-attacher"
	// CSINodeDriverRegistrarImageName is the name of the CSI driver registrar image.
	CSINodeDriverRegistrarImageName = "csi-node-driver-registrar"
	// CSIProvisionerImageName is the name of the CSI provisioner image.
	CSIProvisionerImageName = "csi-provisioner"
	// CSIDriverControllerImageName is the name of the CSI driver controller plugin image.
	CSIDriverControllerImageName = "hcloud-csi-driver-controller"
	// CSIDriverNodeImageName is the name of the CSI driver node plugin image.
	CSIDriverNodeImageName = "hcloud-csi-driver-node"
	// CSIDriverSyncerImageName is the name of the HCloud CSI Syncer image.
	CSIDriverSyncerImageName = "hcloud-csi-driver-syncer"
	// CSIResizerImageName is the name of the csi-resizer image.
	CSIResizerImageName = "csi-resizer"
	// LivenessProbeImageName is the name of the liveness-probe image.
	LivenessProbeImageName = "liveness-probe"

	HcloudToken    = "hcloudToken"
	HcloudTokenMCM = "hcloudTokenMCM"
	HcloudTokenCCM = "hcloudTokenCCM"
	HcloudTokenCSI = "hcloudTokenCSI"
	// Host is a constant for the key in a cloud provider secret holding the HCloud host name
	// Host = "hcloudHost"
	// Username is a constant for the key in a cloud provider secret holding the HCloud user name (optional, for all components)
	// Username = "hcloudUsername"
	// Password is a constant for the key in a cloud provider secret holding the HCloud password (optional, for all components)
	// Password = "hcloudPassword"
	// 	// Username is a constant for the key in a cloud provider secret holding the HCloud user name (specific for MachineControllerManager)
	// 	UsernameMCM = "hcloudUsernameMCM"
	// 	// Password is a constant for the key in a cloud provider secret holding the HCloud password (specific for MachineControllerManager)
	// 	PasswordMCM = "hcloudPasswordMCM"
	// 	// Username is a constant for the key in a cloud provider secret holding the HCloud user name (specific for CloudControllerManager)
	// 	UsernameCCM = "hcloudUsernameCCM"
	// 	// Password is a constant for the key in a cloud provider secret holding the HCloud password (specific for CloudControllerManager)
	// 	PasswordCCM = "hcloudPasswordCCM"
	// 	// Username is a constant for the key in a cloud provider secret holding the HCloud user name (specific for CSI)
	// 	UsernameCSI = "hcloudUsernameCSI"
	// 	// Password is a constant for the key in a cloud provider secret holding the HCloud password (specific for CSI)
	// 	PasswordCSI = "hcloudPasswordCSI"
	// 	// InsecureSSL is a constant for the key in a cloud provider secret holding the boolean flag to allow insecure HTTPS connections to the HCloud host
	// InsecureSSL = "hcloudInsecureSSL"

	// 	// NSXTUsername is a constant for the key in a cloud provider secret holding the NSX-T user name with role 'Enterprise Admin' (optional, for all components)
	// 	NSXTUsername = "nsxtUsername"
	// 	// Password is a constant for the key in a cloud provider secret holding the NSX-T password for user with role 'Enterprise Admin'
	// 	NSXTPassword = "nsxtPassword"
	// 	// NSXTUsernameLBAdmin is a constant for the key in a cloud provider secret holding the NSX-T user name with role 'LB Admin' (needed for CloudControllerManager)
	// 	NSXTUsernameLBAdmin = "nsxtUsernameLBAdmin"
	// 	// NSXTPasswordLBAdmin is a constant for the key in a cloud provider secret holding the NSX-T password for user with role 'LB Admin'
	// 	NSXTPasswordLBAdmin = "nsxtPasswordLBAdmin"
	// 	// NSXTUsernameNE is a constant for the key in a cloud provider secret holding the NSX-T user name with role 'Network Engineer' (needed for infrastructure and IP pools address allocation in CloudControllerManager)
	// 	NSXTUsernameNE = "nsxtUsernameNE"
	// 	// NSXTPasswordNE is a constant for the key in a cloud provider secret holding the NSX-T password for user with role 'Network Engineer'
	// 	NSXTPasswordNE = "nsxtPasswordNE"

	// 	// CloudProviderConfig is the name of the configmap containing the cloud provider config.
	CloudProviderConfig = "cloud-provider-config"
	// 	// CloudProviderConfigMapKey is the key storing the cloud provider config as value in the cloud provider configmap.
	CloudProviderConfigMapKey = "cloudprovider.conf"
	// 	// SecretCSIHcloudConfig is a constant for the secret containing the CSI HCloud config.
	SecretCSIHcloudConfig = "csi-hcloud-config"
	// 	// MachineControllerManagerName is a constant for the name of the machine-controller-manager.
	MachineControllerManagerName = "machine-controller-manager"
	// 	// MachineControllerManagerVpaName is the name of the VerticalPodAutoscaler of the machine-controller-manager deployment.
	MachineControllerManagerVpaName = "machine-controller-manager-vpa"
	// 	// MachineControllerManagerMonitoringConfigName is the name of the ConfigMap containing monitoring stack configurations for machine-controller-manager.
	MachineControllerManagerMonitoringConfigName = "machine-controller-manager-monitoring-config"

	// 	// CloudControllerManagerName is the constant for the name of the CloudController deployed by the control plane controller.
	CloudControllerManagerName = "cloud-controller-manager"

	// CloudControllerManagerServerName is the constant for the name of the CloudController deployed by the control plane controller.
	CloudControllerManagerServerName = "cloud-controller-manager-server"
	// CSIProvisionerName is a constant for the name of the csi-provisioner component.
	CSIProvisionerName = "csi-provisioner"
	// CSIAttacherName is a constant for the name of the csi-attacher component.
	CSIAttacherName = "csi-attacher"
	// CSIResizerName is a constant for the name of the csi-resizer component.
	CSIResizerName = "csi-resizer"
	// HcloudCSIController is a constant for the name of the hcloud-csi-controller component.
	HcloudCSIController = "hcloud-csi-controller"
	// HcloudCSISyncer is a constant for the name of the hcloud-csi-syncer component.
	HcloudCSISyncer = "csi-syncer"
	// CSINodeName is a constant for the chart name for a CSI node deployment in the shoot.
	CSINodeName = "hcloud-csi-node"
	// CSIDriverName is a constant for the name of the csi-driver component.
	CSIDriverName = "csi-driver"
)

var (
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")

	// UsernamePrefix is a constant for the username prefix of components deployed by OpenStack.
	UsernamePrefix = extensionsv1alpha1.SchemeGroupVersion.Group + ":" + Name + ":"
)
