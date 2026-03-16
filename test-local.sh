#!/bin/bash

# Local testing script for SPA Terraform Provider
# This script automates the local testing workflow

set -e

# Default plugin directory
DEFAULT_PLUGIN_DIR="$HOME/.terraform/plugins"

# Get plugin directory from environment or command line
get_plugin_dir() {
    if [ -n "$TF_PLUGIN_DIR" ]; then
        echo "$TF_PLUGIN_DIR"
    else
        echo "$DEFAULT_PLUGIN_DIR"
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if terraform.tfvars exists
check_credentials() {
    if [ ! -f "test-local/terraform.tfvars" ]; then
        print_error "terraform.tfvars not found!"
        print_warning "Please copy terraform.tfvars.example to terraform.tfvars and fill in your credentials"
        echo "  cp test-local/terraform.tfvars.example test-local/terraform.tfvars"
        echo "  vim test-local/terraform.tfvars"
        exit 1
    fi
    print_status "Credentials file found"
}

# Check if service principal credentials exist
check_sp_credentials() {
    if [ ! -f "test-local/service-principal.tfvars" ]; then
        print_error "service-principal.tfvars not found!"
        print_warning "Please copy service-principal.tfvars.example to service-principal.tfvars and fill in your credentials"
        echo "  cp test-local/service-principal.tfvars.example test-local/service-principal.tfvars"
        echo "  vim test-local/service-principal.tfvars"
        exit 1
    fi
    print_status "Service principal credentials file found"
}

# Build the provider
build_provider() {
    print_status "Building provider..."
    make build
    if [ $? -eq 0 ]; then
        print_status "Provider built successfully"
    else
        print_error "Failed to build provider"
        exit 1
    fi
}

# Install provider locally
install_provider() {
    print_status "Installing provider locally..."
    make install-local
    if [ $? -eq 0 ]; then
        print_status "Provider installed successfully"
    else
        print_error "Failed to install provider"
        exit 1
    fi
}

# Initialize terraform
init_terraform() {
    print_status "Initializing Terraform..."
    cd test-local
    TF_CLI_CONFIG_FILE=../.terraformrc terraform init -reconfigure
    cd ..
    if [ $? -eq 0 ]; then
        print_status "Terraform initialized successfully"
    else
        print_error "Failed to initialize Terraform"
        exit 1
    fi
}

# Validate configuration
validate_config() {
    print_status "Validating Terraform configuration..."
    cd test-local
    TF_CLI_CONFIG_FILE=../.terraformrc terraform validate
    cd ..
    if [ $? -eq 0 ]; then
        print_status "Configuration is valid"
    else
        print_error "Configuration validation failed"
        exit 1
    fi
}

# Plan changes
plan_changes() {
    print_status "Planning Terraform changes..."
    cd test-local
    TF_CLI_CONFIG_FILE=../.terraformrc terraform plan -out=tfplan
    cd ..
    if [ $? -eq 0 ]; then
        print_status "Plan completed successfully"
    else
        print_error "Plan failed"
        exit 1
    fi
}

# Plan changes with service principal
plan_sp_changes() {
    print_status "Planning Terraform changes with service principal authentication..."
    cd test-local
    TF_CLI_CONFIG_FILE=../.terraformrc terraform plan -var-file=service-principal.tfvars service-principal.tf -out=tfplan
    cd ..
    if [ $? -eq 0 ]; then
        print_status "Service principal plan completed successfully"
    else
        print_error "Service principal plan failed"
        exit 1
    fi
}

# Apply changes (optional)
apply_changes() {
    print_warning "Do you want to apply the changes? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Applying Terraform changes..."
        cd test-local
        TF_CLI_CONFIG_FILE=../.terraformrc terraform apply tfplan
        cd ..
        if [ $? -eq 0 ]; then
            print_status "Apply completed successfully"
        else
            print_error "Apply failed"
            exit 1
        fi
    else
        print_status "Skipping apply"
    fi
}

# Apply changes with service principal (optional)
apply_sp_changes() {
    print_warning "Do you want to apply the service principal changes? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Applying Terraform changes with service principal..."
        cd test-local
        TF_CLI_CONFIG_FILE=../.terraformrc terraform apply tfplan
        cd ..
        if [ $? -eq 0 ]; then
            print_status "Service principal apply completed successfully"
        else
            print_error "Service principal apply failed"
            exit 1
        fi
    else
        print_status "Skipping service principal apply"
    fi
}

# Cleanup function
cleanup() {
    print_status "Cleaning up..."
    rm -f test-local/tfplan
    rm -f test-local/terraform.tfstate.backup
}

# Main execution
main() {
    print_status "Starting SPA Terraform Provider local testing..."
    
    # Parse command line arguments
    auth_method="direct"
    apply_changes_flag=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --service-principal|--sp)
                auth_method="service-principal"
                shift
                ;;
            --apply)
                apply_changes_flag="--apply"
                shift
                ;;
            --plugin-dir)
                export TF_PLUGIN_DIR="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Display current plugin directory
    current_plugin_dir=$(get_plugin_dir)
    print_status "Using plugin directory: $current_plugin_dir"
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    # Check prerequisites based on authentication method
    if [ "$auth_method" = "service-principal" ]; then
        check_sp_credentials
    else
        check_credentials
    fi
    
    # Build and test workflow
    build_provider
    install_provider
    init_terraform
    
    if [ "$auth_method" = "service-principal" ]; then
        plan_sp_changes
        
        # Optionally apply changes
        if [ "$apply_changes_flag" = "--apply" ]; then
            apply_sp_changes
        else
            print_status "Service principal test completed successfully!"
            print_warning "Run with --apply flag to actually apply changes"
        fi
    else
        validate_config
        plan_changes
        
        # Optionally apply changes
        if [ "$1" = "--apply" ]; then
            apply_changes
        else
            print_status "Test completed successfully!"
            print_warning "Run with --apply flag to actually apply changes"
        fi
    fi
}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Local testing script for SPA Terraform Provider"
    echo ""
    echo "OPTIONS:"
    echo "  --service-principal, --sp   Use service principal authentication"
    echo "  --apply                     Apply the planned changes (default: plan only)"
    echo "  --plugin-dir <path>         Specify custom plugin directory"
    echo "  --help                      Show this help message"
    echo ""
    echo "ENVIRONMENT VARIABLES:"
    echo "  TF_PLUGIN_DIR              Override plugin directory (default: ~/.terraform/plugins)"
    echo ""
    echo "AUTHENTICATION METHODS:"
    echo "  1. Direct Token (default):"
    echo "     - Copy test-local/terraform.tfvars.example to test-local/terraform.tfvars"
    echo "     - Fill in your Citrix Cloud customer ID and auth token"
    echo ""
    echo "  2. Service Principal (recommended):"
    echo "     - Copy test-local/service-principal.tfvars.example to test-local/service-principal.tfvars"
    echo "     - Fill in your Citrix Cloud customer ID, client ID, and client secret"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Test with direct token authentication"
    echo "  $0 --service-principal      # Test with service principal authentication"
    echo "  $0 --sp --apply            # Test with service principal and apply changes"
    echo "  $0 --apply                 # Test with direct token and apply changes"
    echo "  $0 --plugin-dir /tmp/plugins # Use custom plugin directory"
    echo "  TF_PLUGIN_DIR=/custom/path $0  # Use custom plugin directory via environment"
}

# Execute main function with all arguments
main "$@"
