#!/bin/bash

# import_automation.sh
# Automation script for discovering and importing SPA resources

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TERRAFORM_DIR="."
STATE_FILE="terraform.tfstate"
BACKUP_DIR="./import_backups"
DEBUG_LOGGING=false

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_debug() {
    if [ "$DEBUG_LOGGING" = true ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Function to create backup
create_backup() {
    print_status "Creating backup of current state..."
    mkdir -p "$BACKUP_DIR"
    if [ -f "$STATE_FILE" ]; then
        cp "$STATE_FILE" "$BACKUP_DIR/terraform.tfstate.backup.$(date +%Y%m%d_%H%M%S)"
        print_status "Backup created in $BACKUP_DIR"
    else
        print_warning "No existing state file found"
    fi
}

# Function to check terraform and provider
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed"
        exit 1
    fi
    
    if ! terraform version | grep -q "Terraform"; then
        print_error "Terraform is not working properly"
        exit 1
    fi
    
    print_status "Prerequisites check passed"
}

# Function to initialize terraform
init_terraform() {
    print_status "Initializing Terraform..."
    terraform init
}

# Function to generate resource discovery configuration
generate_discovery_config() {
    print_status "Generating resource discovery configuration..."
    
    cat > discovery.tf << 'EOF'
# Temporary configuration for resource discovery
# This file will be used to discover existing resources

terraform {
  required_providers {
    spa = {
      source = "citrix/citrixspa"
    }
  }
}

provider "spa" {
  base_url      = var.base_url
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
  debug         = true  # Enable debug logging for discovery
}

# Variables for provider configuration
variable "base_url" {
  description = "Base URL for the SPA API"
  type        = string
  default     = "https://api.cloud.com"
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Citrix Cloud Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret"
  type        = string
  sensitive   = true
}

# Data sources for discovery
data "spa_applications" "all_apps" {
  offset = 0
  limit  = 100
}

output "discovered_applications" {
  value = data.spa_applications.all_apps.applications[*].id
  description = "IDs of all discovered applications"
}

output "discovered_application_names" {
  value = data.spa_applications.all_apps.applications[*].name
  description = "Names of all discovered applications"
}

# Note: Add more data sources as needed for other resource types
EOF
    
    print_status "Discovery configuration generated"
}

# Function to discover resources
discover_resources() {
    print_status "Discovering existing resources..."
    print_debug "Running terraform plan to refresh data sources..."
    
    # Run terraform plan to refresh data sources
    if [ "$DEBUG_LOGGING" = true ]; then
        terraform plan -refresh-only || true
    else
        terraform plan -refresh-only > /dev/null 2>&1 || true
    fi
    
    print_debug "Extracting application IDs from terraform output..."
    
    # Get application IDs
    if terraform output -json discovered_applications 2>/dev/null | jq -r '.[]' > discovered_apps.txt 2>/dev/null; then
        local app_count=$(wc -l < discovered_apps.txt)
        print_status "Discovered $app_count applications"
        if [ "$DEBUG_LOGGING" = true ] && [ "$app_count" -gt 0 ]; then
            print_debug "Discovered application IDs:"
            while IFS= read -r app_id; do
                print_debug "  - $app_id"
            done < discovered_apps.txt
        fi
    else
        print_warning "Could not discover applications (this is normal if none exist)"
        touch discovered_apps.txt
    fi
    
    # Get application names for reference
    if terraform output -json discovered_application_names 2>/dev/null | jq -r '.[]' > discovered_app_names.txt 2>/dev/null; then
        print_status "Application names saved to discovered_app_names.txt"
        if [ "$DEBUG_LOGGING" = true ]; then
            print_debug "Discovered application names:"
            while IFS= read -r app_name; do
                print_debug "  - $app_name"
            done < discovered_app_names.txt
        fi
    else
        touch discovered_app_names.txt
    fi
}

# Function to generate main configuration
generate_main_config() {
    print_status "Generating main Terraform configuration..."
    print_debug "Debug logging enabled: $DEBUG_LOGGING"
    
    # Remove discovery configuration
    rm -f discovery.tf
    
    # Generate main configuration based on discovered resources
    cat > main.tf << 'EOF'
terraform {
  required_providers {
    spa = {
      source = "citrix/citrixspa"
    }
  }
}

provider "spa" {
  base_url      = var.base_url
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
EOF
    
    # Add debug logging configuration if enabled
    if [ "$DEBUG_LOGGING" = true ]; then
        print_debug "Adding debug logging to provider configuration"
        cat >> main.tf << 'EOF'
  debug = true  # Debug logging enabled via command line option
EOF
    fi
    
    cat >> main.tf << 'EOF'
}

# Variables for provider configuration
variable "base_url" {
  description = "Base URL for the SPA API"
  type        = string
  default     = "https://api.cloud.com"
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Citrix Cloud Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret"
  type        = string
  sensitive   = true
}

# Import your existing resources here
# Example application resources will be generated below
EOF
    
    # Generate resource blocks for discovered applications
    local app_count=0
    while IFS= read -r app_id; do
        if [ ! -z "$app_id" ]; then
            app_count=$((app_count + 1))
            print_debug "Adding resource block for application: $app_id"
            cat >> main.tf << EOF

resource "spa_application" "imported_app_$app_count" {
  # This resource will be populated after import
  # Import with: terraform import spa_application.imported_app_$app_count $app_id
}
EOF
        fi
    done < discovered_apps.txt
    
    print_status "Generated configuration for $app_count applications"
    if [ "$DEBUG_LOGGING" = true ]; then
        print_debug "Debug logging has been enabled in the provider configuration"
    fi
}

# Function to generate import script
generate_import_script() {
    print_status "Generating import script..."
    print_debug "Creating import_resources.sh script..."
    
    cat > import_resources.sh << 'EOF'
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

EOF
    
    # Add application imports
    local app_count=0
    while IFS= read -r app_id; do
        if [ ! -z "$app_id" ]; then
            app_count=$((app_count + 1))
            print_debug "Adding import command for application: $app_id"
            cat >> import_resources.sh << EOF
print_status "Importing application: $app_id"
print_debug "Running: terraform import spa_application.imported_app_$app_count $app_id"
terraform import spa_application.imported_app_$app_count "$app_id" || print_error "Failed to import application $app_id"

EOF
        fi
    done < discovered_apps.txt
    
    # Add browser mode import
    print_debug "Adding browser mode import command"
    cat >> import_resources.sh << 'EOF'
print_status "Importing browser mode configuration"
print_debug "Running: terraform import spa_browser_mode.browser_mode browser_mode"
terraform import spa_browser_mode.browser_mode browser_mode || print_error "Failed to import browser mode"

print_status "Import script completed"
print_status "Run 'terraform plan' to verify the imported resources"
EOF
    
    chmod +x import_resources.sh
    print_status "Import script generated: import_resources.sh"
    print_debug "Import script contains $app_count application imports plus browser mode"
}

# Function to run imports
run_imports() {
    print_status "Running import script..."
    print_debug "Executing ./import_resources.sh"
    
    if [ -f "import_resources.sh" ]; then
        if [ "$DEBUG_LOGGING" = true ]; then
            # Show terraform output when debug is enabled
            ./import_resources.sh
        else
            # Suppress detailed output in normal mode
            ./import_resources.sh 2>/dev/null
        fi
    else
        print_error "Import script not found"
        exit 1
    fi
}

# Function to verify imports
verify_imports() {
    print_status "Verifying imports..."
    print_debug "Running terraform plan to check for configuration drift..."
    
    if [ "$DEBUG_LOGGING" = true ]; then
        # Show terraform plan output when debug is enabled
        if terraform plan -detailed-exitcode; then
            print_status "All imports successful - no configuration drift detected"
        else
            print_warning "Configuration drift detected - check with 'terraform plan'"
            print_warning "You may need to update your configuration to match the imported resources"
        fi
    else
        # Check silently in normal mode
        if terraform plan -detailed-exitcode > /dev/null 2>&1; then
            print_status "All imports successful - no configuration drift detected"
        else
            print_warning "Configuration drift detected - check with 'terraform plan'"
            print_warning "You may need to update your configuration to match the imported resources"
        fi
    fi
}

# Function to clean up temporary files
cleanup() {
    print_status "Cleaning up temporary files..."
    rm -f discovered_apps.txt discovered_app_names.txt
    print_status "Cleanup completed"
}

# Function to parse command line arguments
parse_arguments() {
    local command=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--debug)
                DEBUG_LOGGING=true
                shift
                ;;
            discover|import|help|-h|--help)
                command="$1"
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                print_error "Use '$0 --help' for usage information"
                exit 1
                ;;
        esac
    done
    
    # Set default command if none provided
    if [ -z "$command" ]; then
        command="discover"
    fi
    
    echo "$command"
}

# Function to show help
show_help() {
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  discover  - Discover existing resources and generate configuration (default)"
    echo "  import    - Run the import script"
    echo "  help      - Show this help message"
    echo ""
    echo "Options:"
    echo "  -d, --debug    - Enable debug logging"
    echo "  -h, --help     - Show this help message"
    echo ""
    echo "Example workflow:"
    echo "  1. $0 --debug discover"
    echo "  2. Review generated main.tf"
    echo "  3. $0 --debug import"
    echo ""
    echo "Debug logging will:"
    echo "  - Add 'debug = true' to the provider configuration"
    echo "  - Show detailed output during terraform operations"
    echo "  - Display additional information about discovered resources"
}

# Main execution
main() {
    # Parse command line arguments
    local command=$(parse_arguments "$@")
    
    print_status "Starting SPA Terraform import automation..."
    if [ "$DEBUG_LOGGING" = true ]; then
        print_debug "Debug logging is enabled"
        export TF_LOG=DEBUG
    fi
    
    # Execute the requested command
    case "$command" in
        "discover")
            check_prerequisites
            create_backup
            init_terraform
            generate_discovery_config
            discover_resources
            generate_main_config
            generate_import_script
            print_status "Discovery completed! Next steps:"
            print_status "1. Review the generated main.tf configuration"
            print_status "2. Run './import_resources.sh' to import resources"
            print_status "3. Run 'terraform plan' to verify imports"
            if [ "$DEBUG_LOGGING" = true ]; then
                print_debug "Debug logging has been enabled in the generated configuration"
            fi
            ;;
        "import")
            run_imports
            verify_imports
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
