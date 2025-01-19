PROJECT_NAME := Gatecam
BINARY_NAME := $(PROJECT_NAME)

PROJECT_ROOT := $(CURDIR)
BACKEND_SOURCE_DIR := $(PROJECT_ROOT)/app
FRONTEND_SOURCE_DIR := $(PROJECT_ROOT)/web/prod

BUILD_DIR := $(PROJECT_ROOT)/build
CURRENT_BUILD_DIR := $(BUILD_DIR)/Gatecam
DIST_DIR := $(CURRENT_BUILD_DIR)/dist

DEPLOY_HOST ?= razul-server.local
DEPLOY_USER ?= gatecam
DEPLOY_DIR ?= ~/
DEPLOY_SERVICE_NAME ?= gatecam.service

VERSION := $(shell grep '^VERSION=' $(PROJECT_ROOT)/.env | cut -d '=' -f 2)
ARCHS := amd64 arm
OS := linux

.PHONY: all
all: build

.PHONY: build
build: | $(CURRENT_BUILD_DIR)
	@echo "Building backend..."
	@CGO_ENABLED=1 GOOS=$(OS) GOARCH=amd64 go build -o $(CURRENT_BUILD_DIR)/$(BINARY_NAME) $(BACKEND_SOURCE_DIR)
	@cp -r $(BACKEND_SOURCE_DIR)/config.json $(CURRENT_BUILD_DIR)/config.json
	@cp -r $(BACKEND_SOURCE_DIR)/.env $(CURRENT_BUILD_DIR)/.env
	@cp -r $(BACKEND_SOURCE_DIR)/deploy.prototxt $(CURRENT_BUILD_DIR)/deploy.prototxt
	@cp -r $(BACKEND_SOURCE_DIR)/mobilenet_iter_73000.caffemodel $(CURRENT_BUILD_DIR)/mobilenet_iter_73000.caffemodel

.PHONY: package
package: all
	@echo "Packaging tarballs for version $(VERSION)..."
	@for arch in $(ARCHS); do \
		BUILD_ARCHIVE=$(BUILD_DIR)/$(shell echo $(PROJECT_NAME) | tr '[:upper:]' '[:lower:]')-$(VERSION)-$(OS)-static-$$arch.tar.gz; \
		echo "Building for $$arch..."; \
		TEMP_DIR=$(BUILD_DIR)/temp_$$arch; \
		mkdir -p $$TEMP_DIR/Gatecam; \
		GOOS=$(OS) GOARCH=$$arch CGO_ENABLED=1 go build -o $$TEMP_DIR/Gatecam/$(BINARY_NAME) $(BACKEND_SOURCE_DIR); \
		cp -r $(CURRENT_BUILD_DIR)/config.json $$TEMP_DIR/Gatecam/config.json; \
		echo "Creating archive $$BUILD_ARCHIVE..."; \
		tar -czvf $$BUILD_ARCHIVE -C $$TEMP_DIR Gatecam; \
		rm -rf $$TEMP_DIR; \
	done
	@echo "Packaging complete. Tarballs are in $(BUILD_DIR)"

.PHONY: docker-build
docker-build:
	@echo "Building Docker images with version $(VERSION)..."
	@VERSION=$(VERSION) $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build

.PHONY: docker-up
docker-up:
	@echo "Starting Docker containers..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

.PHONY: docker-clean
docker-clean:
	@echo "Removing Docker containers, images, and volumes..."
	@$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down --rmi all --volumes --remove-orphans

.PHONY: clean
clean:
	@echo "Cleaning up build files..."
	@rm -rf $(BUILD_DIR)

.PHONY: clean-tarballs
clean-tarballs:
	@echo "Cleaning up tarballs..."
	@find $(BUILD_DIR) -type f -name '*.tar.gz' -delete

.PHONY: deploy
deploy: all
	@echo "Deploying to $(DEPLOY_HOST)..."
	@rsync -avz --	 $(BUILD_DIR)/$(PROJECT_NAME) $(DEPLOY_USER)@$(DEPLOY_HOST):$(DEPLOY_DIR)
	@ssh -t $(DEPLOY_USER)@$(DEPLOY_HOST) 'sudo systemctl restart $(DEPLOY_SERVICE_NAME) && sudo systemctl reload nginx'
	@echo "Deployment complete."

.PHONY: help
help:
	@echo "Usage:"
	@echo "  make                      Build the Go backend and Vue.js frontend and install NPM dependencies"
	@echo "  make build                Build the Go backend"
	@echo "  make package              Build and package the application for all architectures"
	@echo "  make docker-build         Build Docker images"
	@echo "  make docker-up            Start Docker containers"
	@echo "  make docker-down          Stop Docker containers"
	@echo "  make docker-clean         Clean up Docker containers, images, and volumes"
	@echo "  make deploy               Deploy the application to the remote server"
	@echo "  make clean                Clean up build files"
	@echo "  make clean-tarballs       Clean up tarball archives"
	@echo "  make help                 Show this help message"

$(CURRENT_BUILD_DIR):
	@mkdir -p $(CURRENT_BUILD_DIR)
