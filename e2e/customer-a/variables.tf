variable "base_url" {
  description = "SPA API Base URL."
  type        = string
  default     = "https://api.cloud.com/accessSecurity"
}

variable "token_url" {
  description = "Citrix Cloud trust endpoint for OAuth2 token requests."
  type        = string
  default     = "https://api.cloud.com"
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID (dedicated E2E test tenant)."
  type        = string
  sensitive   = true
}

variable "citrix_client_id" {
  description = "Citrix Cloud Service Principal Client ID."
  type        = string
  sensitive   = true
}

variable "citrix_client_secret" {
  description = "Citrix Cloud Service Principal Client Secret."
  type        = string
  sensitive   = true
}

variable "max_concurrent" {
  description = "Maximum concurrent mutating API requests enforced by the provider semaphore."
  type        = number
  default     = 3
}

variable "name_prefix" {
  description = "Prefix applied to every E2E-created resource. cleanup.sh sweeps resources whose name/fqdn starts with this value."
  type        = string
  default     = "e2e"
}

variable "run_tag" {
  description = "Unique per-run tag (e.g. the GitHub run id) appended to resource names to avoid collisions between concurrent runs."
  type        = string
  default     = "local"
}

# --- migrate scenario only -----------------------------------------------------
# Second (target) tenant credentials. These are NOT used by this Terraform
# configuration — they are read by tooling/run-migrate-e2e.sh from
# terraform.tfvars. They are declared here only so Terraform accepts the *_b
# keys in terraform.tfvars without an "Unexpected attribute" warning.
variable "citrix_customer_id_b" {
  description = "Second (target) tenant Customer ID. Used only by the migrate scenario."
  type        = string
  sensitive   = true
  default     = ""
}

variable "citrix_client_id_b" {
  description = "Second (target) tenant Service Principal Client ID. Used only by the migrate scenario."
  type        = string
  sensitive   = true
  default     = ""
}

variable "citrix_client_secret_b" {
  description = "Second (target) tenant Service Principal Client Secret. Used only by the migrate scenario."
  type        = string
  sensitive   = true
  default     = ""
}
