#!/usr/bin/env bash

TOOLS_DIR=$(dirname "$(pwd)/${BASH_SOURCE[0]}")
TOOLS_BIN_DIR="$TOOLS_DIR/bin"

HELM="$TOOLS_BIN_DIR/helm"
KIND="$TOOLS_BIN_DIR/kind"
KUBECTL="$TOOLS_BIN_DIR/kubectl"
YQ="$TOOLS_BIN_DIR/yq"

TOOLS_KERNEL="$(uname -s | tr '[:upper:]' '[:lower:]')"
TOOLS_ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

_versionFile() {
  echo "$TOOLS_BIN_DIR/.version_$(basename $1)"
}

_isStale() {
  local dep=$1
  local ver=$2
  local versionFile=$(_versionFile $dep)

  if [ ! -f "$dep" ]; then
    return 0
  elif [ ! -f "$versionFile" ]; then
    return 0
  elif [[ $ver != $(cat $versionFile) ]]; then
    return 0
  fi

  return 1
}

_setVersion() {
  local dep=$1
  local ver=$2
  local versionFile=$(_versionFile $dep)

  echo -n "$ver" > "$versionFile"
}

install_kind() {
  # not under renovate control
  VERSION=v0.20.0

  if _isStale "$KIND" "$VERSION"; then
		curl -L -o "$KIND" "https://github.com/kubernetes-sigs/kind/releases/download/${VERSION}/kind-${TOOLS_KERNEL}-${TOOLS_ARCH}"
		chmod +x "$KIND"

    _setVersion "$KIND" "$VERSION"
  fi
}


kind() {
		install_kind
		$KIND "$@"
}

install_kubectl() {
  # renovate: datasource=github-tags depName=kubernetes/kubectl
  VERSION=v0.28.4

  if _isStale $KUBECTL $VERSION; then
    curl -Lo $KUBECTL "https://dev-tools-proxy.ingress.23ke-releases.23t-prod.okeanos.dev/kubectl/release/${VERSION/v0/v1}/bin/$TOOLS_KERNEL/$TOOLS_ARCH/kubectl"
    chmod +x $KUBECTL

    _setVersion $KUBECTL $VERSION
  fi
}

kubectl () {
		install_kubectl
		$KUBECTL "$@"
}

install_yq() {
  # renovate: datasource=github-releases depName=mikefarah/yq
  VERSION=v4.40.3

  if _isStale $YQ $VERSION; then
    curl -L -o $YQ "https://github.com/mikefarah/yq/releases/download/$VERSION/yq_${TOOLS_KERNEL}_$TOOLS_ARCH"
    chmod +x $YQ

    _setVersion $YQ $VERSION
  fi
}

yq () {
		install_yq
		$YQ "$@"
}

install_helm() {
  # renovate: datasource=github-tags depName=helm/helm
  VERSION=v3.13.2

  if _isStale $HELM $VERSION; then
    curl -L "https://get.helm.sh/helm-$VERSION-$TOOLS_KERNEL-$TOOLS_ARCH.tar.gz" | tar -xzm --strip-components 1 -C $TOOLS_BIN_DIR "$TOOLS_KERNEL-$TOOLS_ARCH/helm"
    chmod +x $HELM

    _setVersion $HELM $VERSION
  fi
}

helm () {
		install_helm
		$HELM "$@"
}
