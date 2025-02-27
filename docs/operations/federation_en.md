# Federation

Federation is a feature that enables you to expose metrics from your Prometheus instance to another Prometheus instance.

## Use-Case: Collecting metrics from CES-instances in a centralized Prometheus

In this use-case, we may have multiple CES-instances and want to collect their metrics in a centralized Prometheus.

You can do this for as many CES-instances as you like.

Files for testing this scenario can be found in the `samples` folder.

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

### Test federation

If you configured your local ecosystem (running at `192.168.56.2`) with the auth-presets from the `samples` folder,
you can then start a Prometheus in docker with federation enabled using the following command: 
```shell
docker run \
    -p 9090:9090 \
    -v $(pwd)/samples/prometheus.yaml:/etc/prometheus/prometheus.yml \
    prom/prometheus
```