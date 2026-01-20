.SILENT:

.DEFAULT_GOAL := run

get:
	go get

wire:
	go run github.com/google/wire/cmd/wire

build: wire
	go build -o bin .

run:
	go run .

test_services:
	cd pkg/services && go test -v -cover

# Docker commands
# Variables for Docker
DOCKER_IMAGE_NAME ?= berliner-backend
DOCKER_TAG ?= latest
DOCKER_REGISTRY ?= docker.io
DOCKER_USERNAME ?= asyli1

docker_build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

docker_run:
	docker run -p 8080:8080 --env-file configs/.env $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker_tag:
	docker tag $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker_push: docker_tag
	docker push $(DOCKER_REGISTRY)/$(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker_login:
	docker login $(DOCKER_REGISTRY)

# Build, tag and push in one command
docker_release: docker_build docker_push
	@echo "Docker image released: $(DOCKER_REGISTRY)/$(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"
