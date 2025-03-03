# Entwicklung des Auth Sidecar-Container

Die `k8s-prometheus` Komponente verfügt über einen Sidecar-Container der die folgenden Aufgaben übernimmt:
 * Verwaltung von Service-Accounts: Bereitstellen einer API, die vom `k8s-dogu-operator` verwendet wird.
 * Authentifizierung des Zugangs zur Prometheus-API anhand der Service-Accounts

Die beiden Dienste laufen auf den Ports `8086` (Auth-Proxy) und `8087` (Service-Account API) in denselben Container jeweils als eigener [Gin](https://github.com/gin-gonic/gin) Service.

### Verwaltung von Service-Accounts
Damit der Dogu-Operator die Service-Account API verwenden kann, muss definiert sein, wie diese zu erreichen ist.
Dazu ermittelt der Dogu-Operator den zugehörigen Kubernetes-Service anhand des Labels `ces.cloudogu.com/serviceaccount-provider`.
Dieser Service ist der [sa-provider-svc.yaml](../../k8s/helm/templates/sa-provider-svc.yaml) definiert.

Die weiteren benötigten Informationen sind in den Annotations am Service angegeben:

| Annotation                                    | Beschreibung                                                                          | Wert                                 |
|-----------------------------------------------|---------------------------------------------------------------------------------------|--------------------------------------|
| `ces.cloudogu.com/serviceaccount/port`        | Gibt den Port an, unter dem die ServiceAccount-API verfügbar ist                      | `8080`                               |
| `ces.cloudogu.com/serviceaccount/path`        | Gibt den Pfad an, unter dem die ServiceAccount-API verfügbar ist                      | `/serviceaccounts`                   |
| `ces.cloudogu.com/serviceaccount/secret-name` | Gibt den Namen des Secrets an, dass den API-Key zur Authentifizierung enthält         | `k8s-prometheus-service-account-api` |
| `ces.cloudogu.com/serviceaccount/secret-key`  | Gibt den Key im Secrets an, unter dem der API-Key zur Authentifizierung zu finden ist | `apiKey`                             |


Mit dem `secret-name` und dem `secret-key` kann der Dogu-Operator den ApiKey aus dem angegebenen Secret der Komponente lesen.
Mit diesen Informationen und der IP-Adresse des Services kann der Dogu-Operator Requests an die ServiceAccount-API der jeweiligen Komponente stellen.

Das Secret für den ApiKey wird mit dem Helm-Template [sa-api-secret](../../k8s/helm/templates/sa-api-secret.yaml) erstellt. 

Die erstellten Service-Accounts werden in einer YAML-Datei in dem [Config-Volume](../../k8s/helm/templates/config-pvc.yaml) gespeichert.

### Auth-Proxy
Um den Zugang auf die Prometheus-API abzusichern, wird der eingehende Netzwerkverkehr über eine [NetworkPolicy](../../k8s/helm/templates/network-policy.yaml) blockiert.
Lediglich die eingehende Kommunikation zu den Ports `8086` und `8087` ist erlaubt.

Der Auth-Proxy im Sidecar-Container läuft auf Port `8086` und fungiert als Reverse-Proxy für den Prometheus-Container.
Dabei wird jeder Request über Basic-Auth überprüft. 
Im Basic-Auth-Header müssen die Daten eines vorher erstellten Service-Accounts enthalten sein.

### Konfiguration des Sidecar-Container
Der Sidecar-Container wird in der [values.yaml](../../k8s/helm/values.yaml) unter dem Pfad `kube-prometheus-stack.prometheus.promtheusSpec.containers` konfiguriert.

```yaml
kube-prometheus-stack:
  prometheus:
    prometheusSpec:
      containers:
        - name: auth
          image: cloudogu/k8s-prometheus-auth:2.50.1-1
          imagePullPolicy: Always
          env:
            - name: LOG_LEVEL
              value: "INFO"
            - name: PROMETHEUS_URL
              value: "http://localhost:9090"
            - name: WEB_PRESETS_FILE
              value: "/presets/web.presets.yaml"
            - name: WEB_CONFIG_FILE
              value: "/config/web.config.yaml"
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: k8s-prometheus-service-account-api
                  key: apiKey
          ports:
            - name: auth-proxy
              containerPort: 8086
            - name: sa-provider
              containerPort: 8087
          volumeMounts:
            - name: ces-config
              mountPath: /config
            - name: ces-presets
              readOnly: true
              mountPath: /presets
```

Folgende Umgebungsvariablen können konfiguriert werden:

| Name              | Beschreibung                                                                                       |
|-------------------|----------------------------------------------------------------------------------------------------|
| LOG_LEVEL         | Das zu verwendende Log-Level(`DEBUG`, `INFO`, `WARN`, `ERROR`). Default `INFO`                     |
| PROMETHEUS_URL    | Die URL für den Auth-Proxy unter der Prometheus zu erreichen ist                                   |
| WEB_PRESETS_FILE  | Read-only-Datei in der vorkonfigurierte Service-Accounts gespeichert werden, z.B. aus einem Secret |
| WEB_CONFIG_FILE   | Die Datei in der die Service-Accounts gespeichert werden                                           |
| API_KEY           | Der API-Key für die Service-Account API                                                            |

## Entwicklung im lokalen CES-Cluster
Um die `k8s-prometheus` Komponente im lokalen CES-Cluster zu testen, können die folgenden Make-Targets verwendet werden:

Initial, damit die Komponente installiert werden kann:
```bash
make helm-update-dependencies
```

```bash
# Installiert k8s-prometheus im Cluster
make component-apply
```

```bash
# Löscht k8s-prometheus aus dem Cluster
make component-delete
```

```bash
# Löscht k8s-prometheus aus dem Cluster und installiert es asnschließend erneut
make component-reinstall
```

```bash
# Installiert das Helm-Chart von k8s-prometheus im Cluster (ohne die CES-Komponente)
make helm-apply
```

```bash
# Löscht das Helm-Chart von k8s-prometheus aus dem Cluster (ohne die CES-Komponente)
make helm-delete
```