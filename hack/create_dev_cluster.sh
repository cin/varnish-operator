#!/bin/bash

set -ex

kube_version="1.35.1" # see https://hub.docker.com/r/kindest/node/tags for published images
if [ -n "${KUBERNETES_VERSION}" ]; then
  kube_version="${KUBERNETES_VERSION}"
fi

# Ensure kindest/node:v<version> exists locally; build if no pre-built image is published yet.
function ensure_kind_node_image() {
  local version="$1"
  local image="kindest/node:v${version}"

  if docker image inspect "${image}" >/dev/null 2>&1; then
    echo "${image}"
    return 0
  fi

  if docker pull "${image}" >/dev/null 2>&1; then
    echo "${image}"
    return 0
  fi

  echo "Pre-built ${image} not found; building locally with kind (may take a few minutes)..." >&2
  kind build node-image "v${version}"
  echo "${image}"
}

if ! which docker; then
    echo -e "Install docker first"
    exit 1
fi

if ! which kind >/dev/null; then
    echo -e "Install kind first"
    exit 1
fi

if ! which helm >/dev/null; then
    echo -e "Install helm first"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
KUBECONFIG_FILE="${ROOT_DIR}/e2e-tests-kubeconfig"

function e2e_kubeconfig_valid() {
  kubectl config current-context --kubeconfig="${KUBECONFIG_FILE}" >/dev/null 2>&1
}

function refresh_e2e_kubeconfig_from_kind() {
  if kind get clusters 2>/dev/null | grep -qx "${cluster_name}"; then
    kind get kubeconfig --name "${cluster_name}" > "${KUBECONFIG_FILE}"
    return 0
  fi
  return 1
}

function use_e2e_kubeconfig() {
  if [[ -f "${KUBECONFIG_FILE}" ]] && ! e2e_kubeconfig_valid; then
    echo "warning: ${KUBECONFIG_FILE} is empty or stale; refreshing from kind cluster ${cluster_name}..." >&2
    rm -f "${KUBECONFIG_FILE}"
  fi
  if [[ ! -f "${KUBECONFIG_FILE}" ]]; then
    if ! refresh_e2e_kubeconfig_from_kind; then
      echo "error: no valid kubeconfig at ${KUBECONFIG_FILE} and kind cluster '${cluster_name}' is not running." >&2
      echo "Run: ${SCRIPT_DIR}/create_dev_cluster.sh   # without -v or -b" >&2
      exit 1
    fi
  fi
  if ! e2e_kubeconfig_valid; then
    echo "error: ${KUBECONFIG_FILE} is not a usable kubeconfig." >&2
    exit 1
  fi
  export KUBECONFIG="${KUBECONFIG_FILE}"
}

varnish_namespace="varnish-operator"
cluster_name="e2e-tests"
repo="cinple"
build_args="--build-arg VERSION=local"
platform="linux/amd64"
podman_in_use=false
ignore_podman=false
manage_cluster=true
use_buildx=false
create_vc=false
create_backends=false
skip_docker_build=false
dry_run=false

function usage {
  cat << !
USAGE: $0 [-c cluster] [-n namespace] [-p platform] [-r repo] [-b] [-s] [-v] [-x]

Creates a dev cluster and varnish-operator install

-c|--cluster                | cluster
-n|--namespace              | namespace
-p|--platform               | platform (not validated so know which build you're calling)
-r|--repo                   | CR repository
-b|--backends               | create nginx backends (requires cluster; run script without -v/-b first)
-s|--skip-docker-build      | skip docker build
-v|--create-varnishcluster  | create sample VarnishCluster (requires cluster; run script without -v/-b first)
-x|--ignore-podman          | ignore podman's presence
!
}

function default_vc_namespace {
  use_e2e_kubeconfig
  if [[ "$varnish_namespace" == "varnish-operator" ]]; then
    if [ "$(kubectl get namespace --no-headers | grep varnish-cluster | wc -l | xargs echo -n)" -eq 0 ]; then
      kubectl create namespace varnish-cluster
    fi
    varnish_namespace="varnish-cluster"
  fi
}

function load_local_images_into_kind() {
  use_e2e_kubeconfig
  if ! kind get clusters 2>/dev/null | grep -qx "${cluster_name}"; then
    echo "error: kind cluster '${cluster_name}' not found." >&2
    exit 1
  fi
  for image in "${workload_images[@]}"; do
    if ! docker image inspect "${image}" >/dev/null 2>&1; then
      echo "error: local image ${image} not found. Run ${SCRIPT_DIR}/create_dev_cluster.sh (without -v/-b) to build images first." >&2
      exit 1
    fi
    kind load docker-image --name "${cluster_name}" "${image}"
  done
}

function create_nginx_backends {
  if [ "$dry_run" = true ]; then
    echo "dry-run: would otherwise be installing nginx"
    return 0
  fi
  default_vc_namespace
  kubectl create deployment nginx-backend --image nginx -n $varnish_namespace --port=80
}

function create_varnishcluster {
  if [ "$dry_run" = true ]; then
    echo "dry-run: would otherwise be installing varnishcluster"
    return 0
  fi

  default_vc_namespace
  load_local_images_into_kind
  cat <<EOF | kubectl create -f -
apiVersion: caching.ibm.com/v1alpha1
kind: VarnishCluster
metadata:
  name: varnishcluster-example
  namespace: $varnish_namespace
spec:
  vcl:
   configMapName: vcl-config
   entrypointFileName: entrypoint.vcl
  varnish:
    image: ${repo}/varnish:local
    imagePullPolicy: Never
    controller:
      image: ${repo}/varnish-controller:local
      imagePullPolicy: Never
    metricsExporter:
      image: ${repo}/varnish-metrics-exporter:local
      imagePullPolicy: Never
  backend:
    port: 80
    selector:
      app: nginx-backend
  service:
    port: 80 # Varnish pods will be exposed on that port
EOF
}

while (( "$#" )); do
  opt="$1"; shift;
  case "$opt" in
    "-b"|"--backends") create_backends=true;;
    "-d"|"--dry-run") dry_run=true;;
    "-s"|"--skip-docker-build") skip_docker_build=true;;
    "-v"|"--create-varnishcluster") create_vc=true;;
    "-x"|"--ignore-podman") ignore_podman=true;;
    "-c"|"--cluster") cluster_name="$1"; manage_cluster=false; shift;;
    "-n"|"--namespace") varnish_namespace="$1"; shift;;
    "-p"|"--platform") platform="$1"; shift;;
    "-r"|"--repo") repo="$1"; shift;;
    *) echo "invalid option: \""$opt"\"" >&2; usage; exit 1;;
  esac
done

cd "${ROOT_DIR}"

container_image="${repo}/varnish-operator:local"
workload_images=("${repo}/varnish:local" "${repo}/varnish-controller:local" "${repo}/varnish-metrics-exporter:local")
images=("${container_image}" "${workload_images[@]}")

if [ "$create_vc" = true ]; then
  create_varnishcluster
  exit 0
fi

if [ "$create_backends" = true ]; then
  create_nginx_backends
  exit 0
fi

dockerfiles=(Dockerfile Dockerfile.varnishd Dockerfile.controller Dockerfile.exporter)

if [[ $platform =~ ^.*,.*$ ]]; then
  use_buildx=true
  build_args="buildx build $build_args --platform $platform"
else
  build_args="build $build_args --platform $platform"
fi

if [ "$ignore_podman" = false ] && which podman >/dev/null; then
  podman_in_use=true
elif [ "$use_buildx" = true ]; then
  build_args="$build_args --push"
fi

if [ "$dry_run" = true ]; then
  echo "build_args: $build_args, manage_cluster: $manage_cluster; use_buildx: $use_buildx, podman_in_use: $podman_in_use, ignore_podman: $ignore_podman"
  exit 0
fi

if [ "$manage_cluster" = true ]; then
  kind_node_image="$(ensure_kind_node_image "${kube_version}")"
  kind delete cluster --name $cluster_name > /dev/null 2>&1
  kind create cluster --name $cluster_name --image "${kind_node_image}" --kubeconfig "${KUBECONFIG_FILE}"
fi

use_e2e_kubeconfig

if [ "$(kubectl get namespace --no-headers | grep varnish-operator | wc -l | xargs echo -n)" -eq 0 ]; then
  kubectl create ns $varnish_namespace
fi

# if skipping docker build, ensure all the images are at least in the local docker registry
if [ "$skip_docker_build" = true ]; then
  set +e
  for image in "${images[@]}"; do
    res=$(docker image inspect $image)
    if [ "$?" -ne 0 ]; then
      echo "missing $image. cannot skip docker build"
      skip_docker_build=false
    fi
  done
  set -e
fi

if [ "$skip_docker_build" = false ]; then
  for ((i=0; i<${#images[@]}; i++)); do
    docker $build_args -f ${dockerfiles[$i]} -t ${images[$i]} .
  done

  if [ "$use_buildx" = true ] && [ "$podman_in_use" = false ]; then
    for image in "${images[@]}"; do
      docker pull $image
    done
  fi
fi

for image in "${images[@]}"; do
  kind load docker-image --name "${cluster_name}" "${image}"
done

helm install varnish-operator varnish-operator --namespace=$varnish_namespace --wait --set container.imagePullPolicy=Never --set container.image=$container_image
