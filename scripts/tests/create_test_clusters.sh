#!/bin/bash

# Parse arguments
SKIP_VERSION_CHECK=false
if [ "$1" == "--skip-version-check" ]; then
  SKIP_VERSION_CHECK=true
  shift
fi

# Check for invalid or extra arguments
if [ $# -ne 0 ]; then
  echo "[$(basename $0)] Invalid argument(s): $@"
  echo "Usage: $(basename $0) [--skip-version-check]"
  exit 1
fi

# Change to the directory where the script is located
cd "$(dirname "$0")"

# Run version check unless skipped
if [ "$SKIP_VERSION_CHECK" = false ]; then
  ./version.sh
  if [ $? -ne 0 ]; then
    echo "[$(basename $0)] Versioning check failed. Exiting..."
    exit 1
  fi
fi

# create SKR cluster
if k3d cluster list | grep -q "^skr\s"; then
  echo "[$(basename $0)] Reusing existing SKR cluster..."
  else
  k3d cluster create skr \
        -p 10080:80@loadbalancer \
        -p 10443:443@loadbalancer \
        --k3s-arg --tls-san="skr.cluster.local@server:*" \
        --image rancher/k3s:v1.28.7-k3s1 \
        --k3s-arg --disable="traefik@server:*" \
        --k3s-arg --tls-san="host.k3d.internal@server:*" \
        --k3s-arg --tls-san="skr.cluster.local@server:*"
fi

# create KCP cluster
if k3d cluster list | grep -q "^kcp\s"; then
  echo "[$(basename $0)] Reusing existing KCP cluster..."
  else
  k3d cluster create kcp \
        -p 9443:443@loadbalancer \
        -p 9080:80@loadbalancer \
        -p 9081:8080@loadbalancer \
        --registry-create k3d-kcp-registry.localhost:5111 \
        --image rancher/k3s:v1.28.7-k3s1 \
        --k3s-arg --disable="traefik@server:*" \
        --k3s-arg --tls-san="host.k3d.internal@server:*" \
        --k3s-arg --tls-san="skr.cluster.local@server:*"
  
  kubectl config use-context k3d-kcp

  # install istio
  istioctl install --set profile=demo -y

  # install cert-manager
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

  ./add_skr_host_to_coredns.sh

  # create kcp-system namespace
  kubectl create namespace kcp-system

  # label node
  kubectl label nodes k3d-kcp-server-0 iam.gke.io/gke-metadata-server-enabled="true" pool-type=mgmt
fi

# check if .k3d directory exists
if [ ! -d ~/.k3d ]; then
  mkdir ~/.k3d
fi

# export kubeconfigs
k3d kubeconfig get skr > ~/.k3d/skr-local.yaml
k3d kubeconfig get kcp > ~/.k3d/kcp-local.yaml
echo "[$(basename $0)] Kubeconfig for SKR and KCP exported successfully"

echo "[$(basename $0)] Test clusters created successfully"
