include .env
export

export PROJECT_ROOT=$(shell pwd)

env-up:
	@docker compose up -d todoapp-postgres

env-down:
	@docker compose down todoapp-postgres

env-cleanup:
	@read -p "Удалить все volume файлы? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down; \
		docker volume rm todo-api_pgdata; \
	else \
		echo "Очистка отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm migrate-postgres \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up 1"; \
		exit 1; \
	fi; \
	docker compose run --rm migrate-postgres \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todoapp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

todoapp-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=127.0.0.1 && \
	export TASK_EVENTS_BROKERS=localhost:9092 && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/todoapp/main.go

redis-up:
	@docker compose up -d redis

kafka-up:
	@docker compose up -d kafka

kafka-down:
	@docker compose stop kafka

task-audit-consumer:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=127.0.0.1 && \
	export TASK_EVENTS_BROKERS=localhost:9092 && \
	go run ${PROJECT_ROOT}/cmd/task-audit-consumer/main.go

todoapp-deploy:
	@docker compose up -d --build todoapp task-audit-consumer

todoapp-undeploy:
	@docker compose down todoapp

ps:
	@docker compose ps

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/todoapp/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

k8s-cluster:
    @if ! kind get clusters | grep -Fxq "$(KIND_CLUSTER)"; then \
        kind create cluster --name "$(KIND_CLUSTER)"; \
    fi

k8s-config: k8s-cluster
    @kubectl apply -f k8s/namespace.yaml
    @kubectl create secret generic todo-secrets -n "$(K8S_NAMESPACE)" \
        --from-literal=POSTGRES_USER="$(POSTGRES_USER)" \
        --from-literal=POSTGRES_PASSWORD="$(POSTGRES_PASSWORD)" \
        --from-literal=POSTGRES_DB="$(POSTGRES_DB)" \
        --dry-run=client -o yaml | kubectl apply -f -
    @kubectl create configmap todo-migrations -n "$(K8S_NAMESPACE)" \
        --from-file=migrations/ --dry-run=client -o yaml | kubectl apply -f -

k8s-images: k8s-cluster
    @docker pull postgres:17-bookworm
    @docker pull redis:7-alpine
    @docker pull confluentinc/cp-kafka:7.8.0
    @docker pull migrate/migrate:v4.19.1
    @kind load docker-image postgres:17-bookworm redis:7-alpine \
        confluentinc/cp-kafka:7.8.0 migrate/migrate:v4.19.1 \
        --name "$(KIND_CLUSTER)"

k8s-app-image: k8s-cluster
    @docker build -t "$(K8S_IMAGE)" -f cmd/todoapp/Dockerfile .
    @kind load docker-image "$(K8S_IMAGE)" --name "$(KIND_CLUSTER)"

k8s-up: k8s-config k8s-images k8s-app-image
    @kubectl apply -k k8s/base
    @kubectl wait --for=condition=ready pod -l app=postgres \
        -n "$(K8S_NAMESPACE)" --timeout=180s
    @kubectl wait --for=condition=ready pod -l app=kafka \
        -n "$(K8S_NAMESPACE)" --timeout=600s
    @kubectl wait --for=condition=ready pod -l app=redis \
        -n "$(K8S_NAMESPACE)" --timeout=180s
    @kubectl delete job task-events-migration -n "$(K8S_NAMESPACE)" \
        --ignore-not-found=true
    @kubectl apply -k k8s/migrations
    @kubectl wait --for=condition=complete job/task-events-migration \
        -n "$(K8S_NAMESPACE)" --timeout=180s
    @kubectl apply -k k8s/app
    @kubectl rollout status deployment/todo-api \
        -n "$(K8S_NAMESPACE)" --timeout=180s
    @kubectl rollout status deployment/task-audit-consumer \
        -n "$(K8S_NAMESPACE)" --timeout=180s

k8s-down:
    @kubectl delete -k k8s/app -n "$(K8S_NAMESPACE)" --ignore-not-found=true
    @kubectl delete -k k8s/migrations -n "$(K8S_NAMESPACE)" --ignore-not-found=true
    @kubectl delete -k k8s/base -n "$(K8S_NAMESPACE)" --ignore-not-found=true

k8s-clean:
    @read -r -p "Удалить namespace и данные PVC? [y/N] " answer; \
    if [ "$$answer" = "y" ]; then \
        kubectl delete namespace "$(K8S_NAMESPACE)"; \
    else \
        echo "Namespace и данные оставлены"; \
    fi

k8s-status:
    @kubectl get pods,svc,pvc -n "$(K8S_NAMESPACE)"

k8s-port-forward:
    @kubectl port-forward -n "$(K8S_NAMESPACE)" service/todo-api 5050:5050
	