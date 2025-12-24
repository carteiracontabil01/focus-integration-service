run:
	go run ./cmd/api

swag:
	swag init --dir cmd/api,internal --output docs

build:
	go build -o bin/focus-company-integration-service ./cmd/api


