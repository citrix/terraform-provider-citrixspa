terraform {
  required_providers {
    spa = {
      source = "citrix/citrixspa"
    }
  }
}

provider "spa" {
  base_url      = var.base_url
  token_url     = var.token_url
  customer_id   = var.citrix_customer_id
  client_id     = var.citrix_client_id
  client_secret = var.citrix_client_secret
  rate_limit    = var.rate_limit
}

# Variables for provider configuration
variable "base_url" {
  description = "Base URL for the SPA API"
  type        = string
  default     = "https://api.cloud.com/accessSecurity"
}

variable "token_url" {
  description = "URL for the CC token API"
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

variable "rate_limit" {
  description = "Rate limit for Citrix API calls"
  type        = number
  default     = 15
}

# Import your existing resources here
# Example application resources will be generated below
