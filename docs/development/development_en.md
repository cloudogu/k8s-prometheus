# Development of the Auth sidecar container

The `k8s-prometheus` component has a sidecar container that performs the following tasks:
* Managing service accounts: Providing an API used by the `k8s-dogu-operator`.
* Authentication of access to the Prometheus API using the service accounts

The two services run on ports `8086` (Auth-Proxy) and `8087` (Service-Account API) in the same container, each as a separate [Gin](https://github.com/gin-gonic/gin) service.

### Administration of service accounts
In order for the Dogu operator to be able to use the service account API, it must be defined how this can be accessed.
To do this, the Dogu operator determines the associated Kubernetes service using the label `ces.cloudogu.com/serviceaccount-provider`.
This service is defined in [sa-provider-svc.yaml](../../k8s/helm/templates/sa-provider-svc.yaml).

The other required information is specified in the annotations on the service:

| annotation                                    | description                                                                             | value                                |
|-----------------------------------------------|-----------------------------------------------------------------------------------------|--------------------------------------|
| `ces.cloudogu.com/serviceaccount/port`        | Specifies the port under which the ServiceAccount API is available                      | `8080`                               |
| `ces.cloudogu.com/serviceaccount/path`        | Specifies the path under which the ServiceAccount API is available                      | `/serviceaccounts`                   |
| `ces.cloudogu.com/serviceaccount/secret-name` | Specifies the name of the secret that contains the API key for authentication           | `k8s-prometheus-service-account-api` |
| `ces.cloudogu.com/serviceaccount/secret-key`  | Specifies the key in the secret under which the API key for authentication can be found | `apiKey`                             |

With the `secret-name` and the `secret-key`, the Dogu operator can read the ApiKey from the specified secret of the component.
With this information and the IP address of the service, the Dogu operator can make requests to the ServiceAccount API of the respective component.

The secret for the ApiKey is created with the Helm template [sa-api-secret](../../k8s/helm/templates/sa-api-secret.yaml).

The created service accounts are saved in a YAML file in the [Config-Volume](../../k8s/helm/templates/config-pvc.yaml).

### Auth-Proxy
To secure access to the Prometheus API, incoming network traffic is blocked via a [NetworkPolicy](../../k8s/helm/templates/network-policy.yaml).
Only incoming communication to ports `8086` and `8087` is permitted.

The auth proxy in the sidecar container runs on port `8086` and acts as a reverse proxy for the Prometheus container.
Each request is checked via Basic-Auth.
The Basic-Auth header must contain the data of a previously created service account.

### Configuration of the sidecar container
The sidecar container is configured in the [values.yaml](../../k8s/helm/values.yaml) under the path `kube-prometheus-stack.prometheus.promtheusSpec.containers`.

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
```

The following environment variables can be configured:

| name              | description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| LOG_LEVEL         | The log level to be used (`DEBUG`, `INFO`, `WARN`, `ERROR`). Default `INFO` |
| PROMETHEUS_URL    | The URL for the auth proxy under which Prometheus can be reached            |
| WEB_PRESETS_FILE  | Read-only file for preset Service-Accounts, e.g. from a secret              |
| WEB_CONFIG_FILE   | The file in which the service accounts are saved                            |
| API_KEY           | The API key for the service account API                                     |


## Development in the local CES cluster
To test the `k8s-prometheus` component in the local CES cluster, the following make targets can be used:

Initial, so that the component can be installed:
```bash
make helm-update-dependencies
```

```bash
# Installs k8s-prometheus in the cluster
make component-apply
```

```bash
# Deletes k8s-prometheus from the cluster
make component-delete
```

```bash
# Deletes k8s-prometheus from the cluster and then reinstalls it
make component-reinstall
```

```bash
# Installs the Helm chart of k8s-prometheus in the cluster (without the CES component)
make helm-apply
```

```bash
# Deletes the helmet chart of k8s-prometheus from the cluster (without the CES component)
make helm-delete
```