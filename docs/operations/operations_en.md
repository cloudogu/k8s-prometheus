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
          retention: 20d
```

### Storage 

Prometheus stores all metrics in the volume of the associated StatefulSet.
The size of the volume can be specified in the values under `kube-prometheus-stack.prometheus.promtheusSpec.storageSpec` (see [example above](#example-valuesyamloverwrite)).
The default size of the volume is `10Gi` (10 gigabytes).

After an upgrade, a Helm hook adjusts the size of the volume. The annotations for this job can be configured.
There are two sets of annotations, one for the job and another for the resources that provide permissions for the job.
The annotation `helm.sh/hook-weight` ensures that the permissions for the job are created before the job runs.

If the job is completed successfully, the job and the associated permissions are deleted.

For use with ArgoCD, the annotations must be adjusted accordingly.
[Here are instructions for doing so.](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-waves/)

```yaml
# default configuration
prometheusStorageResize:
  annotations:
    # annotations for the job
    job:
      "helm.sh/hook": post-upgrade
      "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
    # annotations for the resources that provide authorization; Role, RoleBinding, ServiceAccount
    jobRBAC:
      "helm.sh/hook": post-upgrade
      "helm.sh/hook-weight": "-5"
      "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
```


### Retention

The retention specifies the period for which the metrics are to be saved.
This can be specified in the values under `kube-prometheus-stack.prometheus.promtheusSpec.retention` (see [example above](#example-valuesyamloverwrite)).
The default value for retention is `10d` (10 days).