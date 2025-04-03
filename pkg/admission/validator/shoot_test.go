// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator_test

import (
	"context"
	"encoding/json"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	mockclient "github.com/gardener/gardener/third_party/mock/controller-runtime/client"
	mockmanager "github.com/gardener/gardener/third_party/mock/controller-runtime/manager"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/23technologies/gardener-extension-provider-hcloud/pkg/admission/validator"
	apishcloud "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud"
	apishcloudv1alpha "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/v1alpha1"
	hcloudv1alpha1 "github.com/23technologies/gardener-extension-provider-hcloud/pkg/apis/hcloud/v1alpha1"
)

var _ = Describe("Shoot validator", func() {
	Describe("#Validate", func() {
		const namespace = "garden-dev"

		var (
			shootValidator extensionswebhook.Validator

			ctrl      *gomock.Controller
			mgr       *mockmanager.MockManager
			c         *mockclient.MockClient
			apiReader *mockclient.MockReader
			shoot     *core.Shoot

			ctx = context.Background()

			regionName   string
			imageName    string
			imageVersion string
			architecture *string
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			scheme := runtime.NewScheme()
			Expect(apishcloud.AddToScheme(scheme)).To(Succeed())
			Expect(apishcloudv1alpha.AddToScheme(scheme)).To(Succeed())
			Expect(gardencorev1beta1.AddToScheme(scheme)).To(Succeed())

			c = mockclient.NewMockClient(ctrl)
			apiReader = mockclient.NewMockReader(ctrl)

			mgr = mockmanager.NewMockManager(ctrl)
			mgr.EXPECT().GetScheme().Return(scheme).Times(2)
			mgr.EXPECT().GetClient().Return(c)
			mgr.EXPECT().GetAPIReader().Return(apiReader)
			shootValidator = validator.NewShootValidator(mgr)

			regionName = "eu-de-1"
			imageName = "Foo"
			imageVersion = "1.0.0"
			architecture = ptr.To("analog")

			shoot = &core.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: namespace,
				},
				Spec: core.ShootSpec{
					CloudProfile: &core.CloudProfileReference{
						Kind: "CloudProfile",
						Name: "cloudProfile",
					},
					Provider: core.Provider{
						Type: "hcloud",
						Workers: []core.Worker{
							{
								Name: "worker-1",
								Volume: &core.Volume{
									VolumeSize: "50Gi",
									Type:       ptr.To("volumeType"),
								},
								Zones: []string{"zone1"},
								Machine: core.Machine{
									Image: &core.ShootMachineImage{
										Name:    imageName,
										Version: imageVersion,
									},
									Architecture: architecture,
								},
							},
						},
						InfrastructureConfig: &runtime.RawExtension{
							Raw: encode(&hcloudv1alpha1.InfrastructureConfig{
								TypeMeta: metav1.TypeMeta{
									APIVersion: hcloudv1alpha1.SchemeGroupVersion.String(),
									Kind:       "InfrastructureConfig",
								},
								Networks: &hcloudv1alpha1.Networks{
									Workers: "10.250.0.0/19",
								},
							}),
						},
						ControlPlaneConfig: &runtime.RawExtension{
							Raw: encode(&hcloudv1alpha1.ControlPlaneConfig{
								TypeMeta: metav1.TypeMeta{
									APIVersion: hcloudv1alpha1.SchemeGroupVersion.String(),
									Kind:       "ControlPlaneConfig",
								},
							}),
						},
					},
					Region: "eu-de-1",
					Networking: &core.Networking{
						Nodes: ptr.To("10.250.0.0/16"),
					},
				},
			}
		})

		Context("Workerless Shoot", func() {
			BeforeEach(func() {
				shoot.Spec.Provider.Workers = nil
			})

			It("should not validate", func() {
				err := shootValidator.Validate(ctx, shoot, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Shoot creation", func() {
			var (
				cloudProfileKey           client.ObjectKey
				namespacedCloudProfileKey client.ObjectKey

				cloudProfile           *gardencorev1beta1.CloudProfile
				namespacedCloudProfile *gardencorev1beta1.NamespacedCloudProfile
			)

			BeforeEach(func() {
				cloudProfileKey = client.ObjectKey{Name: "hcloud"}
				namespacedCloudProfileKey = client.ObjectKey{Name: "hcloud-nscpfl", Namespace: namespace}

				cloudProfile = &gardencorev1beta1.CloudProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hcloud",
					},
					Spec: gardencorev1beta1.CloudProfileSpec{
						Regions: []gardencorev1beta1.Region{
							{
								Name: regionName,
								Zones: []gardencorev1beta1.AvailabilityZone{
									{
										Name: "zone1",
									},
									{
										Name: "zone2",
									},
								},
							},
						},
						ProviderConfig: &runtime.RawExtension{
							Raw: encode(&apishcloudv1alpha.CloudProfileConfig{
								TypeMeta: metav1.TypeMeta{
									APIVersion: apishcloudv1alpha.SchemeGroupVersion.String(),
									Kind:       "CloudProfileConfig",
								},
								MachineImages: []apishcloudv1alpha.MachineImages{
									{
										Name: imageName,
										Versions: []apishcloudv1alpha.MachineImageVersion{
											{
												Version: imageVersion,
											},
										},
									},
								},
							}),
						},
					},
				}

				namespacedCloudProfile = &gardencorev1beta1.NamespacedCloudProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hcloud-nscpfl",
					},
					Spec: gardencorev1beta1.NamespacedCloudProfileSpec{
						Parent: gardencorev1beta1.CloudProfileReference{
							Kind: "CloudProfile",
							Name: "hcloud",
						},
					},
					Status: gardencorev1beta1.NamespacedCloudProfileStatus{
						CloudProfileSpec: cloudProfile.Spec,
					},
				}
			})

			It("should work for CloudProfile referenced from Shoot", func() {
				shoot.Spec.CloudProfile = &core.CloudProfileReference{
					Kind: "CloudProfile",
					Name: "hcloud",
				}
				c.EXPECT().Get(ctx, cloudProfileKey, &gardencorev1beta1.CloudProfile{}).SetArg(2, *cloudProfile)

				err := shootValidator.Validate(ctx, shoot, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should work for CloudProfile referenced from cloudProfileName", func() {
				shoot.Spec.CloudProfileName = ptr.To("hcloud")
				shoot.Spec.CloudProfile = nil
				c.EXPECT().Get(ctx, cloudProfileKey, &gardencorev1beta1.CloudProfile{}).SetArg(2, *cloudProfile)

				err := shootValidator.Validate(ctx, shoot, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should work for NamespacedCloudProfile referenced from Shoot", func() {
				shoot.Spec.CloudProfile = &core.CloudProfileReference{
					Kind: "NamespacedCloudProfile",
					Name: "hcloud-nscpfl",
				}
				c.EXPECT().Get(ctx, namespacedCloudProfileKey, &gardencorev1beta1.NamespacedCloudProfile{}).SetArg(2, *namespacedCloudProfile)

				err := shootValidator.Validate(ctx, shoot, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fail for a missing cloud profile provider config", func() {
				shoot.Spec.CloudProfile = &core.CloudProfileReference{
					Kind: "NamespacedCloudProfile",
					Name: "hcloud-nscpfl",
				}
				namespacedCloudProfile.Status.CloudProfileSpec.ProviderConfig = nil
				c.EXPECT().Get(ctx, namespacedCloudProfileKey, &gardencorev1beta1.NamespacedCloudProfile{}).SetArg(2, *namespacedCloudProfile)

				err := shootValidator.Validate(ctx, shoot, nil)
				Expect(err).To(MatchError(And(
					ContainSubstring("providerConfig is not given for cloud profile"),
					ContainSubstring("NamespacedCloudProfile"),
					ContainSubstring("hcloud-nscpfl"),
				)))
			})

			Context("", func() {
				BeforeEach(func() {
					shoot.Spec.CloudProfile = &core.CloudProfileReference{
						Kind: "CloudProfile",
						Name: "hcloud",
					}
				})

				It("should return err when networking is configured to use dual-stack", func() {
					c.EXPECT().Get(ctx, cloudProfileKey, &gardencorev1beta1.CloudProfile{}).SetArg(2, *cloudProfile)
					shoot.Spec.Networking.IPFamilies = []core.IPFamily{core.IPFamilyIPv4, core.IPFamilyIPv6}

					err := shootValidator.Validate(ctx, shoot, nil)
					Expect(err).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("spec.networking.ipFamilies"),
					}))))
				})

				It("should return err when networking is configured to use IPv6-only", func() {
					c.EXPECT().Get(ctx, cloudProfileKey, &gardencorev1beta1.CloudProfile{}).SetArg(2, *cloudProfile)
					shoot.Spec.Networking.IPFamilies = []core.IPFamily{core.IPFamilyIPv6}

					err := shootValidator.Validate(ctx, shoot, nil)
					Expect(err).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("spec.networking.ipFamilies"),
					}))))
				})
			})
		})
	})
})

func encode(obj runtime.Object) []byte {
	data, _ := json.Marshal(obj)
	return data
}
