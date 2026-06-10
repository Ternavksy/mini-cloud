.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make minio-up      - Start MinIO in Docker"
	@echo "  make minio-down    - Stop MinIO"
	@echo "  make deps          - Download Go dependencies"
	@echo "  make build         - Build the application"
	@echo "  make generate      - Generate API and test mock code"
	@echo "  make run           - Run the application"
	@echo "  make dev           - Run with S3 storage"
	@echo "  make local         - Run with local storage"
	@echo "  make test-upload   - Test upload endpoint"
	@echo "  make test-list     - Test list endpoint"

.PHONY: minio-up
minio-up:
	docker-compose up -d
	@echo "MinIO запущен:"
	@echo "  API:     http://localhost:9000"
	@echo "  Console: http://localhost:9001"
	@echo "  Login:   minioadmin / minioadmin"

.PHONY: minio-down
minio-down:
	docker-compose down

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: build
build:
	go build -o mini-cloud main.go

.PHONY: generate
generate:
	go generate ./internal/api ./internal/server

.PHONY: run
run: build
	./mini-cloud server

.PHONY: dev
dev:
	STORAGE_TYPE=s3 go run main.go server

.PHONY: local
local:
	STORAGE_TYPE=local go run main.go server

.PHONY: test-upload
test-upload:
	@echo "Creating test file..."
	@echo "Hello, World!" > /tmp/test.txt
	@echo "Uploading..."
	curl -X POST -F "file=@/tmp/test.txt" http://localhost:8080/upload

.PHONY: test-list
test-list:
	curl http://localhost:8080/files

.PHONY: clean
clean:
	rm -f mini-cloud
	rm -rf storage/

.PHONY: clean-minio
clean-minio:
	docker-compose down -v

.PHONY: full-clean
full-clean: clean clean-minio
	@echo "Full cleanup done"
