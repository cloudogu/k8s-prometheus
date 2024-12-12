# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v57.1.1-7] - 2024-12-12
### Added
- [#14] add network policies for prometheus and kube-stack-promotheus related components

### Fixed
- [#12] building component locally by using component-apply target possible again.

## [v57.1.1-6] - 2024-10-28
### Changed
- [#10] Use `ces-container-registries` secret for pulling container images as default.

## [v57.1.1-5] - 2024-09-19
### Changed
- Relicense to AGPL-3.0-only

## [v57.1.1-4] - 2024-09-16
### Fixed
- [#6] Use `crypto/rand` instead of `math/rand` for generating passwords.

## [v57.1.1-3] - 2024-06-28
### Changed
- [#4] Changed docker tag from `k8s-prometheus-service-account-provider` to `k8s-prometheus-auth`.

## [v57.1.1-2] - 2024-04-23
### Added
- [#2] Add missing 'node' labels for metrics from node-exporter

## [v57.1.1-1] - 2024-04-16
- initial release