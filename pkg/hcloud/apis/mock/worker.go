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

// Package mock provides all methods required to simulate a driver
package mock
/*
import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/golang/mock/gomock"
	"github.com/onsi/ginkgo"
)

func NewWorkerSpec() MockTestEnv {
	ctrl := gomock.NewController(ginkgo.GinkgoT())

	return &extensionsv1alpha1.Worker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
		Spec: extensionsv1alpha1.WorkerSpec{
			SecretRef: corev1.SecretReference{
				Name:      "secret",
				Namespace: namespace,
			},
			Region: region,
			InfrastructureProviderStatus: &runtime.RawExtension{
				Raw: encode(&api.InfrastructureStatus{
					VPC: api.VPCStatus{
						ID: vpcID,
						Subnets: []api.Subnet{
							{
								ID:      subnetZone1,
								Purpose: "nodes",
								Zone:    zone1,
							},
							{
								ID:      subnetZone2,
								Purpose: "nodes",
								Zone:    zone2,
							},
						},
						SecurityGroups: []api.SecurityGroup{
							{
								ID:      securityGroupID,
								Purpose: "nodes",
							},
						},
					},
					IAM: api.IAM{
						InstanceProfiles: []api.InstanceProfile{
							{
								Name:    instanceProfileName,
								Purpose: "nodes",
							},
						},
					},
					EC2: api.EC2{
						KeyName: keyName,
					},
				}),
			},
			Pools: []extensionsv1alpha1.WorkerPool{
				{
					Name:           namePool1,
					Minimum:        minPool1,
					Maximum:        maxPool1,
					MaxSurge:       maxSurgePool1,
					MaxUnavailable: maxUnavailablePool1,
					MachineType:    machineType,
					MachineImage: extensionsv1alpha1.MachineImage{
						Name:    machineImageName,
						Version: machineImageVersion,
					},
					ProviderConfig: &runtime.RawExtension{
						Raw: encode(&api.WorkerConfig{
							Volume: &api.Volume{
								IOPS: &volumeIOPS,
							},
							DataVolumes: []api.DataVolume{
								{
									Name: dataVolume1Name,
									Volume: api.Volume{
										IOPS: &dataVolume1IOPS,
									},
								},
								{
									Name:       dataVolume2Name,
									SnapshotID: &dataVolume2SnapshotID,
								},
							},
						}),
					},
					UserData: userData,
					Volume: &extensionsv1alpha1.Volume{
						Type:      &volumeType,
						Size:      fmt.Sprintf("%dGi", volumeSize),
						Encrypted: &volumeEncrypted,
					},
					DataVolumes: []extensionsv1alpha1.DataVolume{
						{
							Name:      dataVolume1Name,
							Type:      &dataVolume1Type,
							Size:      fmt.Sprintf("%dGi", dataVolume1Size),
							Encrypted: &dataVolume1Encrypted,
						},
						{
							Name:      dataVolume2Name,
							Type:      &dataVolume2Type,
							Size:      fmt.Sprintf("%dGi", dataVolume2Size),
							Encrypted: &dataVolume2Encrypted,
						},
					},
					Zones: []string{
						zone1,
						zone2,
					},
					Labels: labels,
				},
				{
					Name:           namePool2,
					Minimum:        minPool2,
					Maximum:        maxPool2,
					MaxSurge:       maxSurgePool2,
					MaxUnavailable: maxUnavailablePool2,
					MachineType:    machineType,
					MachineImage: extensionsv1alpha1.MachineImage{
						Name:    machineImageName,
						Version: machineImageVersion,
					},
					UserData: userData,
					Volume: &extensionsv1alpha1.Volume{
						Type: &volumeType,
						Size: fmt.Sprintf("%dGi", volumeSize),
					},
					Zones: []string{
						zone1,
						zone2,
					},
					Labels: labels,
				},
			},
		},
	}
}
*/
