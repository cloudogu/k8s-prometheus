ARTIFACT_ID=k8s-prometheus
MAKEFILES_VERSION=9.0.2
VERSION=2.50.1-1

.DEFAULT_GOAL:=help

IMAGE?=cloudogu/${ARTIFACT_ID}:${VERSION}
IMAGE_IMPORT_TARGET=image-import

ADDITIONAL_CLEAN=clean_charts
clean_charts:
	rm -rf ${K8S_HELM_RESSOURCES}/charts

include build/make/variables.mk
include build/make/clean.mk
include build/make/release.mk
include build/make/self-update.mk

##@ Release

include build/make/k8s-component.mk

.PHONY: prometheus-release
loki-release: ## Interactively starts the release workflow for loki
	@echo "Starting git flow release..."
	@build/make/release.sh loki
