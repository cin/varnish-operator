# Image URL to use in all building/pushing image targets
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION ?= local
REPO ?= cinple
PUBLISH_IMG ?= varnish-operator:${VERSION}
IMG ?= ${PUBLISH_IMG}-dev
VARNISH_PUBLISH_IMG ?= varnish:${VERSION}
VARNISH_IMG ?= ${VARNISH_PUBLISH_IMG}-dev
VARNISH_CONTROLLER_PUBLISH_IMG ?= varnish-controller:${VERSION}
VARNISH_CONTROLLER_IMG ?= ${VARNISH_CONTROLLER_PUBLISH_IMG}-dev
VARNISH_METRICS_PUBLISH_IMG ?= varnish-metrics-exporter:${VERSION}
VARNISH_METRICS_IMG ?= ${VARNISH_METRICS_PUBLISH_IMG}-dev
NAMESPACE ?= "default"
CRD_OPTIONS ?= "crd:crdVersions=v1"
PLATFORM ?= "linux/amd64"

# operator-sdk generate bundle requires semver; VERSION=local is fine for images but not for bundles.
ifeq ($(VERSION),local)
BUNDLE_VERSION ?= 0.0.0-local
else
BUNDLE_VERSION ?= $(VERSION)
endif

# all: test varnish-operator
all: test varnish-operator varnish-controller

# Run tests
test: generate fmt vet manifests
	go test github.com/cin/varnish-operator/pkg/... github.com/cin/varnish-operator/cmd/... github.com/cin/varnish-operator/api/... -coverprofile cover.out

# Run lint tools
lint:
	golangci-lint run

# Build varnish-operator binary
varnish-operator: generate fmt vet
	go build -o ${ROOT_DIR}bin/varnish-operator github.com/cin/varnish-operator/cmd/varnish-operator

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	NAMESPACE=${NAMESPACE} LOGLEVEL=debug LOGFORMAT=console CONTAINER_IMAGE=${REPO}/${VARNISH_IMG} LEADERELECTION_ENABLED=false WEBHOOKS_ENABLED=false go run ${ROOT_DIR}cmd/varnish-operator/main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
install: manifests
	kustomize build ${ROOT_DIR}config/crd | kubectl apply -f -

uninstall:
	kustomize build ${ROOT_DIR}config/crd | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	# CRD apiextensions.k8s.io/v1
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=varnish-operator paths="./..." output:crd:artifacts:config=config/crd/bases
	kustomize build ${ROOT_DIR}config/crd > $(ROOT_DIR)varnish-operator/crds/varnishcluster.yaml

	# ClusterRole
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=varnish-operator paths="./..." output:crd:none output:rbac:stdout > $(ROOT_DIR)varnish-operator/templates/clusterrole.yaml

# Run goimports against code
GOIMPORTS = $(shell pwd)/bin/goimports
goimports:
	$(call go-get-tool,$(GOIMPORTS),golang.org/x/tools/cmd/goimports@v0.45.0)

fmt: goimports
	cd ${ROOT_DIR} && $(GOIMPORTS) -w ./pkg ./cmd ./api

# Run go vet against code
vet:
	cd ${ROOT_DIR} && go vet ./pkg/... ./cmd/... ./api/...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

helm-upgrade: manifests
	helm upgrade --install --namespace ${NAMESPACE} --force varnish-operator --set operator.controllerImage.tag=${VERSION} --set namespace=${NAMESPACE} ${ROOT_DIR}varnish-operator

# Build the docker image
# docker-build: test
docker-build: test
	docker build --platform ${PLATFORM} ${ROOT_DIR} -t ${IMG} -f Dockerfile

# Tag and push the docker image
docker-tag-push:
ifndef PUBLISH
	docker tag ${IMG} ${REPO}/${IMG}
	docker push ${REPO}/${IMG}
else
	docker tag ${IMG} ${REPO}/${PUBLISH_IMG}
	docker push ${REPO}/${PUBLISH_IMG}
endif

varnish-controller: fmt vet
	go build -o ${ROOT_DIR}bin/varnish-controller ${ROOT_DIR}cmd/varnish-controller/

# Build the docker image with varnishd itself and varnish modules
docker-build-varnish:
	docker build --platform ${PLATFORM} ${ROOT_DIR} -t ${VARNISH_IMG} -f Dockerfile.varnishd

docker-tag-push-varnish:
ifndef PUBLISH
	docker tag ${VARNISH_IMG} ${REPO}/${VARNISH_IMG}
	docker push ${REPO}/${VARNISH_IMG}
else
	docker tag ${VARNISH_IMG} ${REPO}/${VARNISH_PUBLISH_IMG}
	docker push ${REPO}/${VARNISH_PUBLISH_IMG}
endif

# Build the docker image with varnish controller
docker-build-varnish-controller: fmt vet
	docker build --platform ${PLATFORM} ${ROOT_DIR} -t ${VARNISH_CONTROLLER_IMG} -f Dockerfile.controller

docker-tag-push-varnish-controller:
ifndef PUBLISH
	docker tag ${VARNISH_CONTROLLER_IMG} ${REPO}/${VARNISH_CONTROLLER_IMG}
	docker push ${REPO}/${VARNISH_CONTROLLER_IMG}
else
	docker tag ${VARNISH_CONTROLLER_IMG} ${REPO}/${VARNISH_CONTROLLER_PUBLISH_IMG}
	docker push ${REPO}/${VARNISH_CONTROLLER_PUBLISH_IMG}
endif

# Build the docker image with varnish metrics exporter
PROMETHEUS_VARNISH_EXPORTER_VERSION ?= v1.8.3
docker-build-varnish-exporter:
	docker build --platform ${PLATFORM} ${ROOT_DIR} -t ${VARNISH_METRICS_IMG} -f Dockerfile.exporter \
		--build-arg PROMETHEUS_VARNISH_EXPORTER_VERSION=${PROMETHEUS_VARNISH_EXPORTER_VERSION}

docker-tag-push-varnish-exporter:
ifndef PUBLISH
	docker tag ${VARNISH_METRICS_IMG} ${REPO}/${VARNISH_METRICS_IMG}
	docker push ${REPO}/${VARNISH_METRICS_IMG}
else
	docker tag ${VARNISH_METRICS_IMG} ${REPO}/${VARNISH_METRICS_PUBLISH_IMG}
	docker push ${REPO}/${VARNISH_METRICS_PUBLISH_IMG}
endif

docker-build-pod: docker-build-varnish docker-build-varnish-exporter docker-build-varnish-controller
docker-tag-push-pod: docker-tag-push-varnish docker-tag-push-varnish-exporter docker-tag-push-varnish-controller

# find or download controller-gen
# download controller-gen if necessary
CONTROLLER_GEN ?= $(shell go env GOPATH)/bin/controller-gen

controller-gen:
	@test -s $(CONTROLLER_GEN) || go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.21.0

e2e-tests:
	bash $(ROOT_DIR)hack/create_dev_cluster.sh
	KUBECONFIG=$(ROOT_DIR)e2e-tests-kubeconfig go test ./tests
	bash $(ROOT_DIR)hack/delete_dev_cluster.sh

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5@v5.8.1)

# go-get-tool fetches a tool binary via `go install` into the project bin directory.
define go-get-tool
@[ -f $(1) ] || { \
set -e; \
TMP_DIR=$$(mktemp -d); \
cd $$TMP_DIR; \
go mod init tmp; \
GOBIN=$(shell pwd)/bin go install $(2); \
rm -rf $$TMP_DIR; \
}
endef

# operator-sdk v1.42.x is required for bundle generation (install separately).
# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	yq -i '(.spec.template.spec.containers[0].env[] | select(.name == "CONTAINER_IMAGE") | .value) = "$(PUBLISH_IMG)"' config/manager/deployment.yaml
	yq -i '.metadata.annotations.containerImage = "$(PUBLISH_IMG)"' config/manifests/bases/varnish-operator.clusterserviceversion.yaml
	yq -i '.metadata.annotations.createdAt = now' config/manifests/bases/varnish-operator.clusterserviceversion.yaml
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(PUBLISH_IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle
	cp Dockerfile.bundle ./bundle/Dockerfile
	rm -rf ./$(VERSION)
	mv ./bundle ./$(VERSION)
