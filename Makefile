ARTIFACT_ID=k8s-prometheus
MAKEFILES_VERSION=10.10.0
VERSION=87.5.1-3

GOTAG=1.27.4
LINT_VERSION=v2.15.0
MOCKERY_VERSION=v2.53.8
.DEFAULT_GOAL:=help

IMAGE?=cloudogu/${ARTIFACT_ID}-tools:${VERSION}
IMAGE_DEV?=${CES_REGISTRY_HOST}${CES_REGISTRY_NAMESPACE}/${ARTIFACT_ID}-tools
IMAGE_IMPORT_TARGET=image-import
K8S_COMPONENT_TARGET_VALUES = ${HELM_TARGET_DIR}/values.yaml
HELM_POST_GENERATE_TARGETS = helm-values-replace-image-repo

BINARY_AUTH=$(TARGET_DIR)/$(ARTIFACT_ID)
BINARY_AUTH_SRC=cmd/auth/main.go

BINARY_STORAGE_RESIZE=$(TARGET_DIR)/storageresize
BINARY_STORAGE_RESIZE_SRC=cmd/storageresize/main.go

ADDITIONAL_CLEAN=clean_charts
clean_charts:
	rm -rf ${K8S_HELM_RESSOURCES}/charts

include build/make/variables.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/mocks.mk
include build/make/release.mk
include build/make/self-update.mk
include build/make/k8s-component.mk

.PHONY: helm-values-replace-image-repo
helm-values-replace-image-repo: $(BINARY_YQ)
	@if [[ ${STAGE} == "development" ]]; then \
		echo "Setting dev image repo in target values.yaml to ${IMAGE_DEV}:${VERSION}..." ;\
		$(BINARY_YQ) -i e ".kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image=\"${IMAGE_DEV}:${VERSION}\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
		$(BINARY_YQ) -i e ".prometheusStorageResize.image = \"${IMAGE_DEV}:${VERSION}\"" ${K8S_COMPONENT_TARGET_VALUES} ; \
	else \
		echo "Setting prod image repo in target values.yaml to ${IMAGE}..." ;\
		$(BINARY_YQ) -i e ".kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image=\"${IMAGE}\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
		$(BINARY_YQ) -i e ".prometheusStorageResize.image = \"${IMAGE}\"" ${K8S_COMPONENT_TARGET_VALUES} ; \
	fi

.PHONY: prometheus-release
prometheus-release: ${BINARY_YQ} ## Interactively starts the release workflow for k8s-prometheus
	@echo "Starting git flow release..."
	@build/make/release.sh k8s-prometheus

#.PHONY: mocks
#mocks: ${MOCKERY_BIN} ${MOCKERY_YAML} ## target is used to generate mocks for all interfaces in a project.
#	${MOCKERY_BIN}
#	@echo "Mocks successfully created."

compile-generic:
	@echo "Compile binary $(BINARY_AUTH)"
	@$(GO_ENV_VARS) go build -mod=vendor -a -tags netgo,osusergo $(LDFLAGS) -o $(BINARY_AUTH) $(BINARY_AUTH_SRC)
	@echo "Compile binary $(BINARY_STORAGE_RESIZE)"
	@$(GO_ENV_VARS) go build -mod=vendor -a -tags netgo,osusergo $(LDFLAGS) -o $(BINARY_STORAGE_RESIZE) $(BINARY_STORAGE_RESIZE_SRC)