# k8s-prometheus betreiben

## Installation

`k8s-prometheus` kann als Komponente über den Komponenten-Operator des CES installiert werden.
Dazu muss eine entsprechende Custom-Resource (CR) für die Komponente erstellt werden.

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
  version: 57.1.1-8
```

Die neue yaml-Datei kann anschließend im Kubernetes-Cluster erstellt werden:

```shell
kubectl apply -f k8s-prometheus.yaml --namespace ecosystem
```

Der Komponenten-Operator erstellt nun die `k8s-prometheus`-Komponente im `ecosystem`-Namespace.

## Upgrade

Zum Upgrade muss die gewünschte Version in der Custom-Resource angegeben werden.
Dazu wird die erstellte CR yaml-Datei editiert und die gewünschte Version eingetragen.
Anschließend die editierte yaml Datei erneut auf den Cluster anwenden:

```shell
kubectl apply -f k8s-prometheus.yaml --namespace ecosystem
```

## Konfiguration

Die Komponente kann über das Feld `spec.valuesYamlOverwrite`. 
Die Konfigurationsmöglichkeiten entsprechen denen vom [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack/values.yaml).
Die Konfiguration für das "kube-prometheus-stack" Helm-Chart muss in der `values.yaml` unter dem Key `kube-prometheus-stack` abgelegt werden.

#### Example valuesYamlOverwrite:
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
  version: 57.1.1-8
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

Prometheus speichert alle Metriken im Volume des zugehörigen StatefulSet. 
Die Größe des Volumes kann in den values unter `kube-prometheus-stack.prometheus.promtheusSpec.storageSpec` angegeben werden (siehe [Beispiel oben](#example-valuesyamloverwrite)).
Die Default-Größe des Volumes ist `10Gi` (10 Gigabyte).

Nach einem Upgrade passt ein Helm-Hook die Größe des Volumes an. Die Annotations für diesen Job
können konfiguriert werden. Es gibt zwei Sets von Annotations, eins für den Job und ein anders
für die Ressourcen, die die Berechtigung für den Job bereitstellen. Die Annotation `helm.sh/hook-weight`
stellt sicher, dass die Berechtigungen für den Job erstellt werden, bevor dieser ausgeführt wird.

Wird der Job erfolgreich beendet, wird der Job und die zugehörige Berechtigung gelöscht.

Für die Verwendung mit ArgoCD müssen die Annotations entsprechend angepasst werden. 
[Hier gibt es eine Anleitung dafür.](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-waves/)

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

Die Retention gibt an für welchen Zeitraum die Metriken gespeichert werden sollen.
Dies kann in den values unter `kube-prometheus-stack.prometheus.promtheusSpec.retention` angegeben werden (siehe [Beispiel oben](#beispiel-valuesyamloverwrite)).
Der Default-Wert für die Rentention ist `10d` (10 Tage).