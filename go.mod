module github.com/23technologies/gardener-extension-provider-hcloud

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/gardener/etcd-druid v0.5.0
	github.com/gardener/gardener v1.32.0
	github.com/gardener/machine-controller-manager v0.40.0
	github.com/go-logr/logr v0.4.0
	github.com/golang/mock v1.6.0
	github.com/hetznercloud/hcloud-go v1.32.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/autoscaler v0.0.0-20190805135949-100e91ba756e
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.21.2
	k8s.io/component-base v0.21.2
	k8s.io/kubelet v0.21.2
	sigs.k8s.io/controller-runtime v0.9.1
)

replace (
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.9.5+incompatible // keep this value in sync with k8s.io/apiserver
	github.com/gardener/gardener => github.com/gardener/gardener v1.32.0
	github.com/gardener/gardener-resource-manager/api => github.com/gardener/gardener-resource-manager/api v0.25.0
	github.com/gardener/machine-controller-manager => github.com/gardener/machine-controller-manager v0.40.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.0 // keep this value in sync with sigs.k8s.io/controller-runtime
	google.golang.org/grpc => google.golang.org/grpc v1.27.1 // keep this value in sync with k8s.io/apiserver
	k8s.io/api => k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.2
	k8s.io/apiserver => k8s.io/apiserver v0.21.2
	k8s.io/client-go => k8s.io/client-go v0.21.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.2
	k8s.io/code-generator => k8s.io/code-generator v0.21.2
	k8s.io/component-base => k8s.io/component-base v0.21.2
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.2
)
