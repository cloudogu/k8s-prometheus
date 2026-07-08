# PoC: Test the ServiceAccountProducer CR in the cluster

This guide tests the Prometheus service-account API directly from a debug pod inside the cluster.
It reads the endpoint, secret name, and secret key from the `ServiceAccountProducer` CR so the test uses exactly the addressing configured in the cluster.

## Prerequisites
```bash
kubectl get crd serviceaccountproducers.k8s.cloudogu.com
kubectl -n ecosystem get serviceaccountproducer k8s-prometheus -o yaml
```

## Read the data from the producer CR
```bash
NAMESPACE="ecosystem"
PRODUCER_NAME="k8s-prometheus"

ENDPOINT="$(kubectl -n "${NAMESPACE}" get serviceaccountproducer "${PRODUCER_NAME}" -o jsonpath='{.spec.http.endpoint}')"
AUTH_SECRET_NAME="$(kubectl -n "${NAMESPACE}" get serviceaccountproducer "${PRODUCER_NAME}" -o jsonpath='{.spec.http.authSecret.name}')"
AUTH_SECRET_KEY="$(kubectl -n "${NAMESPACE}" get serviceaccountproducer "${PRODUCER_NAME}" -o jsonpath='{.spec.http.authSecret.key}')"
API_KEY="$(kubectl -n "${NAMESPACE}" get secret "${AUTH_SECRET_NAME}" -o "jsonpath={.data.${AUTH_SECRET_KEY}}" | base64 -d)"
```

Optionally, print the resolved endpoint:
```bash
echo "${ENDPOINT}"
```

## Add network policy to allow access to the producer endpoint
```bash
kubectl -n "${NAMESPACE}" apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: sa-producer-debug-netpol
spec:
  podSelector:
    matchLabels:
      app: k8s-prometheus
  ingress:
  - from:
    - podSelector:
        matchLabels:
          run: sa-producer-debug
EOF
```

## Start a debug pod
The debug pod runs in the same namespace and receives the endpoint and API key as environment variables.

```bash
kubectl -n "${NAMESPACE}" run sa-producer-debug \
  --image=curlimages/curl:8.12.1 \
  --restart=Never \
  --env="ENDPOINT=${ENDPOINT}" \
  --env="API_KEY=${API_KEY}" \
  --command -- sleep 3600
```

Wait until the pod is ready:
```bash
kubectl -n "${NAMESPACE}" wait --for=condition=Ready pod/sa-producer-debug --timeout=120s
```

## Create a service account via HTTP request
The producer endpoint already contains the `/serviceaccounts` path.
For the create request, call `${ENDPOINT}/`.

```bash
kubectl -n "${NAMESPACE}" exec sa-producer-debug -- \
  curl -i \
    -X POST "${ENDPOINT}/" \
    -H "Content-Type: application/json" \
    -H "X-CES-SA-API-KEY: ${API_KEY}" \
    -d '{"consumer":"manual-test","params":[]}'
```

Expected result:
- HTTP `201 Created`
- JSON containing `username` and `password`

## Optional: print generated credentials again
If the response was not stored, the request can be repeated with another consumer name:

```bash
kubectl -n "${NAMESPACE}" exec sa-producer-debug -- \
  curl -s \
    -X POST "${ENDPOINT}/" \
    -H "Content-Type: application/json" \
    -H "X-CES-SA-API-KEY: ${API_KEY}" \
    -d '{"consumer":"manual-test-2","params":[]}'
```

## Delete the service account again
```bash
kubectl -n "${NAMESPACE}" exec sa-producer-debug -- \
  curl -i \
    -X DELETE "${ENDPOINT}/manual-test" \
    -H "X-CES-SA-API-KEY: ${API_KEY}"
```

Expected result:
- HTTP `204 No Content`

## Cleanup
```bash
kubectl -n "${NAMESPACE}" delete pod sa-producer-debug
kubectl -n "${NAMESPACE}" delete networkpolicy sa-producer-debug-netpol
```
