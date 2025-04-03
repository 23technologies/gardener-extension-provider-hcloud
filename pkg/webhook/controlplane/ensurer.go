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

// Package controlplane contains functions used to provide /controlplane
package controlplane

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/coreos/go-systemd/v22/unit"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/component/nodemanagement/machinecontrollermanager"
	versionutils "github.com/gardener/gardener/pkg/utils/version"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controllerapis "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/controller"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
)

// NewEnsurer creates a new controlplane ensurer.
func NewEnsurer(logger logr.Logger) genericmutator.Ensurer {
	return &ensurer{
		logger: logger.WithName("hcloud-controlplane-ensurer"),
	}
}

type ensurer struct {
	genericmutator.NoopEnsurer
	client client.Client
	logger logr.Logger
}

// ImageVector is exposed for testing.
var ImageVector = controllerapis.ImageVector()

// InjectClient injects the given client into the ensurer.
func (e *ensurer) InjectClient(client client.Client) error {
	e.client = client
	return nil
}

// EnsureMachineControllerManagerDeployment ensures that the machine-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerDeployment(_ context.Context, _ gcontext.GardenContext, newObj, _ *appsv1.Deployment) error {
	image, err := ImageVector.FindImage(hcloud.MCMProviderHcloudImageName)
	if err != nil {
		return err
	}

	newObj.Spec.Template.Spec.Containers = extensionswebhook.EnsureContainerWithName(
		newObj.Spec.Template.Spec.Containers,
		machinecontrollermanager.ProviderSidecarContainer(newObj.Namespace, hcloud.Name, image.String()),
	)
	return nil
}

// EnsureMachineControllerManagerVPA ensures that the machine-controller-manager VPA conforms to the provider requirements.
func (e *ensurer) EnsureMachineControllerManagerVPA(_ context.Context, _ gcontext.GardenContext, newObj, _ *vpaautoscalingv1.VerticalPodAutoscaler) error {
	if newObj.Spec.ResourcePolicy == nil {
		newObj.Spec.ResourcePolicy = &vpaautoscalingv1.PodResourcePolicy{}
	}

	newObj.Spec.ResourcePolicy.ContainerPolicies = extensionswebhook.EnsureVPAContainerResourcePolicyWithName(
		newObj.Spec.ResourcePolicy.ContainerPolicies,
		machinecontrollermanager.ProviderSidecarVPAContainerPolicy(hcloud.Name),
	)
	return nil
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	template := &new.Spec.Template
	ps := &template.Spec
	cluster, err := gctx.GetCluster(ctx)
	if err != nil {
		return err
	}
	k8sVersion, err := semver.NewVersion(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return err
	}
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		ensureKubeAPIServerCommandLineArgs(c, k8sVersion)
	}
	return nil
}

// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeControllerManagerDeployment(_ context.Context, _ gcontext.GardenContext, newObj, _ *appsv1.Deployment) error {
	template := &newObj.Spec.Template
	ps := &template.Spec

	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-controller-manager"); c != nil {
		ensureKubeControllerManagerCommandLineArgs(c)
		ensureKubeControllerManagerVolumeMounts(c)
	}

	ensureKubeControllerManagerLabels(template)
	ensureKubeControllerManagerVolumes(ps)
	return nil
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container, k8sVersion *semver.Version) {
	// Ensure CSI-related admission plugins
	if versionutils.ConstraintK8sLess131.Check(k8sVersion) {
		c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
			"PersistentVolumeLabel", ",")
		c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
			"PersistentVolumeLabel", ",")
	}

	// Ensure CSI-related feature gates
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSINodeInfo=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSIDriverRegistry=false", ",")
}

func ensureKubeControllerManagerCommandLineArgs(c *corev1.Container) {
	c.Command = extensionswebhook.EnsureStringWithPrefix(c.Command, "--cloud-provider=", "external")
	c.Command = extensionswebhook.EnsureNoStringWithPrefix(c.Command, "--cloud-config=")
	c.Command = extensionswebhook.EnsureNoStringWithPrefix(c.Command, "--external-cloud-volume-plugin=")
}

func ensureKubeControllerManagerLabels(t *corev1.PodTemplateSpec) {
	// TODO: This can be removed in a future version.
	delete(t.Labels, v1beta1constants.LabelNetworkPolicyToBlockedCIDRs)

	delete(t.Labels, v1beta1constants.LabelNetworkPolicyToPublicNetworks)
	delete(t.Labels, v1beta1constants.LabelNetworkPolicyToPrivateNetworks)
}

var (
	etcSSLName        = "etc-ssl"
	etcSSLVolumeMount = corev1.VolumeMount{
		Name:      etcSSLName,
		MountPath: "/etc/ssl",
		ReadOnly:  true,
	}
	directoryOrCreate = corev1.HostPathDirectoryOrCreate
	etcSSLVolume      = corev1.Volume{
		Name: etcSSLName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/etc/ssl",
				Type: &directoryOrCreate,
			},
		},
	}

	usrShareCACertificatesName        = "usr-share-ca-certificates"
	usrShareCACertificatesVolumeMount = corev1.VolumeMount{
		Name:      usrShareCACertificatesName,
		MountPath: "/usr/share/ca-certificates",
		ReadOnly:  true,
	}
	usrShareCACertificatesVolume = corev1.Volume{
		Name: usrShareCACertificatesName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/usr/share/ca-certificates",
			},
		},
	}
)

func ensureKubeControllerManagerVolumeMounts(c *corev1.Container) {
	c.VolumeMounts = extensionswebhook.EnsureNoVolumeMountWithName(c.VolumeMounts, etcSSLVolumeMount.Name)
	c.VolumeMounts = extensionswebhook.EnsureNoVolumeMountWithName(c.VolumeMounts, usrShareCACertificatesVolumeMount.Name)
}

func ensureKubeControllerManagerVolumes(ps *corev1.PodSpec) {
	ps.Volumes = extensionswebhook.EnsureNoVolumeWithName(ps.Volumes, etcSSLVolume.Name)
	ps.Volumes = extensionswebhook.EnsureNoVolumeWithName(ps.Volumes, usrShareCACertificatesVolume.Name)
}

// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
func (e *ensurer) EnsureKubeletServiceUnitOptions(_ context.Context, _ gcontext.GardenContext, _ *semver.Version, newObj, _ []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(newObj, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command)
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}

	newObj = extensionswebhook.EnsureUnitOption(newObj, &unit.UnitOption{
		Section: "Service",
		Name:    "ExecStartPre",
		Value:   `/bin/sh -c 'hostnamectl set-hostname $(cat /etc/hostname | cut -d '.' -f 1)'`,
	})
	return newObj, nil
}

func ensureKubeletCommandLineArgs(command []string) []string {
	return extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "external")
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(_ context.Context, _ gcontext.GardenContext, _ *semver.Version, newObj, _ *kubeletconfigv1beta1.KubeletConfiguration) error {
	newObj.EnableControllerAttachDetach = ptr.To(true)
	return nil
}
