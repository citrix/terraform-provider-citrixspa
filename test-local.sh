#!/bin/bash

# Local testing script for SPA Terraform Provider
# This script automates the full local development workflow:
#   1. Build the provider binary
#   2. Install it to the local plugin directory
#   3. Generate .terraformrc to redirect Terraform to the local binary
#   4. Run terraform init / validate / plan (and optionally apply)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if terraform.tfvars exists in the given test directory
check_credentials() {
    local test_dir="$1"
    if [ ! -f "${test_dir}/terraform.tfvars" ]; then
        print_error "terraform.tfvars not found in ${test_dir}/"
        print_warning "Please copy terraform.tfvars.example and fill in your credentials:"
        echo "  cp ${test_dir}/terraform.tfvars.example ${test_dir}/terraform.tfvars"
        exit 1
    fi
    print_status "Credentials file found: ${test_dir}/terraform.tfvars"
}

build_provider() {
    print_status "Building provider..."
    make build
    print_status "Provider built successfully"
}

install_provider() {
    print_status "Installing provider locally..."
    make install
    print_status "Provider installed successfully"
}

generate_terraformrc() {
    if [ ! -f ".terraformrc" ]; then
        print_status "Generating .terraformrc..."
        ./generate-terraformrc.sh
    else
        print_status ".terraformrc already exists, skipping generation"
    fi
}

run_terraform() {
    local test_dir="$1"
    local apply_flag="$2"

    print_status "Initializing Terraform in ${test_dir}/..."
    cd "${test_dir}"
    TF_CLI_CONFIG_FILE=../.terraformrc terraform init -reconfigure

    print_status "Validating configuration..."
    TF_CLI_CONFIG_FILE=../.terraformrc terraform validate

    print_status "Planning changes..."
    TF_CLI_CONFIG_FILE=../.terraformrc terraform plan -out=tfplan

    if [ "$apply_flag" = "yes" ]; then
        print_warning "Do you want to apply the changes? (y/N)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            print_status "Applying changes..."
            TF_CLI_CONFIG_FILE=../.terraformrc terraform apply tfplan
            print_status "Apply completed successfully"
        else
            print_status "Skipping apply"
        fi
    else
        print_status "Plan completed. Run with --apply to apply changes."
    fi

    cd ..
}

cleanup() {
    print_status "Cleaning up plan files..."
    rm -f test-local/tfplan
    rm -f test-local-sp/tfplan
}

full_cleanup() {
    print_status "Removing all artifacts created by this script..."
    rm -f terraform-provider-spa
    rm -f .terraformrc
    rm -rf test-local/.terraform test-local/.terraform.lock.hcl test-local/tfplan test-local/terraform.tfstate test-local/terraform.tfstate.backup
    rm -rf test-local-sp/.terraform test-local-sp/.terraform.lock.hcl test-local-sp/tfplan test-local-sp/terraform.tfstate test-local-sp/terraform.tfstate.backup
    rm -rf ~/.terraform.d/plugins/registry.terraform.io/citrix/spa/
    print_status "Cleanup complete. All test artifacts removed."
}

main() {
    print_status "Starting SPA Terraform Provider local testing..."

    auth_method="direct"
    apply_flag="no"

    while [[ $# -gt 0 ]]; do
        case $1 in
            --service-principal|--sp)
                auth_method="service-principal"
                shift
                ;;
            --apply)
                apply_flag="yes"
                shift
                ;;
            --cleanup)
                full_cleanup
                exit 0
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

    trap cleanup EXIT

    # Select test directory based on auth method
    if [ "$auth_method" = "service-principal" ]; then
        test_dir="test-local-sp"
        print_status "Using service principal authentication (${test_dir}/)"
    else
        test_dir="test-local"
        print_status "Using direct token authentication (${test_dir}/)"
    fi

    check_credentials "$test_dir"
    build_provider
    install_provider
    generate_terraformrc
    run_terraform "$test_dir" "$apply_flag"

    print_status "Done!"
}

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Local testing script for SPA Terraform Provider."
    echo "Builds, installs, and runs the provider against a local test configuration."
    echo ""
    echo "OPTIONS:"
    echo "  --service-principal, --sp   Use service principal authentication (test-local-sp/)"
    echo "  --apply                     Prompt to apply after planning (default: plan only)"
    echo "  --cleanup                   Remove all artifacts created by this script and exit"
    echo "  --help                      Show this help message"
    echo ""
    echo "SETUP:"
    echo "  Direct Token:"
    echo "    cp test-local/terraform.tfvars.example test-local/terraform.tfvars"
    echo "    # fill in citrix_customer_id and citrix_auth_token"
    echo ""
    echo "  Service Principal:"
    echo "    cp test-local-sp/terraform.tfvars.example test-local-sp/terraform.tfvars"
    echo "    # fill in citrix_customer_id, citrix_client_id, and citrix_client_secret"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Plan with direct token auth"
    echo "  $0 --sp                     # Plan with service principal auth"
    echo "  $0 --apply                  # Plan and prompt to apply"
    echo "  $0 --sp --apply             # Plan with SP auth and prompt to apply"
    echo "  $0 --cleanup                # Remove all test artifacts"
}

main "$@"
