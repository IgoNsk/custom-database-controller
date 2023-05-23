.PHONY: build
build: init ## Build application
	mkdir -p build
	GO111MODULE=on GOOS=linux go build -o build/custom_database_controller cmd/customdatabase-controller/main.go

.PHONY: lint
lint: ## Run linter on sources todo
	#GO111MODULE=on golangci-lint run --exclude-use-default=false --timeout 10m ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: gen
gen:
	tools/update-codegen.sh

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: register-crd
register-crd:
	minikube kubectl -- create -f artifacts/crd.yaml

.PHONY: unregister-crd
unregister-crd:
	minikube kubectl -- delete crd customdatabases.igor.yatsevich.ru || exit 1

.PHONY: init
init: clean vendor gen unregister-crd register-crd

.PHONY: clean
clean:
	rm -rf build

.PHONY: dev-run
dev-run: build
	build/custom_database_controller \
		-kubeconfig=/home/igo/.kube/config \
		-pg_host=localhost \
		-pg_port=5432 \
		-pg_admin_user=custom_database_admin \
		-pg_admin_password=admin_password

.PHONY: dev-integration-test
dev-integration-test:
	echo "cleanup"
	-@minikube kubectl -- delete CustomDatabase example-database
	-@minikube kubectl -- delete CustomDatabase another-database
	echo "create CustomDatabase..."
	minikube kubectl -- create -f artifacts/customdatabase-example.yaml
	minikube kubectl -- create -f artifacts/customdatabase-another-example.yaml
	echo "update CustomDatabase..."
	sleep 5s
	minikube kubectl -- apply -f artifacts/customdatabase-another-example-updated.yaml
	echo "delete CustomDatabase..."
	sleep 5s
	minikube kubectl -- delete CustomDatabase another-database
