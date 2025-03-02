# Federation

Federation ist eine Funktion, mit der Metriken aus einer Prometheus-Instanz für eine andere Prometheus-Instanz verfügbar gemacht werden können.

## Anwendungsfall: Sammeln von Metriken aus CES-Instanzen in einem zentralen Prometheus

In diesem Anwendungsfall haben wir möglicherweise mehrere CES-Instanzen und möchten deren Metriken in einem zentralen Prometheus sammeln.

Sie können dies für beliebig viele CES-Instanzen tun.

### Konfiguration vom CES-Prometheus

Zuerst muss der CES-Prometheus außerhalb des Clusters vervügbar gemacht werden.
Folgende `valuesYamlOverwrite` erstellt über einen Ingress eine Route dorthin:
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

Da sich CES-Prometheus hinter einem Authentifizierungsproxy befindet, muss ein Service-Account erstellt werden, indem ein Secret wie das Folgende angewendet wird:
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

Das Passwort muss mit bcrypt gehasht werden, z.B. mit folgendem Befehl aus dem `apache2-utils`-Paket:
```shell
  htpasswd -bnBC 10 "" <your-password> | tr -d ':\n' | sed 's/$2y/$2a/'
```

Prometheus muss anschließend neugestartet werden, damit der Auth-Proxy die Datei lädt.

Die Authentifizierung und die Föderation kann mit diesem Befehl getestet werden:
```shell
wget --header="Content-Type: application/json" --header="Accept: application/json" \
  --auth-no-challenge --user=prometheus-exposed --ask-password \
  -O- "https://<your ces-fqdn>/prometheus/federate?match[]=%7B__name__%3D~%22.%2B%22%7D"
```

### Konfiguration des zentralisierten Prometheus

Analog zum folgenden Abschnitt die `prometheus.yaml` des zentralisierten Prometheus erweitern, um Metriken aus dem CES-Prometheus zu sammeln:
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

Wenn ein lokales EcoSystem (unter `192.168.56.2`) mit den auth-presets aus dem Ordner `samples` konfiguriert wird,
kann anschließend ein Prometheus in Docker mit aktivierter Föderation mit dem folgenden Befehl gestartet werden:
```shell
docker run \
    -p 9090:9090 \
    -v $(pwd)/samples/prometheus.yaml:/etc/prometheus/prometheus.yml \
    prom/prometheus
```

Nun kann z.B. [scrape_samples_scraped{job="federate-ces"}](http://localhost:9090/query?g0.expr=scrape_samples_scraped%7Bjob%3D%22federate-ces%22%7D) abgefragt werden und das Ergebnis rechts sollte nicht 0 sein.