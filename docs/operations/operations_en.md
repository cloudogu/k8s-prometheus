# operate k8s-prometheus

## Installation

`k8s-prometheus` can be installed as a component via the CES component operator.
To do this, a corresponding custom resource (CR) must be created for the component.

```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: k8s-prometheus
  labels:
    app: ces
spec:
  name: k8s-prometheus
  namespace: k8s
  version: 2.50.1-1
```

The new yaml file can then be created in the Kubernetes cluster:

```shell
kubectl apply -f k8s-prometheus.yaml --namespace ecosystem
```

The component operator now creates the `k8s-prometheus` component in the `ecosystem` namespace.

## Upgrade

To upgrade, the desired version must be specified in the custom resource.
To do this, the CR yaml file created is edited and the desired version is entered.
Then reapply the edited yaml file to the cluster:

```shell
kubectl apply -f k8s-prometheus.yaml --namespace ecosystem
```

## Configuration

The component can be configured via the `spec.valuesYamlOverwrite` field.
The configuration options correspond to those of [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack/values.yaml).
The configuration for the "kube-prometheus-stack" Helm chart must be stored in `values.yaml` under the key `kube-prometheus-stack`.

**Example:**
```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: k8s-prometheus
  labels:
    app: ces
spec:
  name: k8s-prometheus
  namespace: k8s
  version: 2.50.1-1
  valuesYamlOverwrite: |
    kube-prometheus-stack:
      prometheus:
        prometheusSpec:
          storageSpec:
            volumeClaimTemplate:
              spec:
                resources:
                  requests:
                    storage: 8Gi
```

