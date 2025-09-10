ARTIFACT_ID=k8s-prometheus
MAKEFILES_VERSION=10.2.1
VERSION=75.3.5-3

GOTAG=1.22.1
MOCKERY_VERSION=v2.42.2
LINT_VERSION?=v1.57.2
.DEFAULT_GOAL:=help

IMAGE?=cloudogu/${ARTIFACT_ID}-auth:${VERSION}
IMAGE_DEV?=${CES_REGISTRY_HOST}${CES_REGISTRY_NAMESPACE}/${ARTIFACT_ID}-auth
IMAGE_IMPORT_TARGET=image-import
K8S_COMPONENT_TARGET_VALUES = ${HELM_TARGET_DIR}/values.yaml
HELM_POST_GENERATE_TARGETS = helm-values-replace-image-repo

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
	else \
		echo "Setting prod image repo in target values.yaml to ${IMAGE}..." ;\
		$(BINARY_YQ) -i e ".kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image=\"${IMAGE}\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
	fi

.PHONY: prometheus-release
prometheus-release: ${BINARY_YQ} ## Interactively starts the release workflow for k8s-prometheus
	@echo "Starting git flow release..."
	@build/make/release.sh k8s-prometheus

.PHONY: mocks
mocks: ${MOCKERY_BIN} ${MOCKERY_YAML} ## target is used to generate mocks for all interfaces in a project.
	${MOCKERY_BIN}
	@echo "Mocks successfully created."
