terraform {
  required_version = ">= 0.13"
  required_providers {
    citrixspa = {
      source  = "citrix/citrixspa"
      version = "1.1.0"
    }
  }
}

provider "citrixspa" {
  base_url      = "https://api.cloud.com"
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
}

# Variables
variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Service Principal Client Secret"
  type        = string
  sensitive   = true
}

# Data sources for existing resources
data "citrixspa_applications" "existing_apps" {
  # List existing applications
}

# Create a new application
resource "citrixspa_application" "web_app" {
  name        = "My Web Application"
  description = "Web application managed by Terraform with Service Principal auth"
  # Add other required attributes based on your provider schema
}

# Create a security group
resource "citrixspa_security_group" "web_sg" {
  name        = "web-security-group"
  description = "Security group for web applications"
  # Add security rules and other attributes
}

# Create an access policy
resource "citrixspa_access_policy" "web_policy" {
  name        = "web-access-policy"
  description = "Access policy for web applications"
  # Configure access rules
}

# Outputs
output "application_id" {
  description = "ID of the created application"
  value       = citrixspa_application.web_app.id
}

output "security_group_id" {
  description = "ID of the created security group"
  value       = citrixspa_security_group.web_sg.id
}

output "access_policy_id" {
  description = "ID of the created access policy"
  value       = citrixspa_access_policy.web_policy.id
}

output "authentication_method" {
  description = "Authentication method used"
  value       = "Service Principal OAuth2"
}
