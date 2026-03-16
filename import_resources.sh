#!/bin/bash

# Auto-generated import script
# Run this script to import all discovered resources

set -e

print_status() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

print_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

print_debug() {
    echo -e "\033[0;34m[DEBUG]\033[0m $1"
}

print_status "Importing browser mode configuration"
print_debug "Running: terraform import spa_browser_mode.browser_mode browser_mode"
terraform import spa_browser_mode.browser_mode browser_mode || print_error "Failed to import browser mode"

print_status "Import script completed"
print_status "Run 'terraform plan' to verify the imported resources"
