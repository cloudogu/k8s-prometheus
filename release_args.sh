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
  prometheusKubeStateMetricsRegistry=$(./.bin/yq '.kube-prometheus-stack.kube-state-metrics.image.registry' < "${prometheusValues}")
  prometheusKubeStateMetricsRepo=$(./.bin/yq '.kube-prometheus-stack.kube-state-metrics.image.repository' < "${prometheusValues}")
  prometheusKubeStateMetricsTag=$(./.bin/yq '.kube-prometheus-stack.kube-state-metrics.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.kubeStateMetrics" "${prometheusKubeStateMetricsRegistry}/${prometheusKubeStateMetricsRepo}:${prometheusKubeStateMetricsTag}"

  local prometheusNodeExporterRegistry
  local prometheusNodeExporterRepo
  local prometheusNodeExporterTag
  prometheusNodeExporterRegistry=$(./.bin/yq '.kube-prometheus-stack.prometheus-node-exporter.image.registry' < "${prometheusValues}")
  prometheusNodeExporterRepo=$(./.bin/yq '.kube-prometheus-stack.prometheus-node-exporter.image.repository' < "${prometheusValues}")
  prometheusNodeExporterTag=$(./.bin/yq '.kube-prometheus-stack.prometheus-node-exporter.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusNodeExporter" "${prometheusNodeExporterRegistry}/${prometheusNodeExporterRepo}:${prometheusNodeExporterTag}"

  local admissionWebhooksPatchRegistry
  local admissionWebhooksPatchRepo
  local admissionWebhooksPatchTag
  admissionWebhooksPatchRegistry=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.registry' < "${prometheusValues}")
  admissionWebhooksPatchRepo=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.repository' < "${prometheusValues}")
  admissionWebhooksPatchTag=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.admissionWebhooksPatch" "${admissionWebhooksPatchRegistry}/${admissionWebhooksPatchRepo}:${admissionWebhooksPatchTag}"

  local prometheusOperatorRegistry
  local prometheusOperatorRepo
  local prometheusOperatorTag
  prometheusOperatorRegistry=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.image.registry' < "${prometheusValues}")
  prometheusOperatorRepo=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.image.repository' < "${prometheusValues}")
  prometheusOperatorTag=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusOperator" "${prometheusOperatorRegistry}/${prometheusOperatorRepo}:${prometheusOperatorTag}"

  local configReloaderRegistry
  local configReloaderRepo
  local configReloaderTag
  configReloaderRegistry=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.registry' < "${prometheusValues}")
  configReloaderRepo=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.repository' < "${prometheusValues}")
  configReloaderTag=$(./.bin/yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.configReloader" "${configReloaderRegistry}/${configReloaderRepo}:${configReloaderTag}"

  local prometheusServerRegistry
  local prometheusServerRepo
  local prometheusServerTag
  prometheusServerRegistry=$(./.bin/yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.registry' < "${prometheusValues}")
  prometheusServerRepo=$(./.bin/yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.repository' < "${prometheusValues}")
  prometheusServerTag=$(./.bin/yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusServer" "${prometheusServerRegistry}/${prometheusServerRepo}:${prometheusServerTag}"

  # update version of sidecar-image in values.yaml
  ./.bin/yq -i ".kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image |= sub(\":(([0-9]+)\.([0-9]+)\.([0-9]+)((?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))|(?:\+[0-9A-Za-z-]+))?)\", \":${newReleaseVersion}\")" "${prometheusValues}"

  local authSidecarImage
  authSidecarImage=$(./.bin/yq '.kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.authSidecar" "${authSidecarImage}"
}

setAttributeInComponentPatchTemplate() {
  local key="${1}"
  local value="${2}"

  ./.bin/yq -i "${key} = \"${value}\"" "${componentTemplateFile}"
}

update_versions_stage_modified_files() {
  git add "${prometheusValues}"
  git add "${componentTemplateFile}"
}
