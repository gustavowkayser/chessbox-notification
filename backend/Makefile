.PHONY: all build build-arm64 build-amd64 build-local run test clean tf-init tf-plan tf-apply

APP_NAME ?= chessbox-notification
BUILD_DIR ?= build
BINARY_NAME ?= bootstrap
LAMBDA_ZIP ?= $(BUILD_DIR)/bootstrap.zip
TERRAFORM_DIR ?= terraform

all: build

# Default Lambda build (ARM64 / Graviton - lower cost & higher performance)
build: build-arm64

build-arm64:
	@echo "Building Lambda binary for linux/arm64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Packaging $(LAMBDA_ZIP)..."
	@cd $(BUILD_DIR) && zip -j $(notdir $(LAMBDA_ZIP)) $(BINARY_NAME)
	@echo "Build complete: $(LAMBDA_ZIP)"

build-amd64:
	@echo "Building Lambda binary for linux/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Packaging $(LAMBDA_ZIP)..."
	@cd $(BUILD_DIR) && zip -j $(notdir $(LAMBDA_ZIP)) $(BINARY_NAME)
	@echo "Build complete: $(LAMBDA_ZIP)"

build-local:
	@echo "Building local binary..."
	@mkdir -p bin
	go build -o bin/$(APP_NAME) cmd/main.go

run:
	@echo "Running local server..."
	go run cmd/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning artifacts..."
	rm -rf bin $(BUILD_DIR) terraform/.terraform terraform/.terraform.lock.hcl terraform/terraform.tfstate*

# Terraform helper targets
tf-init:
	cd $(TERRAFORM_DIR) && terraform init

tf-plan: build
	cd $(TERRAFORM_DIR) && terraform plan

tf-apply: build
	cd $(TERRAFORM_DIR) && terraform apply
