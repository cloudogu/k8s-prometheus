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

### Storage & Retention
Prometheus speichert alle Metriken im Volume des zugehörigen StatefulSet. 
Die Größe des Volumes kann in den values unter `kube-prometheus-stack.prometheus.promtheusSpec.storageSpec` angegeben werden (siehe [Beispiel oben](#example-valuesyamloverwrite)).
Die Default-Größe des Volumes ist `4Gi` (4 Gigabyte).

Die Retention gibt an für welchen Zeitraum die Metriken gespeichert werden sollen.
Dies kann in den values unter `kube-prometheus-stack.prometheus.promtheusSpec.retention` angegeben werden (siehe [Beispiel oben](#beispiel-valuesyamloverwrite)).
Der Default-Wert für die Rentention ist `10d` (10 Tage).