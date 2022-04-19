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
	"fmt"
	"strings"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud"
	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/hcloud/apis/transcoder"

	"github.com/coreos/go-systemd/v22/unit"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gcontext "github.com/gardener/gardener/extensions/pkg/webhook/context"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener/extensions/pkg/webhook/controlplane/genericmutator"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	"github.com/Masterminds/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// InjectClient injects the given client into the ensurer.
func (e *ensurer) InjectClient(client client.Client) error {
	e.client = client
	return nil
}

// EnsureKubeAPIServerDeployment ensures that the kube-apiserver deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeAPIServerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	template := &new.Spec.Template
	ps := &template.Spec
	if c := extensionswebhook.ContainerWithName(ps.Containers, "kube-apiserver"); c != nil {
		ensureKubeAPIServerCommandLineArgs(c)
	}
	return e.ensureChecksumAnnotations(ctx, &new.Spec.Template, new.Namespace)
}

// EnsureKubeControllerManagerDeployment ensures that the kube-controller-manager deployment conforms to the provider requirements.
func (e *ensurer) EnsureKubeControllerManagerDeployment(ctx context.Context, gctx gcontext.GardenContext, new, old *appsv1.Deployment) error {
	ensureKubeControllerManagerAnnotations(&new.Spec.Template)
	return e.ensureChecksumAnnotations(ctx, &new.Spec.Template, new.Namespace)
}

func ensureKubeAPIServerCommandLineArgs(c *corev1.Container) {
	// Ensure CSI-related admission plugins
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--enable-admission-plugins=",
		"PersistentVolumeLabel", ",")
	c.Command = extensionswebhook.EnsureStringWithPrefixContains(c.Command, "--disable-admission-plugins=",
		"PersistentVolumeLabel", ",")

	// Ensure CSI-related feature gates
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSINodeInfo=false", ",")
	c.Command = extensionswebhook.EnsureNoStringWithPrefixContains(c.Command, "--feature-gates=",
		"CSIDriverRegistry=false", ",")
}

func ensureKubeControllerManagerAnnotations(t *corev1.PodTemplateSpec) {
	// make sure to always remove this label
	delete(t.Labels, v1beta1constants.LabelNetworkPolicyToBlockedCIDRs)

	t.Labels = extensionswebhook.EnsureAnnotationOrLabel(t.Labels, v1beta1constants.LabelNetworkPolicyToPublicNetworks, v1beta1constants.LabelNetworkPolicyAllowed)
	t.Labels = extensionswebhook.EnsureAnnotationOrLabel(t.Labels, v1beta1constants.LabelNetworkPolicyToPrivateNetworks, v1beta1constants.LabelNetworkPolicyAllowed)
}

func (e *ensurer) ensureChecksumAnnotations(ctx context.Context, template *corev1.PodTemplateSpec, namespace string) error {
	return controlplane.EnsureConfigMapChecksumAnnotation(ctx, template, e.client, namespace, hcloud.CloudProviderConfig)
}

// EnsureKubeletServiceUnitOptions ensures that the kubelet.service unit options conform to the provider requirements.
func (e *ensurer) EnsureKubeletServiceUnitOptions(ctx context.Context, gctx gcontext.GardenContext, kubeletVersion *semver.Version, new, old []*unit.UnitOption) ([]*unit.UnitOption, error) {
	if opt := extensionswebhook.UnitOptionWithSectionAndName(new, "Service", "ExecStart"); opt != nil {
		command := extensionswebhook.DeserializeCommandLine(opt.Value)
		command = ensureKubeletCommandLineArgs(command, kubeletVersion)
		opt.Value = extensionswebhook.SerializeCommandLine(command, 1, " \\\n    ")
	}

	new = extensionswebhook.EnsureUnitOption(new, &unit.UnitOption{
		Section: "Service",
		Name:    "ExecStartPre",
		Value:   `/bin/sh -c 'hostnamectl set-hostname $(cat /etc/hostname | cut -d '.' -f 1)'`,
	})
	return new, nil
}

func ensureKubeletCommandLineArgs(command []string, kubeletVersion *semver.Version) []string {
	firstUnsupportedVersion := semver.MustParse("v1.23")

	if kubeletVersion.LessThan(firstUnsupportedVersion) {
		command = extensionswebhook.EnsureStringWithPrefix(command, "--cloud-provider=", "external")
		command = extensionswebhook.EnsureStringWithPrefix(command, "--enable-controller-attach-detach=", "true")
	}

	return command
}

// EnsureKubeletConfiguration ensures that the kubelet configuration conforms to the provider requirements.
func (e *ensurer) EnsureKubeletConfiguration(ctx context.Context, gctx gcontext.GardenContext, kubeletVersion *semver.Version, new, old *kubeletconfigv1beta1.KubeletConfiguration) error {
	// Make sure CSI-related feature gates are not enabled
	// TODO Leaving these enabled shouldn't do any harm, perhaps remove this code when properly tested?
	delete(new.FeatureGates, "VolumeSnapshotDataSource")
	delete(new.FeatureGates, "CSINodeInfo")
	delete(new.FeatureGates, "CSIDriverRegistry")

	firstUnsupportedVersion := semver.MustParse("v1.23")

	if kubeletVersion.LessThan(firstUnsupportedVersion) {
		new.EnableControllerAttachDetach = pointer.Bool(true)
	}

	return nil
}

// ShouldProvisionKubeletCloudProviderConfig returns true if the cloud provider config file should be added to the kubelet configuration.
func (e *ensurer) ShouldProvisionKubeletCloudProviderConfig(context.Context, gcontext.GardenContext, *semver.Version) bool {
	return true
}

// EnsureKubeletCloudProviderConfig ensures that the cloud provider config file conforms to the provider requirements.
func (e *ensurer) EnsureKubeletCloudProviderConfig(ctx context.Context, gctx gcontext.GardenContext, kubeletVersion *semver.Version, data *string, namespace string) error {
	// Get `cloud-provider-config` ConfigMap
	var cm corev1.ConfigMap
	err := e.client.Get(ctx, kutil.Key(namespace, hcloud.CloudProviderConfig), &cm)
	if err != nil {
		if errors.IsNotFound(err) {
			e.logger.Info("configmap not found", "name", hcloud.CloudProviderConfig, "namespace", namespace)
			return nil
		}
		return fmt.Errorf("could not get configmap '%s/%s': %w", namespace, hcloud.CloudProviderConfig, err)
	}

	// Check if the data has "cloudprovider.conf" key
	if cm.Data == nil || cm.Data[hcloud.CloudProviderConfigMapKey] == "" {
		return nil
	}

	// Overwrite data variable
	*data = cm.Data[hcloud.CloudProviderConfigMapKey]
	return nil
}

// EnsureAdditionalFile ensures additional systemd files
// "old" might be "nil" and must always be checked.
func (e *ensurer) EnsureAdditionalFiles(ctx context.Context, gctx gcontext.GardenContext, new, old *[]extensionsv1alpha1.File) error {
	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromGardenContext(ctx, gctx)
	if err != nil {
		return err
	}

	if cloudProfileConfig.DockerDaemonOptions != nil && cloudProfileConfig.DockerDaemonOptions.HTTPProxyConf != nil {
		addDockerHTTPProxyFile(new, *cloudProfileConfig.DockerDaemonOptions.HTTPProxyConf)
	}

	if cloudProfileConfig.DockerDaemonOptions != nil && len(cloudProfileConfig.DockerDaemonOptions.InsecureRegistries) != 0 {
		addMergeDockerJSONFile(new, cloudProfileConfig.DockerDaemonOptions.InsecureRegistries)
	}

	return nil
}

func addDockerHTTPProxyFile(new *[]extensionsv1alpha1.File, httpProxyConf string) {
	var (
		permissions int32 = 0644
	)

	appendUniqueFile(new, extensionsv1alpha1.File{
		Path:        "/etc/systemd/system/docker.service.d/http-proxy.conf",
		Permissions: &permissions,
		Content: extensionsv1alpha1.FileContent{
			Inline: &extensionsv1alpha1.FileContentInline{
				Encoding: "",
				Data:     httpProxyConf,
			},
		},
	})
}

func addMergeDockerJSONFile(new *[]extensionsv1alpha1.File, insecureRegistries []string) {
	var (
		permissions int32 = 0755
		template          = `#!/bin/sh
DOCKER_CONF=/etc/docker/daemon.json

if [ ! -f ${DOCKER_CONF} ]; then
  echo "{}" > ${DOCKER_CONF}
fi
if [ ! -f ${DOCKER_CONF}.org ]; then
  mv ${DOCKER_CONF} ${DOCKER_CONF}.org
fi
echo '{"insecure-registries":["@@"]}' | jq -s '.[0] * .[1]' ${DOCKER_CONF}.org - > ${DOCKER_CONF}
`
	)

	content := strings.ReplaceAll(template, "@@", strings.Join(insecureRegistries, `","`))
	appendUniqueFile(new, extensionsv1alpha1.File{
		Path:        "/opt/bin/merge-docker-json.sh",
		Permissions: &permissions,
		Content: extensionsv1alpha1.FileContent{
			Inline: &extensionsv1alpha1.FileContentInline{
				Encoding: "",
				Data:     content,
			},
		},
	})
}

// EnsureAdditionalUnits ensures that additional required system units are added.
func (e *ensurer) EnsureAdditionalUnits(ctx context.Context, gctx gcontext.GardenContext, new, _ *[]extensionsv1alpha1.Unit) error {
	var (
		command           = "start"
		trueVar           = true
		customUnitContent = `[Unit]
Description=Extend dockerd configuration file
Before=dockerd.service
[Install]
WantedBy=dockerd.service
[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/opt/bin/merge-docker-json.sh
`
	)

	cloudProfileConfig, err := transcoder.DecodeCloudProfileConfigFromGardenContext(ctx, gctx)
	if err != nil {
		return err
	}

	if cloudProfileConfig.DockerDaemonOptions != nil && len(cloudProfileConfig.DockerDaemonOptions.InsecureRegistries) != 0 {
		extensionswebhook.AppendUniqueUnit(new, extensionsv1alpha1.Unit{
			Name:    "merge-docker-json.service",
			Enable:  &trueVar,
			Command: &command,
			Content: &customUnitContent,
		})
	}
	return nil
}

// appendUniqueFile appends a unit file only if it does not exist, otherwise overwrite content of previous files
func appendUniqueFile(files *[]extensionsv1alpha1.File, file extensionsv1alpha1.File) {
	resFiles := make([]extensionsv1alpha1.File, 0, len(*files))

	for _, f := range *files {
		if f.Path != file.Path {
			resFiles = append(resFiles, f)
		}
	}

	*files = append(resFiles, file)
}
