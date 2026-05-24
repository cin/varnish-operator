# Development

This guide covers building, testing, and running the varnish-operator locally.

For installing a released build in a cluster, see [Installation](installation.md).

## Requirements

* **Go 1.26+**
* **Kubernetes 1.29+** cluster and `kubectl` configured ( [kind](https://kind.sigs.k8s.io/), [minikube](https://minikube.sigs.k8s.io/), or an existing cluster)
* **Docker** (or podman) for image builds and e2e tests
* **Helm 3** for deploying the operator chart
* **kind** v0.20+ for end-to-end tests
* **golangci-lint** v2.9+ for `make lint` (must be built with Go 1.26+ to match `go.mod`)
* **operator-sdk** v1.42+ and **yq** for `make bundle` (installed separately)
* **setup-envtest** for controller unit tests (see [Unit tests](#unit-tests))

The Makefile downloads pinned versions of **controller-gen**, **kustomize**, and **goimports** into `./bin/` on first use. You do not need to install kubebuilder.

## Code structure

The project consists of two components:

* **Varnish operator** — watches `VarnishCluster` resources and manages cluster infrastructure (StatefulSet, Services, RBAC, VCL ConfigMap, and so on).
* **Varnish controller** — runs inside each Varnish pod. It watches Kubernetes resources and reloads VCL when backends, VCL files, or cluster membership change.

Both components share one repository, dependencies, and build tooling.

| Component | Source | Entry point |
| --------- | ------ | ----------- |
| Operator | `pkg/varnishcluster/` | `cmd/varnish-operator/main.go` |
| Varnish controller | `pkg/varnishcontroller/` | `cmd/varnish-controller/main.go` |

Kubebuilder/operator-sdk scaffolding lives under `config/` (CRDs, RBAC, bundle manifests). The Helm chart is in `varnish-operator/`.

## Clone and setup

```bash
git clone https://github.com/cin/varnish-operator.git
cd varnish-operator
go mod download
```

## Running the operator locally

Running the operator on your machine against a real cluster is the fastest way to iterate on operator code. Your local kubeconfig credentials are used instead of in-cluster RBAC.

### Install the CRD

```bash
make install
```

This applies the `VarnishCluster` CRD (`varnishclusters.caching.ibm.com`). Re-run after CRD schema changes.

Verify:

```bash
kubectl get crd varnishclusters.caching.ibm.com
```

### Start the operator

```bash
make run
```

By default the operator watches the `default` namespace. Override with:

```bash
NAMESPACE=varnish-operator make run
```

`make run` sets `LEADERELECTION_ENABLED=false` and `WEBHOOKS_ENABLED=false` for simpler local development. The coupled Varnish image defaults to `cinple/varnish:local-dev` via the `REPO` and `VARNISH_IMG` Makefile variables.

After code changes, stop the process and run `make run` again.

### What local run cannot test

Some behavior only works when the operator runs as a pod with its ServiceAccount and webhooks enabled:

* Validating and mutating webhooks (`WEBHOOKS_ENABLED=true`, plus TLS certs)
* In-cluster RBAC (local run uses your kubeconfig user, not the operator ClusterRole)

For those cases, build and deploy the operator image (below).

## Deploying the operator in a cluster

Build and load the operator image, regenerate manifests, and install with Helm:

```bash
# Build operator image (runs unit tests first)
make docker-build REPO=cinple VERSION=local

# Regenerate CRD and ClusterRole in the Helm chart
make manifests

# Install or upgrade via Helm
helm upgrade --install varnish-operator ./varnish-operator \
  --namespace varnish-operator --create-namespace \
  --set container.registry=cinple \
  --set container.repository=varnish-operator \
  --set container.tag=local
```

Check logs:

```bash
kubectl logs -n varnish-operator -l app=varnish-operator --tail=50
```

## Developing the varnish controller and sidecar images

Varnish pods can only be exercised in Kubernetes. After changing varnish-controller, varnishd, or metrics-exporter code, rebuild the relevant image and point your `VarnishCluster` at it.

Build all pod images (operator image is separate):

```bash
make docker-build-pod REPO=cinple VERSION=local
```

This produces:

| Image | Dockerfile |
| ----- | ---------- |
| `cinple/varnish:local` | `Dockerfile.varnishd` |
| `cinple/varnish-controller:local` | `Dockerfile.controller` |
| `cinple/varnish-metrics-exporter:local` | `Dockerfile.exporter` |

Override images in your `VarnishCluster`:

```yaml
spec:
  varnish:
    image: cinple/varnish:local
    controller:
      image: cinple/varnish-controller:local
    metricsExporter:
      image: cinple/varnish-metrics-exporter:local
```

If you reuse the same tag, set `spec.statefulSet.container.imagePullPolicy: Always` and restart pods (or delete them) so Kubernetes pulls the new layers.

The metrics exporter image accepts `PROMETHEUS_VARNISH_EXPORTER_VERSION` as a build argument:

```bash
docker build --build-arg PROMETHEUS_VARNISH_EXPORTER_VERSION=1.6.1 \
  -t my-exporter:local -f Dockerfile.exporter .
```

## Code generation and manifests

```bash
make generate   # deepcopy and other generated code
make manifests  # CRD + ClusterRole into varnish-operator/
make fmt        # goimports (pinned in ./bin/)
make vet        # go vet
make lint       # golangci-lint
```

## Kubernetes versions in tests

Unit tests and end-to-end tests pin Kubernetes versions differently. **Do not reuse an e2e/kind version string for envtest** (or vice versa)—`setup-envtest use 1.35.1` fails with `unable to find archive` because envtest does not publish that tag.

| | Unit tests | End-to-end tests |
| --- | --- | --- |
| **Tool** | [envtest](https://book.kubebuilder.io/reference/envtest.html) via `setup-envtest` | [kind](https://kind.sigs.k8s.io/) `kindest/node` images |
| **Tag format** | Minor releases from [controller-tools envtest](https://github.com/kubernetes-sigs/controller-tools/releases) (e.g. `1.36.0`, `1.35.0`, `1.34.1`) | Patch tags published on Docker Hub (e.g. `1.35.1`, `1.34.3`) |
| **CI versions** | **1.36.0** (matches `k8s.io/*` v0.36 in `go.mod`) | **1.34.3**, **1.35.1** |
| **Pick a version** | `setup-envtest list --platform linux/amd64` | [kindest/node tags](https://hub.docker.com/r/kindest/node/tags) or `docker pull kindest/node:vX.Y.Z` |

Why they differ:

* **envtest** ships pre-built API server/etcd binaries per controller-tools release. Tags are sparse (often `X.Y.0` or `X.Y.1`), and newer tags may exist before a matching kind image is published.
* **kind** needs a full node OCI image. CI only uses tags that exist on Docker Hub; unreleased versions (e.g. `1.36.0` at the time of writing) require `kind build node-image` locally via `hack/create_dev_cluster.sh`.

## Unit tests

Controller tests use envtest (a local Kubernetes API server). Install setup-envtest and point `KUBEBUILDER_ASSETS` at the binaries for an envtest release (see [Kubernetes versions in tests](#kubernetes-versions-in-tests)):

```bash
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
export KUBEBUILDER_ASSETS="$(setup-envtest use 1.36.0 -p path)"
make test
```

Without `KUBEBUILDER_ASSETS`, the envtest-based controller suite will fail to start.

## End-to-end tests

E2e tests create a kind cluster, build all images, install the operator from the local Helm chart, and run tests in `./tests`.

```bash
make e2e-tests
```

This runs `hack/create_dev_cluster.sh`, executes tests with `KUBECONFIG=./e2e-tests-kubeconfig`, then tears the cluster down.

Use a specific Kubernetes version (must be a valid `kindest/node` tag—see [Kubernetes versions in tests](#kubernetes-versions-in-tests)):

```bash
KUBERNETES_VERSION=1.35.1 make e2e-tests
```

For versions without a pre-built `kindest/node` image, the dev script builds the node image locally with `kind build node-image`.

The helper script builds images as `cinple/*:local` and sets `imagePullPolicy=Never` so kind can use locally built images.

Manual workflow:

```bash
./hack/create_dev_cluster.sh
KUBECONFIG=./e2e-tests-kubeconfig go test ./tests
./hack/delete_dev_cluster.sh
```

Optional flags for `create_dev_cluster.sh`:

* `-s` — skip Docker build (images must already exist locally)
* `-v` — create a sample `VarnishCluster`
* `-b` — create nginx backend deployments

CI runs e2e against Kubernetes **1.34.3** and **1.35.1** (see [Kubernetes versions in tests](#kubernetes-versions-in-tests)).

## OperatorHub bundle generation

Bundles are generated with [operator-sdk](https://sdk.operatorframework.io/). Source manifests live under `config/` (CRD, RBAC, samples, ClusterServiceVersion).

```bash
# Semver bundle version; use any tag for the container image
VERSION=0.37.0 make bundle

# Local/dev image tag maps to bundle version 0.0.0-local
VERSION=local make bundle
```

Output is written to `./$(VERSION)/` (for example `./local/`). The target validates the bundle, copies `Dockerfile.bundle`, and replaces any previous output directory with the same name.

Review the generated manifests before publishing. Bundles can be tested with the [community-operators testing guide](https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md).

## Useful Makefile variables

| Variable | Default | Purpose |
| -------- | ------- | ------- |
| `VERSION` | `local` | Image tag suffix |
| `REPO` | `cinple` | Container registry / namespace prefix |
| `NAMESPACE` | `default` | Namespace for `make run` and `helm-upgrade` |
| `PLATFORM` | `linux/amd64` | Docker build platform |

Examples:

```bash
make docker-build REPO=myregistry VERSION=dev
make run NAMESPACE=varnish-operator REPO=myregistry
```
