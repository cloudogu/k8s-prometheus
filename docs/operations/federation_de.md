# Federation

Federation ist eine Funktion, mit der Sie Metriken aus Ihrer Prometheus-Instanz für eine andere Prometheus-Instanz verfügbar machen können.

## Anwendungsfall: Sammeln von Metriken aus CES-Instanzen in einem zentralen Prometheus

In diesem Anwendungsfall haben wir möglicherweise mehrere CES-Instanzen und möchten deren Metriken in einem zentralen Prometheus sammeln.

Sie können dies für beliebig viele CES-Instanzen tun.

### Konfigurieren von CES-Prometheus

Zuerst müssen wir unser CES-Prometheus außerhalb des Clusters verfügbar machen.
Am einfachsten geht dies, indem man über einen Ingress eine Route dorthin erstellt.
Dies kann mit dem folgenden `valuesYamlOverwrite` erfolgen:
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
    global:
      ingress:
        enabled: true
```

Da sich CES-Prometheus hinter einem Authentifizierungsproxy befindet, müssen wir einen Service-Account erstellen, indem wir einen Secret wie das Folgende anwenden:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8s-prometheus-auth-presets
  labels:
    app: ces
    k8s.cloudogu.com/component.name: k8s-prometheus
stringData:
  web.presets.yaml: |
    basic_auth_users:
      prometheus-exposed: bcrypt-hashed-password
```

Stellen Sie sicher, dass das Passwort mit bcrypt gehasht wird.

Starten Sie Prometheus neu, damit der Auth-Proxy die Datei lädt.

Testen Sie die Authentifizierung und die Föderation mit diesem Befehl:
```shell
wget --header="Content-Type: application/json" --header="Accept: application/json" \
  --auth-no-challenge --user=prometheus-exposed --ask-password \
  -O- "https://<your ces-fqdn>/prometheus/federate?match[]=%7B__name__%3D~%22.%2B%22%7D"
```

### Konfigurieren Sie das zentralisierte Prometheus

Sie können dann einen Ausschnitt wie diesen an die `prometheus.yaml` Ihres zentralisierten Prometheus anhängen, um Metriken aus dem CES-Prometheus zu sammeln:
```yaml
  - job_name: 'federate-ces'
    scrape_interval: 15s
    honor_labels: true
    metrics_path: '/prometheus/federate'
    params:
      'match[]':
        - '{__name__=~".+"}'
    static_configs:
      - targets: ['<your ces-fqdn>']
    scheme: https
    basic_auth:
      username: 'prometheus-exposed'
      password_file: '/path/to/your/password-file'
```

# Federation

Federation is a feature that enables you to expose metrics from your Prometheus instance to another Prometheus instance.

## Use-Case: Collecting metrics from CES-instances in a centralized Prometheus

In this use-case, we may have multiple CES-instances and want to collect their metrics in a centralized Prometheus.

You can do this for as many CES-instances as you like.

### Configure CES-Prometheus

First we'll have to expose our CES-Prometheus to outside the cluster.
The easiest way to do this is to create a route to it via an Ingress.
This can be done using the following `valuesYamlOverwrite`:
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
    global:
      ingress:
        enabled: true
```

Since the CES-Prometheus is behind an authentication proxy, we have to create a service account by applying a Secret like the following:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8s-prometheus-auth-presets
  labels:
    app: ces
    k8s.cloudogu.com/component.name: k8s-prometheus
stringData:
  web.presets.yaml: |
    basic_auth_users:
      prometheus-exposed: bcrypt-hashed-password
```

Make sure to hash the password with bcrypt.

Restart the Prometheus to make the auth-proxy load the file.

Test authentication and federation with this:
```shell
wget --header="Content-Type: application/json" --header="Accept: application/json" \
  --auth-no-challenge --user=prometheus-exposed --ask-password \
  -O- "https://<your ces-fqdn>/prometheus/federate?match[]=%7B__name__%3D~%22.%2B%22%7D"
```

### Configure centralized Prometheus

You can then append a snippet like this to the `prometheus.yaml` of your centralized Prometheus to collect metrics from the CES-Prometheus:
```yaml
  - job_name: 'federate-ces'
    scrape_interval: 15s
    honor_labels: true
    metrics_path: '/prometheus/federate'
    params:
      'match[]':
        - '{__name__=~".+"}'
    static_configs:
      - targets: ['<your ces-fqdn>']
    scheme: https
    basic_auth:
      username: 'prometheus-exposed'
      password_file: '/path/to/your/password-file'
```