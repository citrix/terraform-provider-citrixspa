terraform {
  required_providers {
    spa = {
      source = "citrix/citrixspa"
    }
  }
}

provider "spa" {
  # Configuration options
  base_url    = "https://api.cloud.com"
  customer_id = var.citrix_customer_id
  auth_token  = var.citrix_auth_token
}

variable "citrix_customer_id" {
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}

variable "citrix_auth_token" {
  description = "Citrix Cloud Auth Token"
  type        = string
  sensitive   = true
}

# Example: Create a web application
resource "spa_application" "web_app" {
  name        = "My Web Application"
  type        = "web"
  description = "A sample web application"
  url         = "https://example.com"
  category    = "Productivity"
  
  # Required for non-ZTNA applications
  using_template = true
  related_urls   = ["https://example.com", "https://api.example.com"]
  
  # Optional settings
  hidden           = false
  agentless_access = true
  mobile_security  = true
  
  keywords = ["productivity", "web", "example"]
  
  # Optional icon (base64 encoded)
  # icon = "base64-encoded-icon-data"
}

# Example: Create a ZTNA application
resource "spa_application" "ztna_app" {
  name        = "Internal Database"
  type        = "ztna"
  description = "Internal database server"
  category    = "Database"
  
  # ZTNA specific destinations
  destination {
    destination = "database.internal.com"
    port        = "5432"
    protocol    = "PROTOCOL_TCP"
    subtype     = "SUBTYPE_HOSTNAME"
  }
  
  destination {
    destination = "10.0.1.0/24"
    port        = "3306"
    protocol    = "PROTOCOL_TCP"
    subtype     = "SUBTYPE_IP_AND_CIDR"
  }
  
  keywords = ["database", "internal", "ztna"]
}

# Example: Create a SaaS application
resource "spa_application" "saas_app" {
  name        = "Office 365"
  type        = "saas"
  description = "Microsoft Office 365 Suite"
  url         = "https://office.com"
  category    = "Productivity"
  
  using_template = true
  related_urls   = [
    "https://office.com",
    "https://outlook.office.com",
    "https://teams.microsoft.com"
  ]
  
  template_name = "Office365"
  keywords      = ["office", "productivity", "microsoft"]
}

# Example: Create an access policy
resource "spa_access_policy" "require_mfa" {
  name        = "Require MFA for Sensitive Apps"
  description = "Require multi-factor authentication for sensitive applications"
  type        = "access"
  enabled     = true
  priority    = 100
  
  conditions = {
    user_risk    = "high"
    device_trust = "untrusted"
  }
  
  actions = {
    require_mfa = "true"
    block_access = "false"
  }
}

# Example: Create a security group
resource "spa_security_group" "developers" {
  name        = "Developers"
  description = "Development team security group"
  type        = "user_group"
  
  users = [
    "user1@example.com",
    "user2@example.com"
  ]
  
  groups = [
    "CN=Developers,OU=Groups,DC=example,DC=com"
  ]
  
  rules {
    type     = "access"
    action   = "allow"
    priority = 1
    
    conditions = {
      time_of_day = "business_hours"
      location    = "office"
    }
  }
}

# Example: Create a routing domain
resource "spa_routing_domain" "internal_domain" {
  fqdn     = "internal.example.com"
  type     = "internal"
  app_type = "web"
  comment  = "Internal domain for web applications"
  flag     = "enabled"
  
  location_ids = [
    "location-uuid-1",
    "location-uuid-2"
  ]
}

# Example: Create a certificate
resource "spa_certificate" "app_cert" {
  certificate_name = "example-app-cert"
  certificate      = base64encode(file("${path.module}/certs/app.pfx"))
  
  # Optional: Assign to application
  application_id = spa_application.web_app.id
  domain         = "example.com"
}

# Example: Data sources to fetch existing resources
data "spa_application" "existing_app" {
  name = "Existing Application"
}

data "spa_access_policy" "existing_policy" {
  name = "Default Policy"
}

data "spa_security_group" "existing_group" {
  name = "All Users"
}

data "spa_routing_domain" "existing_domain" {
  fqdn = "api.example.com"
}

# Example: Data sources for new endpoints

# Get current browser mode
data "spa_browser_mode" "current" {}

# Get hybrid configuration
data "spa_hybrid_config" "current" {}

# Get last activity
data "spa_last_activity" "current" {}

# Example: Browser mode resource (read-only)
resource "spa_browser_mode" "browser_config" {
  browser_mode = "CEB"  # or "CEP"
}

# Example: Terminate machine access
resource "spa_terminate_machine_access" "machine_access" {
  name   = "my-machine-001"
  status = "active"
}

# Get terminate machine access by ID
data "spa_terminate_machine_access" "machine_info" {
  id = spa_terminate_machine_access.machine_access.id
}

# Output examples
output "web_app_id" {
  value = spa_application.web_app.id
}

output "web_app_state" {
  value = spa_application.web_app.state
}

output "existing_app_details" {
  value = {
    id          = data.spa_application.existing_app.id
    name        = data.spa_application.existing_app.name
    type        = data.spa_application.existing_app.type
    description = data.spa_application.existing_app.description
  }
}

output "browser_mode" {
  value = data.spa_browser_mode.current.browser_mode
}

output "is_hybrid" {
  value = data.spa_hybrid_config.current.is_hybrid
}

output "last_activity" {
  value = data.spa_last_activity.current.last_activity
}

output "machine_access_info" {
  value = {
    name      = data.spa_terminate_machine_access.machine_info.name
    status    = data.spa_terminate_machine_access.machine_info.status
    last_seen = data.spa_terminate_machine_access.machine_info.last_seen
  }
}
