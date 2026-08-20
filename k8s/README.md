# Kubernetes

These manifests run the core XKCD Search stack in the Minikube profile `xkcd`.
PostgreSQL and API secrets are created separately and are not tracked by Git.

## Build local images

```bash
docker build -t words:local -f search-services/Dockerfile.words search-services
docker build -t update:local -f search-services/Dockerfile.update search-services
docker build -t search:local -f search-services/Dockerfile.search search-services
docker build -t api:local -f search-services/Dockerfile.api search-services
docker build -t frontend:local -f search-services/Dockerfile.frontend search-services
```

Load them into Minikube:

```bash
minikube image load words:local --profile xkcd
minikube image load update:local --profile xkcd
minikube image load search:local --profile xkcd
minikube image load api:local --profile xkcd
minikube image load frontend:local --profile xkcd
```

## Create secrets

Create the namespace first:

```bash
kubectl apply -f k8s/base/namespace.yaml
```

Create `k8s/base/postgres-secret.yaml` from `postgres-secret.example.yaml`, replace
the placeholder consistently, and apply it:

```bash
kubectl apply -f k8s/base/postgres-secret.yaml
```

Load the API credentials from the ignored `.env` file:

```bash
set -a
source .env
set +a
kubectl create secret generic api-secret \
  --namespace xkcd \
  --from-literal=ADMIN_USER="$ADMIN_USER" \
  --from-literal=ADMIN_PASSWORD="$ADMIN_PASSWORD" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --dry-run=client \
  --output yaml | kubectl apply -f -
```

## Apply and verify

```bash
kubectl apply -k k8s/base
kubectl get pods,services,pvc,hpa -n xkcd
kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/xkcd/pods
```

Open the frontend:

```bash
minikube service frontend --namespace xkcd --profile xkcd
```

Check the API directly:

```bash
kubectl port-forward service/api 28080:8080 --namespace xkcd
curl http://localhost:28080/api/ping
```
