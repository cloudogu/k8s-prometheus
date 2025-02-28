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

Stellen Sie sicher, dass das Passwort mit bcrypt gehasht wird, z.B. mit folgendem Befehl aus dem `apache2-utils`-Paket:
```shell
  htpasswd -bnBC 10 "" <your-password> | tr -d ':\n' | sed 's/$2y/$2a/'
```

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

### Föderation testen

Wenn Sie Ihr lokales EcoSystem (unter `192.168.56.2`) mit den auth-presets aus dem Ordner `samples` konfiguriert haben,
können Sie dann einen Prometheus in Docker mit aktivierter Föderation mit dem folgenden Befehl starten:
```shell
docker run \
    -p 9090:9090 \
    -v $(pwd)/samples/prometheus.yaml:/etc/prometheus/prometheus.yml \
    prom/prometheus
```

Sie können dann [`scrape_samples_scraped{job="federate-ces"}` abfragen](http://localhost:9090/query?g0.expr=scrape_samples_scraped%7Bjob%3D%22federate-ces%22%7D) und das Ergebnis rechts sollte nicht 0 sein.