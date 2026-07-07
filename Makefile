# Detect OS and architecture for plugin directory
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
PLUGIN_ARCH := $(GOOS)_$(GOARCH)

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run unit tests
.PHONY: test
test:
	go test -v ./...

# Build the provider
.PHONY: build
build:
	go build -o terraform-provider-citrixspa

# Install the provider locally
.PHONY: install
install: build
	@echo "Installing provider for $(PLUGIN_ARCH) architecture..."
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)
	cp terraform-provider-citrixspa ~/.terraform.d/plugins/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)/
	@echo "Provider installed successfully for $(PLUGIN_ARCH)"

# Format code
.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Generate documentation
.PHONY: docs
docs:
	tfplugindocs

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-citrixspa
	rm -rf dist/

# Run all checks
.PHONY: check
check: fmt lint test

# Release build
.PHONY: release
release:
	goreleaser release --rm-dist

# Development install for local testing
.PHONY: dev-install
dev-install: build
	@echo "Installing provider for development ($(PLUGIN_ARCH))..."
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)
	cp terraform-provider-citrixspa ~/.terraform.d/plugins/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)/
	@echo "Development provider installed successfully for $(PLUGIN_ARCH)"

# Install provider for local filesystem testing
.PHONY: install-local
install-local: build
	@echo "Installing provider for local filesystem testing ($(PLUGIN_ARCH))..."
	$(eval TERRAFORM_PLUGIN_DIR := $(if $(TF_PLUGIN_DIR),$(TF_PLUGIN_DIR),$(HOME)/.terraform.d/plugins))
	mkdir -p $(TERRAFORM_PLUGIN_DIR)/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)
	cp terraform-provider-citrixspa $(TERRAFORM_PLUGIN_DIR)/registry.terraform.io/citrix/citrixspa/0.1.0/$(PLUGIN_ARCH)/
	@echo "Provider installed to $(TERRAFORM_PLUGIN_DIR) for $(PLUGIN_ARCH)"
	@echo "Generating .terraformrc file..."
	@./generate-terraformrc.sh
	@echo "Use TF_CLI_CONFIG_FILE=.terraformrc when running terraform commands"

# Set up local test environment
.PHONY: setup-local-test
setup-local-test: install-local
	@echo "Setting up local test environment..."
	cd test-local && cp terraform.tfvars.example terraform.tfvars
	@echo "Please edit test-local/terraform.tfvars with your actual credentials"
	@echo "Then run: make test-local"

# Run local tests
.PHONY: test-local
test-local: install-local
	@echo "Running local provider tests..."
	cd test-local && TF_CLI_CONFIG_FILE=../.terraformrc terraform init
	cd test-local && TF_CLI_CONFIG_FILE=../.terraformrc terraform validate
	cd test-local && TF_CLI_CONFIG_FILE=../.terraformrc terraform plan

# Run local tests with service principal authentication
.PHONY: test-local-sp
test-local-sp: install-local
	@echo "Running local provider tests with service principal authentication..."
	cd test-local-sp && TF_CLI_CONFIG_FILE=../.terraformrc terraform init -reconfigure
	cd test-local-sp && TF_CLI_CONFIG_FILE=../.terraformrc terraform validate
	cd test-local-sp && TF_CLI_CONFIG_FILE=../.terraformrc terraform plan

# Set up service principal test environment
.PHONY: setup-sp-test
setup-sp-test: install-local
	@echo "Setting up service principal test environment..."
	cd test-local-sp && cp terraform.tfvars.example terraform.tfvars
	@echo "Please edit test-local-sp/terraform.tfvars with your actual service principal credentials"
	@echo "Then run: make test-local-sp"

# Clean local test environment
.PHONY: clean-local
clean-local:
	@echo "Cleaning local test environment..."
	rm -rf test-local/.terraform*
	rm -f test-local/terraform.tfstate*
	rm -rf test-local-sp/.terraform*
	rm -f test-local-sp/terraform.tfstate*
	$(eval TERRAFORM_PLUGIN_DIR := $(if $(TF_PLUGIN_DIR),$(TF_PLUGIN_DIR),$(HOME)/.terraform.d/plugins))
	rm -rf $(TERRAFORM_PLUGIN_DIR)/registry.terraform.io/citrix/citrixspa

# Initialize Go modules
.PHONY: init
init:
	go mod init code.citrite.net/csvc/spa-sdk/terraform-provider-spa
	go mod tidy

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Run security scan
.PHONY: security
security:
	gosec ./...

# Run benchmarks
.PHONY: bench
bench:
	go test -bench=. ./...

# Coverage report
.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build the provider"
	@echo "  install         - Install the provider locally"
	@echo "  install-local   - Install provider for local filesystem testing"
	@echo "  setup-local-test- Set up local test environment"
	@echo "  test-local      - Run local provider tests"
	@echo "  setup-sp-test   - Set up service principal test environment"
	@echo "  test-local-sp   - Run local tests with service principal authentication"
	@echo "  clean-local     - Clean local test environment"
	@echo "  test            - Run unit tests"
	@echo "  testacc         - Run acceptance tests"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  docs            - Generate documentation"
	@echo "  clean           - Clean build artifacts"
	@echo "  check           - Run all checks (fmt, lint, test)"
	@echo "  release         - Release build"
	@echo "  dev-install     - Development install"
	@echo "  init            - Initialize Go modules"
	@echo "  deps            - Download dependencies"
	@echo "  security        - Run security scan"
	@echo "  bench           - Run benchmarks"
	@echo "  coverage        - Generate coverage report"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Environment Variables:"
	@echo "  TF_PLUGIN_DIR   - Override plugin directory (default: ~/.terraform.d/plugins)"
	@echo ""
	@echo "Detected platform: $(PLUGIN_ARCH)"
