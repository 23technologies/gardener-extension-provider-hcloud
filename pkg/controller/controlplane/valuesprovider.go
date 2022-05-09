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
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/controlplane/genericactuator"
	extensionssecretsmanager "github.com/gardener/gardener/extensions/pkg/util/secret/manager"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/chart"
	gardenerutils "github.com/gardener/gardener/pkg/utils/gardener"
	k8sutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	secretutils "github.com/gardener/gardener/pkg/utils/secrets"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	autoscalingv1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
)

const (
	caNameControlPlane = "ca-" + hcloud.Name + "-controlplane"
)

var (
	configChart = &chart.Chart{
		Name: "cloud-provider-config",
		Path: filepath.Join(hcloud.InternalChartsPath, "cloud-provider-config"),
		Objects: []*chart.Object{
			{Type: &corev1.ConfigMap{}, Name: hcloud.CloudProviderConfig},
		},
	}

	controlPlaneChart = &chart.Chart{
		Name: "seed-controlplane",
		Path: filepath.Join(hcloud.InternalChartsPath, "seed-controlplane"),
		SubCharts: []*chart.Chart{
			{
				Name:   hcloud.CloudControllerManagerName,
				Images: []string{hcloud.CloudControllerImageName},
				Objects: []*chart.Object{
					{Type: &corev1.Service{}, Name: hcloud.CloudControllerManagerName},
					{Type: &appsv1.Deployment{}, Name: hcloud.CloudControllerManagerName},
					{Type: &corev1.ConfigMap{}, Name: hcloud.CloudControllerManagerName + "-observability-config"},
					{Type: &autoscalingv1beta2.VerticalPodAutoscaler{}, Name: hcloud.CloudControllerManagerName + "-vpa"},
				},
			},
			{
				Name: hcloud.CSIControllerName,
				Images: []string{
					hcloud.CSIAttacherImageName,
					hcloud.CSIProvisionerImageName,
					hcloud.CSIDriverControllerImageName,
					hcloud.CSIResizerImageName,
					hcloud.LivenessProbeImageName},
				Objects: []*chart.Object{
					{Type: &appsv1.Deployment{}, Name: hcloud.CSIControllerName},
					{Type: &corev1.ConfigMap{}, Name: hcloud.CSIControllerName + "-observability-config"},
					{Type: &autoscalingv1beta2.VerticalPodAutoscaler{}, Name: hcloud.CSIControllerName + "-vpa"},
				},
			},
		},
	}

	controlPlaneShootChart = &chart.Chart{
		Name: "shoot-system-components",
		Path: filepath.Join(hcloud.InternalChartsPath, "shoot-system-components"),
		SubCharts: []*chart.Chart{
			{
				Name: hcloud.CloudControllerManagerName,
				Objects: []*chart.Object{
					{Type: &rbacv1.ClusterRole{}, Name: "system:cloud-controller-manager"},
					{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:cloud-controller-manager"},
					{Type: &rbacv1.ClusterRole{}, Name: "system:controller:cloud-node-controller"},
					{Type: &rbacv1.ClusterRoleBinding{}, Name: "system:controller:cloud-node-controller"},
				},
			},
			{
				Name: hcloud.CSINodeName,
				Images: []string{
					hcloud.CSINodeDriverRegistrarImageName,
					hcloud.CSIDriverNodeImageName,
					hcloud.LivenessProbeImageName,
				},
				Objects: []*chart.Object{
					// csi-driver
					{Type: &appsv1.DaemonSet{}, Name: hcloud.CSINodeName},
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

	storageClassChart = &chart.Chart{
		Name: "shoot-storageclasses",
		Path: filepath.Join(hcloud.InternalChartsPath, "shoot-storageclasses"),
	}
)

func getSecretConfigs(namespace string) []extensionssecretsmanager.SecretConfigWithOptions {
	return []extensionssecretsmanager.SecretConfigWithOptions{
		{
			Config: &secretutils.CertificateSecretConfig{
				Name:       caNameControlPlane,
				CommonName: caNameControlPlane,
				CertType:   secretutils.CACert,
			},
			Options: []secretsmanager.GenerateOption{secretsmanager.Persist()},
		},
		{
			Config: &secretutils.CertificateSecretConfig{
				Name:                        hcloud.CloudControllerManagerServerName,
				CommonName:                  hcloud.CloudControllerManagerName,
				DNSNames:                    k8sutils.DNSNamesForService(hcloud.CloudControllerManagerName, namespace),
				CertType:                    secretutils.ServerCert,
				SkipPublishingCACertificate: true,
			},
			Options: []secretsmanager.GenerateOption{secretsmanager.SignedByCA(caNameControlPlane)},
		},
	}
}

func getShootAccessSecrets(namespace string) []*gardenerutils.ShootAccessSecret {
	return []*gardenerutils.ShootAccessSecret{
		gardenerutils.NewShootAccessSecret(hcloud.CloudControllerManagerName, namespace),
		gardenerutils.NewShootAccessSecret(hcloud.CSIAttacherName, namespace),
		gardenerutils.NewShootAccessSecret(hcloud.CSIProvisionerName, namespace),
		gardenerutils.NewShootAccessSecret(hcloud.CSIControllerName, namespace),
		gardenerutils.NewShootAccessSecret(hcloud.CSIResizerName, namespace),
	}
}

// NewValuesProvider creates a new ValuesProvider for the generic actuator.
//
// PARAMETERS
// logger   logr.Logger Logger instance
// gardenID string      Garden ID
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
//
// PARAMETERS
// ctx     context.Context                  Execution context
// cp      *extensionsv1alpha1.ControlPlane Control plane struct
// cluster *extensionscontroller.Cluster    Cluster struct
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
		return nil, fmt.Errorf("could not get hcloud credentials from secret '%s/%s': %w", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name, err)
	}

	// Get config chart values
	return vp.getConfigChartValues(cpConfig, cp, cluster, credentials)
}

// GetControlPlaneChartValues returns the values for the control plane chart applied by the generic actuator.
//
// PARAMETERS
// ctx           context.Context                  Execution context
// cp            *extensionsv1alpha1.ControlPlane Control plane struct
// cluster       *extensionscontroller.Cluster    Cluster struct
// secretsReader secretsmanager.Reader            Secrets manager reader
// checksums     map[string]string                Checksums
// scaledDown    bool                             True if scaled down
func (vp *valuesProvider) GetControlPlaneChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	secretsReader secretsmanager.Reader,
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
		return nil, fmt.Errorf("could not decode infrastructureProviderStatus of controlplane '%s': %w", k8sutils.ObjectName(cp), err)
	}

	// Get credentials
	credentials, err := hcloud.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, fmt.Errorf("could not get hcloud credentials from secret '%s/%s': %w", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name, err)
	}

	// Get control plane chart values
	return vp.getControlPlaneChartValues(cpConfig, infraStatus, cp, cluster, secretsReader, credentials, checksums, scaledDown)
}

// GetControlPlaneShootChartValues returns the values for the control plane shoot chart applied by the generic actuator.
//
// PARAMETERS
// ctx     context.Context                  Execution context
// cp      *extensionsv1alpha1.ControlPlane Control plane struct
// cluster *extensionscontroller.Cluster    Cluster struct
// _       secretsmanager.Reader            Secrets manager reader
// _       map[string]string                Checksums
func (vp *valuesProvider) GetControlPlaneShootChartValues(
	ctx context.Context,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	_ secretsmanager.Reader,
	_ map[string]string,
) (map[string]interface{}, error) {
	// Get credentials
	credentials, err := hcloud.GetCredentials(ctx, vp.Client(), cp.Spec.SecretRef)
	if err != nil {
		return nil, fmt.Errorf("could not get hcloud credentials from secret '%s/%s': %w", cp.Spec.SecretRef.Namespace, cp.Spec.SecretRef.Name, err)
	}

	// Get control plane shoot chart values
	return vp.getControlPlaneShootChartValues(cp, cluster, credentials)
}

// GetStorageClassesChartValues returns the values for the shoot storageclasses chart applied by the generic actuator.
//
// PARAMETERS
// _       context.Context                  Execution context
// _       *extensionsv1alpha1.ControlPlane Control plane struct
// cluster *extensionscontroller.Cluster    Cluster struct
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
		"fsType":               cloudProfileConfig.DefaultStorageFsType,
		"volumeBindingMode":    volumeBindingMode,
		"allowVolumeExpansion": true,
	}, nil
}

// getConfigChartValues collects and returns the configuration chart values.
//
// PARAMETERS
// cpConfig    *apis.ControlPlaneConfig         Control plane config struct
// cp          *extensionsv1alpha1.ControlPlane Control plane struct
// cluster     *extensionscontroller.Cluster    Cluster struct
// credentials *hcloud.Credentials              Credentials instance
func (vp *valuesProvider) getConfigChartValues(
	cpConfig *apis.ControlPlaneConfig,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
) (map[string]interface{}, error) {
	zone := cpConfig.Zone

	region := apis.GetRegionFromZone(zone)
	if "" == region {
		region = cp.Spec.Region
	}

	// Collect config chart values
	values := map[string]interface{}{
		"token":  credentials.CCM().Token,
		"region": region,
		"zone":   zone,
	}

	return values, nil
}

// getControlPlaneChartValues collects and returns the control plane chart values.
//
// PARAMETERS
// cpConfig      *apis.ControlPlaneConfig         Control plane config struct
// infraStatus   *apis.InfrastructureStatus       Infrastructure status struct
// cp            *extensionsv1alpha1.ControlPlane Control plane struct
// cluster       *extensionscontroller.Cluster    Cluster struct
// secretsReader secretsmanager.Reader            Secrets manager reader
// credentials   *hcloud.Credentials              Credentials instance
// checksums     map[string]string                Checksums
// scaledDown    bool                             True if scaled down
func (vp *valuesProvider) getControlPlaneChartValues(
	cpConfig *apis.ControlPlaneConfig,
	infraStatus *apis.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	secretsReader secretsmanager.Reader,
	credentials *hcloud.Credentials,
	checksums map[string]string,
	scaledDown bool,
) (map[string]interface{}, error) {
	region := apis.GetRegionFromZone(cpConfig.Zone)
	if "" == region {
		region = cp.Spec.Region
	}

	ccmValues, err := vp.getCCMChartValues(cpConfig, infraStatus, cp, cluster, secretsReader, checksums, scaledDown, region)
	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		"global": map[string]interface{}{
			"genericTokenKubeconfigSecretName": extensionscontroller.GenericTokenKubeconfigSecretNameFromCluster(cluster),
		},
		hcloud.CloudControllerManagerName: ccmValues,
		hcloud.CSIControllerName:          vp.getCSIControllerChartValues(cp, cluster, credentials, checksums, scaledDown, region),
	}

	return values, nil
}

// getCCMChartValues collects and returns the CCM chart values.
//
// PARAMETERS
// cpConfig      *apis.ControlPlaneConfig         Control plane config struct
// infraStatus   *apis.InfrastructureStatus       Infrastructure status struct
// cp            *extensionsv1alpha1.ControlPlane Control plane struct
// cluster       *extensionscontroller.Cluster    Cluster struct
// secretsReader secretsmanager.Reader            Secrets manager reader
// checksums     map[string]string                Checksums
// scaledDown    bool                             True if scaled down
// region        string                           Control plane region
func (vp *valuesProvider) getCCMChartValues(
	cpConfig *apis.ControlPlaneConfig,
	infraStatus *apis.InfrastructureStatus,
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	secretsReader secretsmanager.Reader,
	checksums map[string]string,
	scaledDown bool,
	region string,
) (map[string]interface{}, error) {
	clusterID := vp.calcClusterID(cp)

	ccmSecret, found := secretsReader.Get(hcloud.CloudControllerManagerServerName)
	if !found {
		return nil, fmt.Errorf("secret %q not found", hcloud.CloudControllerManagerServerName)
	}

	values := map[string]interface{}{
		"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
		"clusterName":       clusterID,
		"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		"podAnnotations": map[string]interface{}{
			"checksum/secret-" + v1beta1constants.SecretNameCloudProvider: checksums[v1beta1constants.SecretNameCloudProvider],
			"checksum/configmap-" + hcloud.CloudProviderConfig:            checksums[hcloud.CloudProviderConfig],
		},
		"podLabels": map[string]interface{}{
			v1beta1constants.LabelPodMaintenanceRestart: "true",
		},
		"podRegion":  region,
		"serverSecretName": ccmSecret.Name,
	}

	podNetwork := extensionscontroller.GetPodNetwork(cluster)

	ipAddr, _, err := net.ParseCIDR(podNetwork)
	if err == nil && ipAddr.IsPrivate() {
		values["podNetwork"] = podNetwork
	}

	if cpConfig.CloudControllerManager != nil {
		values["featureGates"] = cpConfig.CloudControllerManager.FeatureGates
	}

	if infraStatus.NetworkIDs != nil && infraStatus.NetworkIDs.Workers != "" {
		values["podNetworkIDs"] = map[string]interface{}{
			"workers": infraStatus.NetworkIDs.Workers,
		}
	}

	return values, nil
}

// getCSIControllerChartValues collects and returns the CSIController chart values.
//
// PARAMETERS
// cp          *extensionsv1alpha1.ControlPlane Control plane struct
// cluster     *extensionscontroller.Cluster    Cluster struct
// credentials *hcloud.Credentials              Credentials instance
// checksums   map[string]string                Checksums
// scaledDown  bool                             True if scaled down
// region      string                           Control plane region
func (vp *valuesProvider) getCSIControllerChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
	checksums map[string]string,
	scaledDown bool,
	region string,
) map[string]interface{} {
	csiClusterID := vp.calcCsiClusterID(cp)

	return map[string]interface{}{
		"replicas":          extensionscontroller.GetControlPlaneReplicas(cluster, scaledDown, 1),
		"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		"clusterID":         csiClusterID,
		"token":             credentials.CSI().Token,
		"csiRegion":         region,
		// "resizerEnabled":    csiResizerEnabled,
		"podAnnotations": map[string]interface{}{
			"checksum/secret-" + hcloud.CSIProvisionerName:                checksums[hcloud.CSIProvisionerName],
			"checksum/secret-" + hcloud.CSIAttacherName:                   checksums[hcloud.CSIAttacherName],
			"checksum/secret-" + hcloud.CSIResizerName:                    checksums[hcloud.CSIResizerName],
			"checksum/secret-" + hcloud.CSIControllerName:                 checksums[hcloud.CSIControllerName],
			"checksum/secret-" + v1beta1constants.SecretNameCloudProvider: checksums[v1beta1constants.SecretNameCloudProvider],
		},
	}
}

// getControlPlaneShootChartValues collects and returns the control plane shoot chart values.
//
// PARAMETERS
// cp          *extensionsv1alpha1.ControlPlane Control plane struct
// cluster     *extensionscontroller.Cluster    Cluster struct
// credentials *hcloud.Credentials              Credentials instance
func (vp *valuesProvider) getControlPlaneShootChartValues(
	cp *extensionsv1alpha1.ControlPlane,
	cluster *extensionscontroller.Cluster,
	credentials *hcloud.Credentials,
) (map[string]interface{}, error) {
	csiClusterID := vp.calcCsiClusterID(cp)

	values := map[string]interface{}{
		hcloud.CSINodeName: map[string]interface{}{
			// "serverName":  serverName,
			"clusterID":         csiClusterID,
			"token":             credentials.CSI().Token,
			"kubernetesVersion": cluster.Shoot.Spec.Kubernetes.Version,
		},
	}

	return values, nil
}

// calcClusterID returns the cluster ID.
//
// PARAMETERS
// cp *extensionsv1alpha1.ControlPlane Control plane struct
func (vp *valuesProvider) calcClusterID(cp *extensionsv1alpha1.ControlPlane) string {
	return cp.Namespace + "-" + vp.gardenID
}

// calcCsiClusterID returns the CSI cluster ID.
//
// PARAMETERS
// cp *extensionsv1alpha1.ControlPlane Control plane struct
func (vp *valuesProvider) calcCsiClusterID(cp *extensionsv1alpha1.ControlPlane) string {
	return shortenID(vp.calcClusterID(cp), 63)
}

// shortenID returns a shortened ID with the given size.
//
// PARAMETERS
// id     string ID
// maxlen int    Maximum length
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
