module github.com/23technologies/gardener-extension-provider-hcloud

go 1.15

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/gardener/controller-manager-library v0.2.1-0.20200810091329-d980dbe10959
	github.com/gardener/etcd-druid v0.3.0
	github.com/gardener/gardener v1.15.1-0.20210112065447-570ae178874b
	github.com/gardener/machine-controller-manager v0.35.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/golang/mock v1.4.4-0.20200731163441-8734ec565a4d
	github.com/google/uuid v1.1.1
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/vmware/go-vmware-nsxt v0.0.0-20200114231430-33a5af043f2e
	github.com/vmware/vsphere-automation-sdk-go/lib v0.3.1
	github.com/vmware/vsphere-automation-sdk-go/runtime v0.3.1
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt v0.4.0
	k8s.io/api v0.18.10
	k8s.io/apiextensions-apiserver v0.18.10
	k8s.io/apimachinery v0.18.10
	k8s.io/apiserver v0.18.10
	k8s.io/autoscaler v0.0.0-20190805135949-100e91ba756e
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider-vsphere v1.1.0
	k8s.io/code-generator v0.18.10
	k8s.io/component-base v0.18.10
	k8s.io/kubelet v0.18.10
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.18.10
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.10
	k8s.io/apiserver => k8s.io/apiserver v0.18.10
	k8s.io/client-go => k8s.io/client-go v0.18.10
	k8s.io/code-generator => k8s.io/code-generator v0.18.10
	k8s.io/component-base => k8s.io/component-base v0.18.10
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.10
)

// needed for infra-cli and load balancer cleanup
replace k8s.io/cloud-provider-vsphere => github.com/MartinWeindel/cloud-provider-vsphere v1.0.1-0.20201008150334-6535a3d35ffc

replace (
	// these replacements are needed for cloud-provider-vsphere
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200715165012-dff82dc0de47
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200715165012-dff82dc0de47
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200715165012-dff82dc0de47
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200715165012-dff82dc0de47
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200715165012-dff82dc0de47
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200715165012-dff82dc0de47
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200715165012-dff82dc0de47
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200715165012-dff82dc0de47
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20200715165012-dff82dc0de47
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200715165012-dff82dc0de47
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200715165012-dff82dc0de47
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200715165012-dff82dc0de47
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200715165012-dff82dc0de47
)
