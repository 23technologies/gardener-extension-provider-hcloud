#! /usr/bin/env bash
set -eu

if [[ ! -v AZURE_DNS_CLIENT_ID ]] ||
   [[ ! -v AZURE_DNS_CLIENT_SECRET ]] ||
   [[ ! -v AZURE_DNS_SUBSCRIPTION_ID ]] ||
   [[ ! -v AZURE_DNS_TENANT_ID ]] ||
   [[ ! -v HCLOUD_TOKEN ]]; then
    cat <<EOF
Please export the following variables with appropriate values

export AZURE_DNS_CLIENT_ID=YOUR-SECRET
export AZURE_DNS_CLIENT_SECRET=YOUR-SECRET
export AZURE_DNS_SUBSCRIPTION_ID=YOUR-SECRET
export AZURE_DNS_TENANT_ID=YOUR-SECRET
export HCLOUD_TOKEN=YOUR-SECRET
EOF
    exit 1
fi

if [[ ! -f hack/ci/secrets/gardener-kubeconfig.yaml ]]; then
    cat <<EOF
Please create the file hack/ci/secrets/gardener-kubeconfig.yaml with a kubeconfig pointing to a Gardener project.
EOF
    exit 1
fi

source hack/tools/install.sh
export PATH=$PATH:$(pwd)/hack/tools/bin/

git clone https://github.com/gardener/gardener.git
cd gardener || exit
git checkout "$(git tag -l 'v1.*' | sort --version-sort | tail -1)"

# Create a shoot on okeanos.dev and use this as seed cluster for the ci run
export KUBECONFIG=../hack/ci/secrets/gardener-kubeconfig.yaml
export SHOOT_NAME=ci-seed-$(openssl rand -hex 2)
yq '.metadata.name=env(SHOOT_NAME)' ../hack/ci/misc/shoot-for-seed.yaml | kubectl apply -f -

echo "Waiting for shoot creation..."
while [ ! "$(kubectl get shoot "$SHOOT_NAME" -o jsonpath="{.status.lastOperation.state}")" == "Succeeded" ]; do
  PERCENTAGE=$(kubectl get shoot "$SHOOT_NAME" -o jsonpath="{.status.lastOperation.progress}")
  echo "Creating shoot: $PERCENTAGE%"
  sleep 5
done
echo "Shoot creation succeeded"

# This will get us the kubeconfig for our seed cluster
kubectl create \
    -f <(printf '{"spec":{"expirationSeconds":360000}}') \
    --raw /apis/core.gardener.cloud/v1beta1/namespaces/garden-23ke-ci/shoots/"${SHOOT_NAME}"/adminkubeconfig | \
    jq -r ".status.kubeconfig" | \
    base64 -d > example/provider-extensions/seed/kubeconfig

cat <<EOF > example/provider-extensions/garden/controlplane/domain-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: default-domain-external-provider-extensions-gardener-cloud
  namespace: garden
  labels:
    app: gardener
    gardener.cloud/role: default-domain
  annotations:
    dns.gardener.cloud/provider: azure-dns
    dns.gardener.cloud/domain: test-cluster.23ke-testbed.23t.dev
    # dns.gardener.cloud/zone: ""
    # dns.gardener.cloud/domain-default-priority: "10"
type: Opaque
data:
  clientID: ${AZURE_DNS_CLIENT_ID}
  clientSecret: ${AZURE_DNS_CLIENT_SECRET}
  subscriptionID: ${AZURE_DNS_SUBSCRIPTION_ID}
  tenantID: ${AZURE_DNS_TENANT_ID}
---
apiVersion: v1
kind: Secret
metadata:
  name: internal-domain-internal-provider-extensions-gardener-cloud
  namespace: garden
  labels:
    app: gardener
    gardener.cloud/role: internal-domain
  annotations:
    dns.gardener.cloud/provider: azure-dns
    dns.gardener.cloud/domain: test-cluster.23ke-testbed.23t.dev
    # dns.gardener.cloud/zone: ""
type: Opaque
data:
  clientID: ${AZURE_DNS_CLIENT_ID}
  clientSecret: ${AZURE_DNS_CLIENT_SECRET}
  subscriptionID: ${AZURE_DNS_SUBSCRIPTION_ID}
  tenantID: ${AZURE_DNS_TENANT_ID}
EOF

cat <<EOF > example/provider-extensions/garden/project/credentials/infrastructure-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: hcloud-secret
  namespace: garden-project-1
type: Opaque
data:
  hcloudToken: ${HCLOUD_TOKEN}
EOF


cat <<EOF > example/provider-extensions/garden/project/credentials/secretbindings.yaml
apiVersion: core.gardener.cloud/v1beta1
kind: SecretBinding
metadata:
  name: hcloud-secret
  namespace: garden-project-1
  labels:
    cloudprofile.garden.sapcloud.io/name: hcloud
provider:
  type: hcloud
secretRef:
  name: hcloud-secret
  namespace: garden-project-1
EOF

cat <<EOF > example/provider-extensions/garden/project/project.yaml
---
apiVersion: core.gardener.cloud/v1beta1
kind: Project
metadata:
  name: garden
spec:
  namespace: garden
---
apiVersion: v1
kind: Namespace
metadata:
  name: "garden-project-1"
  labels:
    gardener.cloud/role: project
    project.gardener.cloud/name: "project-1"
---
apiVersion: core.gardener.cloud/v1beta1
kind: Project
metadata:
  name: "project-1"
spec:
  namespace: "garden-project-1"
EOF

cat <<EOF > example/provider-extensions/gardenlet/values.yaml
config:
  seedConfig:
    apiVersion: core.gardener.cloud/v1beta1
    kind: Seed
    metadata:
      # Automatically set
      name: "provider-extensions"
    spec:
      backup: null
      secretRef: null
      dns:
        provider:
          secretRef:
            # Automatically set when using a Gardener shoot
            name: "internal-domain-internal-provider-extensions-gardener-cloud"
            namespace: garden
          # Automatically set when using a Gardener shoot
          type: "azure-dns"
      ingress:
        controller:
          kind: nginx
        # Enter ingress domain of your seed
        domain: ingress.test-cluster.23ke-testbed.23t.dev
      networks:
        blockCIDRs:
          - 169.254.169.254/32
        # Automatically set when using a Gardener shoot
        pods: "100.73.0.0/16"
        # Automatically set when using a Gardener shoot
        nodes: "10.250.0.0/16"
        # Automatically set when using a Gardener shoot
        services: "100.88.0.0/13"
        shootDefaults:
          pods: 100.1.0.0/24
          services: 100.2.0.0/24
      provider:
        # Automatically set when using a Gardener shoot
        region: "fsn1"
        # Automatically set when using a Gardener shoot
        type: "hcloud"
        # Enter zones of your seed
        zones:
          - nova
      settings:
        dependencyWatchdog:
          weeder:
            enabled: true
          prober:
            enabled: true
        excessCapacityReservation:
          enabled: false
        scheduling:
          visible: true
        verticalPodAutoscaler:
          enabled: true
EOF



# generate controllerregistrations and controllerdeployments via the gardener-community helm-charts
helm repo add gardener-charts https://gardener-community.github.io/gardener-charts
helm repo update

helm template ext-provider-hcloud gardener-charts/provider-hcloud --set controller.enabled=true > example/provider-extensions/garden/controllerregistrations/provider-hcloud.yaml
helm template ext-provider-azure gardener-charts/provider-azure --set controller.enabled=true > example/provider-extensions/garden/controllerregistrations/provider-azure.yaml
helm template ext-provider-openstack gardener-charts/provider-openstack --set controller.enabled=true > example/provider-extensions/garden/controllerregistrations/provider-openstack.yaml
helm template networking-calico gardener-charts/networking-calico --set controller.enabled=true > example/provider-extensions/garden/controllerregistrations/networking-calico.yaml
helm template os-ubuntu gardener-charts/os-ubuntu --set controller.enabled=true > example/provider-extensions/garden/controllerregistrations/os-ubuntu.yaml

# Define a hcloud cloudprofile
cat <<EOF > example/provider-extensions/garden/cloudprofiles/hcloud.yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  annotations:
    meta.helm.sh/release-name: flux-system-cloudprofiles
    meta.helm.sh/release-namespace: flux-system
  labels:
    app.kubernetes.io/managed-by: Helm
    helm.toolkit.fluxcd.io/name: cloudprofiles
    helm.toolkit.fluxcd.io/namespace: flux-system
    provider.extensions.gardener.cloud/hcloud: "true"
  name: hcloud
spec:
  kubernetes:
    versions:
    - classification: deprecated
      version: 1.26.10
    - classification: supported
      version: 1.26.11
    - classification: deprecated
      version: 1.26.5
    - classification: deprecated
      version: 1.26.6
    - classification: deprecated
      version: 1.26.7
    - classification: deprecated
      version: 1.26.8
    - classification: deprecated
      version: 1.26.9
    - classification: preview
      version: 1.27.8
  machineImages:
  - name: ubuntu
    updateStrategy: major
    versions:
    - architectures:
      - amd64
      cri:
      - containerRuntimes:
        - type: gvisor
        name: containerd
      - name: docker
      version: 20.4.20210616
    - architectures:
      - amd64
      cri:
      - containerRuntimes:
        - type: gvisor
        name: containerd
      - name: docker
      version: 22.4.20231020
  machineTypes:
  - architecture: amd64
    cpu: "2"
    gpu: "0"
    memory: 8Gi
    name: cx31
    usable: true
  - architecture: amd64
    cpu: "4"
    gpu: "0"
    memory: 8Gi
    name: cpx31
    usable: true
  - architecture: amd64
    cpu: "4"
    gpu: "0"
    memory: 16Gi
    name: cx41
    usable: true
  - architecture: amd64
    cpu: "8"
    gpu: "0"
    memory: 16Gi
    name: cpx41
    usable: true
  providerConfig:
    apiVersion: hcloud.provider.extensions.gardener.cloud/v1alpha1
    defaultStorageFsType: ext4
    kind: CloudProfileConfig
    machineImages:
    - name: ubuntu
      versions:
      - imageName: ubuntu-20.04
        version: 20.4.20210616
    - name: ubuntu
      versions:
      - imageName: ubuntu-22.04
        version: 22.4.20231020
    regions:
    - name: hel1
    - name: fsn1
    - name: nbg1
    - name: ash
  regions:
  - name: hel1
    zones:
    - name: hel1-dc2
  - name: fsn1
    zones:
    - name: fsn1-dc14
  - name: nbg1
    zones:
    - name: nbg1-dc3
  - name: ash
    zones:
    - name: ash-dc1
      unavailableMachineTypes:
      - cx21
      - cx31
      - cx41
      - cx51
      - ccx11
      - ccx21
      - ccx31
      - ccx41
      - ccx51
  type: hcloud
  seedSelector:
    providerTypes:
    - openstack
EOF




make kind-extensions-up
make gardener-extensions-up

# Create a test shoot now
kind export kubeconfig -n gardener-extensions
kubectl apply -f ../hack/ci/misc/test-shoot.yaml

echo "Waiting for shoot creation..."
while [ ! "$(kubectl get shoot -n garden-project-1 test-shoot -o jsonpath="{.status.lastOperation.state}")" == "Succeeded" ]; do
  PERCENTAGE=$(kubectl get shoot -n garden-project-1 test-shoot -o jsonpath="{.status.lastOperation.progress}")
  echo "Creating shoot: $PERCENTAGE%"
  sleep 5
done
echo "Shoot creation succeeded"

# And delete the test-shoot again
kubectl annotate shoot -n garden-project-1 test-shoot confirmation.gardener.cloud/deletion=true --overwrite=true
kubectl delete shoot -n garden-project-1 test-shoot --wait=true

# Tear down the gardener environment
make gardener-extensions-down
make kind-extensions-clean

# Delete the shoot on okeanos.dev
export KUBECONFIG=../hack/ci/secrets/gardener-kubeconfig.yaml
kubectl annotate shoot "$SHOOT_NAME" confirmation.gardener.cloud/deletion=true --overwrite=true || echo "Annotating shoot failed"
kubectl delete shoot "$SHOOT_NAME" --wait=false || echo "Deleting shoot failed"
