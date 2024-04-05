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
  prometheusKubeStateMetricsRegistry=$(yq '.kube-prometheus-stack.kube-state-metrics.image.registry' < "${prometheusValues}")
  prometheusKubeStateMetricsRepo=$(yq '.kube-prometheus-stack.kube-state-metrics.image.repository' < "${prometheusValues}")
  prometheusKubeStateMetricsTag=$(yq '.kube-prometheus-stack.kube-state-metrics.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.kubeStateMetrics" "${prometheusKubeStateMetricsRegistry}/${prometheusKubeStateMetricsRepo}:${prometheusKubeStateMetricsTag}"

  local prometheusNodeExporterRegistry
  local prometheusNodeExporterRepo
  local prometheusNodeExporterTag
  prometheusNodeExporterRegistry=$(yq '.kube-prometheus-stack.prometheus-node-exporter.image.registry' < "${prometheusValues}")
  prometheusNodeExporterRepo=$(yq '.kube-prometheus-stack.prometheus-node-exporter.image.repository' < "${prometheusValues}")
  prometheusNodeExporterTag=$(yq '.kube-prometheus-stack.prometheus-node-exporter.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusNodeExporter" "${prometheusNodeExporterRegistry}/${prometheusNodeExporterRepo}:${prometheusNodeExporterTag}"

  local admissionWebhooksPatchRegistry
  local admissionWebhooksPatchRepo
  local admissionWebhooksPatchTag
  admissionWebhooksPatchRegistry=$(yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.registry' < "${prometheusValues}")
  admissionWebhooksPatchRepo=$(yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.repository' < "${prometheusValues}")
  admissionWebhooksPatchTag=$(yq '.kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.admissionWebhooksPatch" "${admissionWebhooksPatchRegistry}/${admissionWebhooksPatchRepo}:${admissionWebhooksPatchTag}"

  local prometheusOperatorRegistry
  local prometheusOperatorRepo
  local prometheusOperatorTag
  prometheusOperatorRegistry=$(yq '.kube-prometheus-stack.prometheusOperator.image.registry' < "${prometheusValues}")
  prometheusOperatorRepo=$(yq '.kube-prometheus-stack.prometheusOperator.image.repository' < "${prometheusValues}")
  prometheusOperatorTag=$(yq '.kube-prometheus-stack.prometheusOperator.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusOperator" "${prometheusOperatorRegistry}/${prometheusOperatorRepo}:${prometheusOperatorTag}"

  local configReloaderRegistry
  local configReloaderRepo
  local configReloaderTag
  configReloaderRegistry=$(yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.registry' < "${prometheusValues}")
  configReloaderRepo=$(yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.repository' < "${prometheusValues}")
  configReloaderTag=$(yq '.kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.configReloader" "${configReloaderRegistry}/${configReloaderRepo}:${configReloaderTag}"

  local prometheusServerRegistry
  local prometheusServerRepo
  local prometheusServerTag
  prometheusServerRegistry=$(yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.registry' < "${prometheusValues}")
  prometheusServerRepo=$(yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.repository' < "${prometheusValues}")
  prometheusServerTag=$(yq '.kube-prometheus-stack.prometheus.prometheusSpec.image.tag' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.prometheusServer" "${prometheusServerRegistry}/${prometheusServerRepo}:${prometheusServerTag}"

  # update version of sidecar-image in values.yaml
  yq -i ".kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image |= sub(\":(([0-9]+)\.([0-9]+)\.([0-9]+)((?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))|(?:\+[0-9A-Za-z-]+))?)\", \":${newReleaseVersion}\")" "${prometheusValues}"

  local authSidecarImage
  authSidecarImage=$(yq '.kube-prometheus-stack.prometheus.prometheusSpec.containers.[0].image' < "${prometheusValues}")
  setAttributeInComponentPatchTemplate ".values.images.authSidecar" "${authSidecarImage}"
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
