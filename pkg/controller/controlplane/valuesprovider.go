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

// Package controlplane contains functions used at the controlplane controller
package controlplane

import (
	"context"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	k8sutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
)

var controlPlaneSecrets = &secrets.Secrets{
	CertificateSecretConfigs: map[string]*secrets.CertificateSecretConfig{
		v1beta1constants.SecretNameCACluster: {
			Name:       v1beta1constants.SecretNameCACluster,
			CommonName: "kubernetes",
			CertType:   secrets.CACert,
		},
	},
	SecretConfigsFunc: func(cas map[string]*secrets.Certificate, clusterName string) []secrets.ConfigInterface {
		return []secrets.ConfigInterface{
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         hcloud.CloudControllerManagerName,
					CommonName:   "system:cloud-controller-manager",
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequests: []secrets.KubeConfigRequest{
					{
						ClusterName:  clusterName,
						APIServerHost: v1beta1constants.DeploymentNameKubeAPIServer,
					},
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:       hcloud.CloudControllerManagerServerName,
					CommonName: hcloud.CloudControllerManagerName,
					DNSNames:   k8sutils.DNSNamesForService(hcloud.CloudControllerManagerName, clusterName),
					CertType:   secrets.ServerCert,
					SigningCA:  cas[v1beta1constants.SecretNameCACluster],
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         hcloud.CSIAttacherName,
					CommonName:   hcloud.UsernamePrefix + hcloud.CSIAttacherName,
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequests: []secrets.KubeConfigRequest{
					{
						ClusterName:  clusterName,
						APIServerHost: v1beta1constants.DeploymentNameKubeAPIServer,
					},
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         hcloud.CSIProvisionerName,
					CommonName:   hcloud.UsernamePrefix + hcloud.CSIProvisionerName,
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequests: []secrets.KubeConfigRequest{
					{
						ClusterName:  clusterName,
						APIServerHost: v1beta1constants.DeploymentNameKubeAPIServer,
					},
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         hcloud.HcloudCSIController,
					CommonName:   hcloud.UsernamePrefix + hcloud.HcloudCSIController,
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequests: []secrets.KubeConfigRequest{
					{
						ClusterName:  clusterName,
						APIServerHost: v1beta1constants.DeploymentNameKubeAPIServer,
					},
				},
			},
			&secrets.ControlPlaneSecretConfig{
				CertificateSecretConfig: &secrets.CertificateSecretConfig{
					Name:         hcloud.CSIResizerName,
					CommonName:   hcloud.UsernamePrefix + hcloud.CSIResizerName,
					Organization: []string{user.SystemPrivilegedGroup},
					CertType:     secrets.ClientCert,
					SigningCA:    cas[v1beta1constants.SecretNameCACluster],
				},
				KubeConfigRequests: []secrets.KubeConfigRequest{
					{
						ClusterName:  clusterName,
						APIServerHost: v1beta1constants.DeploymentNameKubeAPIServer,
					},
				},
			},
		}
	},
}

var configChart = &chart.Chart{
	Name: "cloud-provider-config",
	Path: filepath.Join(hcloud.InternalChartsPath, "cloud-provider-config"),
	Objects: []*chart.Object{
		{Type: &corev1.ConfigMap{}, Name: hcloud.CloudProviderConfig},
	},
}

var controlPlaneChart = &chart.Chart{
	Name: "seed-controlplane",
	Path: filepath.Join(hcloud.InternalChartsPath, "seed-controlplane"),
	SubCharts: []*chart.Chart{
		{
			Name:   "hcloud-cloud-controller-manager",
			Images: []string{hcloud.CloudControllerImageName},
			Objects: []*chart.Object{
				{Type: &corev1.Service{}, Name: hcloud.CloudControllerManagerName},
				{Type: &appsv1.Deployment{}, Name: hcloud.CloudControllerManagerName},
				{Type: &corev1.ConfigMap{}, Name: hcloud.CloudControllerManagerName + "-observability-config"},
				{Type: &autoscalingv1beta2.VerticalPodAutoscaler{}, Name: hcloud.CloudControllerManagerName + "-vpa"},
			},
		},
		{
			Name: "csi-hcloud",
			Images: []string{
				hcloud.CSIAttacherImageName,
				hcloud.CSIProvisionerImageName,
				hcloud.CSIDriverControllerImageName,
				hcloud.CSIResizerImageName,
				hcloud.LivenessProbeImageName},
			Objects: []*chart.Object{
				{Type: &appsv1.Deployment{}, Name: hcloud.HcloudCSIController},
				{Type: &corev1.ConfigMap{}, Name: hcloud.HcloudCSIController + "-observability-config"},
				{Type: &autoscalingv1beta2.VerticalPodAutoscaler{}, Name: hcloud.HcloudCSIController + "-vpa"},
			},
		},
	},
}

var controlPlaneShootChart = &chart.Chart{
	Name: "shoot-system-components",
	Path: filepath.Join(hcloud.InternalChartsPath, "shoot-system-components"),
	SubCharts: []*chart.Chart{
		{
			Name: "hcloud-cloud-controller-manager",
			Objects: []*chart.Object{
				{Type: &rbacv1.ClusterRole{}, Name: "system:cloud-controller-manager"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:cloud-controller-manager"},
				{Type: &rbacv1.ClusterRole{}, Name: "system:controller:cloud-node-controller"},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:controller:cloud-node-controller"},
			},
		},
		{
			Name: "csi-hcloud",
			Images: []string{
				hcloud.CSINodeDriverRegistrarImageName,
				hcloud.CSIDriverNodeImageName,
				hcloud.LivenessProbeImageName,
			},
			Objects: []*chart.Object{
				// csi-driver
				{Type: &appsv1.DaemonSet{}, Name: hcloud.CSINodeName},
				//{Type: &storagev1beta1.CSIDriver{}, Name: "csi.hcloud.vmware.com"},
				{Type: &corev1.ServiceAccount{}, Name: hcloud.CSIDriverName + "-node"},
				{Type: &rbacv1.ClusterRole{}, Name: hcloud.UsernamePrefix + hcloud.CSIDriverName},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIDriverName},
				{Type: &policyv1beta1.PodSecurityPolicy{}, Name: strings.Replace(hcloud.UsernamePrefix+hcloud.CSIDriverName, ":", ".", -1)},
				// csi-provisioner
				{Type: &rbacv1.ClusterRole{}, Name: hcloud.UsernamePrefix + hcloud.CSIProvisionerName},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIProvisionerName},
				{Type: &rbacv1.Role{}, Name: hcloud.UsernamePrefix + hcloud.CSIProvisionerName},
				{Type: &rbacv1.RoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIProvisionerName},
				// csi-attacher
				{Type: &rbacv1.ClusterRole{}, Name: hcloud.UsernamePrefix + hcloud.CSIAttacherName},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIAttacherName},
				{Type: &rbacv1.Role{}, Name: hcloud.UsernamePrefix + hcloud.CSIAttacherName},
				{Type: &rbacv1.RoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIAttacherName},
				// csi-resizer
				{Type: &rbacv1.ClusterRole{}, Name: hcloud.UsernamePrefix + hcloud.CSIResizerName},
				{Type: &rbacv1.ClusterRoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIResizerName},
				{Type: &rbacv1.Role{}, Name: hcloud.UsernamePrefix + hcloud.CSIResizerName},
				{Type: &rbacv1.RoleBinding{}, Name: hcloud.UsernamePrefix + hcloud.CSIResizerName},
			},
		},
	},
}

var storageClassChart = &chart.Chart{
	Name: "shoot-storageclasses",
	Path: filepath.Join(hcloud.InternalChartsPath, "shoot-storageclasses"),
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
func NewValuesProvider(logger logr.Logger, gardenID string) genericactuator.ValuesProvider {
	return &valuesProvider{
		logger:   logger.WithName("hcloud-values-provider"),
		gardenID: gardenID,
	}
}

// valuesProvider is a ValuesProvider that provides hcloud-specific values for the 2 charts applied by the generic actuator.
type valuesProvider struct {
	genericactuator.NoopValuesProvider
	common.ClientContext
	logger   logr.Logger
	gardenID string
}

// GetConfigChartValues returns the values for the config chart applied by the generic actuator.
func (vp *valuesProvider) GetConfigChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	cpConfig, err := transcoder.DecodeControlPlaneConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	// Get credentials
	credentials, err := hcloud.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get hcloud credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get config chart values
	return vp.getConfigChartValues(cpConfig, cp, cluster, credentials)
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	cpConfig, err := transcoder.DecodeControlPlaneConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	// Decode infrastructureProviderStatus
	infraStatus, err := transcoder.DecodeInfrastructureStatusFromControlPlane(cp)
	if nil != err {
		return nil, errors.Wrapf(err, "could not decode infrastructureProviderStatus of controlplane '%s'", k8sutils.ObjectName(cp))
	}

	// Get credentials
	credentials, err := hcloud.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get hcloud credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get control plane chart values
	return vp.getControlPlaneChartValues(cpConfig, infraStatus, cp, cluster, credentials, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	checksums map[string]string,
) (map[string]interface{}, error) {
	// Get credentials
	credentials, err := hcloud.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get hcloud credentials from secret '%s/%s'", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name)
	}

	// Get control plane shoot chart values
	return vp.getControlPlaneShootChartValues(cp, cluster, credentials)
}

// GetStorageClassesChartValues returns the values for the shoot storageclasses chart applied by the generic actuator.
func (vp *valuesProvider) GetStorageClassesChartValues(
	_ context.Context,
	_ *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
) (map[string]interface{}, error) {
	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	volumeBindingMode := "Immediate"

	return map[string]interface{}{
		"storagePolicyName":    cloudProfileConfig.DefaultClassStoragePolicyName,
		"volumeBindingMode":    volumeBindingMode,
		"allowVolumeExpansion": true,
	}, nil
}

func splitServerNameAndPort(host string) (name string, port int, err error) {
	parts := strings.Split(host, ":")

	if len(parts) == 1 {
		name = host
		port = 443
	} else if len(parts) == 2 {
		name = parts[0]
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, errors.Wrapf(err, "invalid port for hcloud host: host=%s,port=%s", host, parts[1])
		}
	} else {
		return "", 0, fmt.Errorf("invalid hcloud host: %s (too many parts %v)", host, parts)
	}

	return
}

// getConfigChartValues collects and returns the configuration chart values.
func (vp *valuesProvider) getConfigChartValues(
	cpConfig *apis.ControlPlaneConfig,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
) (map[string]interface{}, error) {
	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	region := apis.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	// Collect config chart values
	values := map[string]interface{}{
		"token": credentials.HcloudCCM().Token,
	}

	return values, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
func (vp *valuesProvider) getControlPlaneChartValues(
	cpConfig *apis.ControlPlaneConfig,
	infraStatus *apis.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	region := apis.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	clusterID, csiClusterID := vp.calcClusterIDs(cp)
	// csiResizerEnabled := cloudProfileConfig.CSIResizerDisabled == nil || !*cloudProfileConfig.CSIResizerDisabled
	values := map[string]interface{}{
		"hcloud-cloud-controller-manager": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"clusterName":       clusterID,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-" + hcloud.CloudControllerManagerName:        checksums[hcloud.CloudControllerManagerName],
				"checksum/secret-" + hcloud.CloudControllerManagerServerName:  checksums[hcloud.CloudControllerManagerServerName],
				"checksum/secret-" + v1beta1constants.SecretNameCloudProvider: checksums[v1beta1constants.SecretNameCloudProvider],
				"checksum/configmap-" + hcloud.CloudProviderConfig:            checksums[hcloud.CloudProviderConfig],
			},
			"podLabels": map[string]interface{}{
				v1beta1constants.LabelPodMaintenanceRestart: "true",
			},
			"podLocation": apis.GetLocationFromRegion(region.Name),
		},
		"csi-hcloud": map[string]interface{}{
			"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
			"clusterID":         csiClusterID,
			"token":             credentials.HcloudCSI().Token,
			// "resizerEnabled":    csiResizerEnabled,
			"podAnnotations": map[string]interface{}{
				"checksum/secret-" + hcloud.CSIProvisionerName:                checksums[hcloud.CSIProvisionerName],
				"checksum/secret-" + hcloud.CSIAttacherName:                   checksums[hcloud.CSIAttacherName],
				"checksum/secret-" + hcloud.CSIResizerName:                    checksums[hcloud.CSIResizerName],
				"checksum/secret-" + hcloud.HcloudCSIController:               checksums[hcloud.HcloudCSIController],
				"checksum/secret-" + v1beta1constants.SecretNameCloudProvider: checksums[v1beta1constants.SecretNameCloudProvider],
			},
		},
	}

	if cpConfig.CloudControllerManager != nil {
		values["hcloud-cloud-controller-manager"].(map[string]interface{})["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	if infraStatus.NetworkIDs != nil && infraStatus.NetworkIDs.Workers != -1 {
		values["hcloud-cloud-controller-manager"].(map[string]interface{})["podNetworkIDs"] = map[string]interface{}{
			"workers": infraStatus.NetworkIDs.Workers,
		}
	}

	return values, nil
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
func (vp *valuesProvider) getControlPlaneShootChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
) (map[string]interface{}, error) {

	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromControllerCluster(cluster)
	if err != nil {
		return nil, err
	}

	region := apis.FindRegion(cluster.Shoot.Spec.Region, cloudProfileConfig)
	if region == nil {
		return nil, fmt.Errorf("region %q not found in cloud profile config", cluster.Shoot.Spec.Region)
	}

	_, csiClusterID := vp.calcClusterIDs(cp)
	values := map[string]interface{}{
		"csi-hcloud": map[string]interface{}{
			// "serverName":  serverName,
			"clusterID":         csiClusterID,
			"token":             credentials.HcloudCSI().Token,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		},
	}

	return values, nil
}

func (vp *valuesProvider) calcClusterIDs(cp *extensionsv1alpha1.ControlPlane) (clusterID string, csiClusterID string) {
	clusterID = cp.Namespace + "-" + vp.gardenID
	csiClusterID = shortenID(clusterID, 63)
	return
}

func shortenID(id string, maxlen int) string {
	if maxlen < 16 {
		panic("maxlen < 16 for shortenID")
	}
	if len(id) <= maxlen {
		return id
	}

	hash := fnv.New64()
	_, _ = hash.Write([]byte(id))
	hashstr := strconv.FormatUint(hash.Sum64(), 36)
	return fmt.Sprintf("%s-%s", id[:62-len(hashstr)], hashstr)
}
