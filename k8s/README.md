## Build and load the application image

Create a kind cluster if needed, then build and load the same image used by the API and audit consumer:

```sh
kind create cluster --name todo-local

docker build -t todo-api:dev -f cmd/todoapp/Dockerfile .
kind load docker-image todo-api:dev --name todo-local
```

If the cluster already exists, skip `kind create cluster`.

## Create local credentials and migration ConfigMap

The credentials are intentionally not stored in these manifests. Replace the sample values with local development credentials:

```sh
kubectl apply -f k8s/namespace.yaml
kubectl create secret generic todo-secrets -n todo-api \
  --from-literal=POSTGRES_USER=todo \
  --from-literal=POSTGRES_PASSWORD=todo-local \
  --from-literal=POSTGRES_DB=todo \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl create configmap todo-migrations -n todo-api \
  --from-file=migrations/ \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Deploy

Apply infrastructure, wait for the stateful dependencies, run migrations, then start the API and consumer:

```sh
kubectl apply -k k8s/base
kubectl wait --for=condition=ready pod -l app=postgres -n todo-api --timeout=180s
kubectl wait --for=condition=ready pod -l app=kafka -n todo-api --timeout=240s
kubectl apply -k k8s/migrations
kubectl wait --for=condition=complete job/task-events-migration -n todo-api --timeout=180s
kubectl apply -k k8s/app
```

To rerun the migration Job after a later migration, delete and recreate it:

```sh
kubectl delete job task-events-migration -n todo-api
kubectl apply -k k8s/migrations
kubectl wait --for=condition=complete job/task-events-migration -n todo-api --timeout=180s
```

## Verify

```sh
kubectl get pods,svc -n todo-api
kubectl logs -n todo-api deployment/todo-api
kubectl logs -n todo-api deployment/task-audit-consumer
kubectl port-forward -n todo-api service/todo-api 5050:5050
```

Then use `http://localhost:5050/swagger/index.html`. After creating, updating, or deleting a task, inspect `todoapp.task_events` in PostgreSQL. Kafka events are delivered at least once; the audit table deduplicates by `event_id`.
