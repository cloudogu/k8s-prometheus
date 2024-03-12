#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

componentTemplateFile="k8s/helm/component-patch-tpl.yaml"
prometheusValues="k8s/helm/values.yaml"

# this function will be sourced from release.sh and be called from release_functions.sh
update_versions_modify_files() {
  newReleaseVersion="${1}"

  echo "Set images in component patch template"

  local prometheusKubeStateMetricsRegistry
  local prometheusKubeStateMetricsRepo
  local prometheusKubeStateMetricsTag
  prometheusKubeStateMetricsRegistry=$(yq '.prometheus.kube-state-metrics.image.registry' < "${prometheusValues}")
  prometheusKubeStateMetricsRepo=$(yq '.prometheus.kube-state-metrics.image.repository' < "${prometheusValues}")
  prometheusKubeStateMetricsTag=$(yq '.prometheus.kube-state-metrics.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.kubeStateMetrics" "${prometheusKubeStateMetricsRegistry}/${prometheusKubeStateMetricsRepo}:${prometheusKubeStateMetricsTag}"

  local prometheusNodeExporterRegistry
  local prometheusNodeExporterRepo
  local prometheusNodeExporterTag
  prometheusNodeExporterRegistry=$(yq '.prometheus.prometheus-node-exporter.image.registry' < "${prometheusValues}")
  prometheusNodeExporterRepo=$(yq '.prometheus.prometheus-node-exporter.image.repository' < "${prometheusValues}")
  prometheusNodeExporterTag=$(yq '.prometheus.prometheus-node-exporter.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusNodeExporter" "${prometheusNodeExporterRegistry}/${prometheusNodeExporterRepo}:${prometheusNodeExporterTag}"

  local configmapReloaderRepo
  local configmapReloaderTag
  configmapReloaderRepo=$(yq '.prometheus.configmapReload.prometheus.image.repository' < "${prometheusValues}")
  configmapReloaderTag=$(yq '.prometheus.configmapReload.prometheus.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.configmapReloader" "${configmapReloaderRepo}:${configmapReloaderTag}"

  local prometheusServerRepo
  local prometheusServerTag
  prometheusServerRepo=$(yq '.prometheus.server.image.repository' < "${prometheusValues}")
  prometheusServerTag=$(yq '.prometheus.server.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusServer" "${prometheusServerRepo}:${prometheusServerTag}"

  local configInitContainerImage
  configInitContainerImage=$(yq '.prometheus.server.extraInitContainers[0].image' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.configInitContainer" "${configInitContainerImage}"

  # update version of sidecar-image in values.yaml
  yq -i ".prometheus.server.sidecarContainers.serviceaccount.image |= sub(\":(([0-9]+)\.([0-9]+)\.([0-9]+)((?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))|(?:\+[0-9A-Za-z-]+))?)\", \":${newReleaseVersion}\")" "${prometheusValues}"

  local saProviderSidecarImage
  saProviderSidecarImage=$(yq '.prometheus.server.sidecarContainers.serviceaccount.image' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.saProviderSidecar" "${saProviderSidecarImage}"
}

setAttributeInComponentPatchTemplate() {
  local key="${1}"
  local value="${2}"

  yq -i "${key} = \"${value}\"" "${componentTemplateFile}"
}

update_versions_stage_modified_files() {
  git add "${prometheusValues}"
  git add "${componentTemplateFile}"
}
