#!/bin/bash

# Generate .terraformrc file with configurable plugin directory
# This script creates a .terraformrc file that uses the specified plugin directory
# or defaults to ~/.terraform/plugins

# Get the plugin directory from command line argument, environment variable, or use default
if [ -n "$1" ]; then
    TERRAFORM_PLUGIN_DIR="$1"
else
    TERRAFORM_PLUGIN_DIR=${TF_PLUGIN_DIR:-$HOME/.terraform.d/plugins}
fi

# Create the .terraformrc file
cat > .terraformrc << EOF
# Terraform CLI configuration for local development
# Place this file in your home directory as ~/.terraformrc or use TF_CLI_CONFIG_FILE

provider_installation {
  # Use ${TERRAFORM_PLUGIN_DIR} as an overridden package directory
  # for the citrix/spa provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # citrix/spa provider in the given directory.
  filesystem_mirror {
    path    = "${TERRAFORM_PLUGIN_DIR}"
    include = ["registry.terraform.io/citrix/spa"]
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {
    exclude = ["registry.terraform.io/citrix/spa"]
  }
}
EOF

echo "Generated .terraformrc using plugin directory: ${TERRAFORM_PLUGIN_DIR}"
