#!/usr/bin/env pwsh
<#
.SYNOPSIS
    SPA Manager - Comprehensive PowerShell replacement for list_resources.sh
    Combines resource discovery, terraform generation, and import functionality
    Platform-independent and no external dependencies required

.DESCRIPTION
    It discovers SPA resources via Terraform data sources and generates Terraform configurations.

.EXAMPLE
    ./spa_manager.ps1 -DebugOutput              # Enable debug mode
    ./spa_manager.ps1 -Level INFO               # Set debug logging level (TRACE, DEBUG, INFO, WARNING, ERROR, CRITICAL)
    ./spa_manager.ps1 -List                     # Discover and list resources (with detailed individual queries)
    ./spa_manager.ps1 -List -Quick              # Discover and list resources (quick mode, list data only)
    ./spa_manager.ps1 -List -Id 'id1' -Id 'id2' # Discover and list resources (with specific IDs)
    ./spa_manager.ps1 -Plan                     # Run Terraform plan
    ./spa_manager.ps1 -Apply                    # Run terraform apply after plan
    ./spa_manager.ps1 -Update                   # Show terraform plan update
    ./spa_manager.ps1 -UpdateDetail             # Show terraform plan update details
    ./spa_manager.ps1 -Enable apps -Enable policies  # Enable specific features (repeat for multiple)
    ./spa_manager.ps1 -Clean                    # Clean up temporary files and generated configurations
    ./spa_manager.ps1 -Limit 10                 # Set limit for resource queries (default: no limit)
    ./spa_manager.ps1 -Test                     # Run tests only
    ./spa_manager.ps1 -Setup                    # Setup terraform environment
    ./spa_manager.ps1 -Validate                 # Validate configuration only
    ./spa_manager.ps1 -VerboseOutput            # Verbose output

.NOTES
    Author: Citrix SPA Team
    Requires: PowerShell 7+
    Requires: Terraform CLI installed and in PATH
    
    Features:
    - Cross-platform PowerShell implementation
    - No external dependencies (except terraform)
    - Comprehensive error handling and validation
    - Enhanced data collection with individual item queries for complete field data
    - Built-in testing and validation functions
    - Available features: applications, access_policies, security_groups, certificates,
        browser_mode, routing_domains, hybrid_config, terminate_access, session_policies
#>

[CmdletBinding(DefaultParameterSetName = 'Default')]
param(
    [Parameter(ParameterSetName = 'List')]
    [Alias('l')]
    [switch]$List,

    [Parameter(ParameterSetName = 'Plan')]
    [Alias('p')]
    [switch]$Plan,

    [Parameter(ParameterSetName = 'Apply')]
    [Alias('a')]
    [switch]$Apply,

    [Parameter(ParameterSetName = 'Validate')]
    [Alias('v')]
    [switch]$Validate,

    [Parameter(ParameterSetName = 'Clean')]
    [Alias('x')]
    [switch]$Clean,

    [Parameter(ParameterSetName = 'Setup')]
    [Alias('s')]
    [switch]$Setup,

    [Parameter(ParameterSetName = 'Test')]
    [Alias('t')]
    [switch]$Test,

    [Parameter(ParameterSetName = 'Update')]
    [Alias('u')]
    [switch]$Update,

    [Parameter(ParameterSetName = 'UpdateDetail')]
    [Alias('ud')]
    [switch]$UpdateDetail,

    [Parameter()]
    [Alias('d')]
    [switch]$DebugOutput,

    [Parameter()]
    [ValidateSet('TRACE', 'DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL')]
    [Alias('lv')]
    [string]$Level = 'DEBUG',

    [Parameter()]
    [Alias('q')]
    [switch]$Quick,

    [Parameter()]
    [Alias('n')]
    [int]$Limit = -1,

    [Parameter()]
    [Alias('e')]
    [string[]]$Enable = @(),

    [Parameter()]
    [Alias('i')]
    [string[]]$Id = @(),

    [Parameter()]
    [switch]$VerboseOutput,

    [Parameter()]
    [string]$WorkDir = '.'
)

# Force UTF-8 for subprocess output so non-ASCII characters are not garbled on Windows.
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

# =============================================================================
# GLOBAL CONSTANTS
# =============================================================================

# Terraform parallelism setting for apply and plan operations
$script:TERRAFORM_PARALLELISM = 1

# =============================================================================
# TEMPLATE CONSTANTS
# =============================================================================

# Provider Configuration Templates
$script:TERRAFORM_PROVIDER_CONFIG = @'
terraform {{
  required_providers {{
    spa = {{
      source  = "registry.terraform.io/citrix/citrixspa"
      version = "0.8.1"
    }}
  }}
}}

provider "spa" {{
  base_url    = var.base_url
  token_url   = var.token_url
  customer_id = var.customer_id
  suppress_asb_notifications = var.suppress_asb_notifications
  {0}
  {1}
  {2}
}}

variable "base_url" {{
  description = "SPA API Base URL"
  type        = string
  default     = "https://api.cloud.com/accessSecurity"
}}

variable "token_url" {{
  description = "CC Base URL"
  type        = string
  default     = "https://api.cloud.com"
}}

variable "customer_id" {{
  description = "Citrix Cloud Customer ID"
  type        = string
  sensitive   = true
}}

{3}

{4}

{5}

variable "suppress_asb_notifications" {{
  description = "Suppress ASB notifications during API operations"
  type        = bool
  default     = false
}}

'@

$script:AUTH_TOKEN_VARIABLES = @'
variable "auth_token" {
    description = "Citrix Cloud Auth Token"
    type        = string
    sensitive   = true
}
'@

$script:CLIENT_CREDENTIALS_VARIABLES = @'
variable "client_id" {
    description = "Citrix Cloud Client ID"
    type        = string
    sensitive   = true
}

variable "client_secret" {
    description = "Citrix Cloud Client Secret"
    type        = string
    sensitive   = true
}
'@

$script:RATE_LIMIT_VARIABLES = @'
variable "rate_limit" {
    description = "Rate limit for API requests"
    type        = number
    default     = 15
}
'@

$script:LIST_DETAILS_VARIABLES = @'
variable "fetch_details_on_list" {
    description = "Fetch detailed information on list queries"
    type        = bool
    default     = true
}
'@

# Data Source Templates
$script:DATA_SOURCE_APPLICATION = @'
data "spa_application" "item" {{
  id = "{0}"
}}

output "item_data" {{
  value = data.spa_application.item
}}
'@

$script:DATA_SOURCE_ACCESS_POLICY = @'
data "spa_access_policy" "item" {{
  id = "{0}"
}}

output "item_data" {{
  value = data.spa_access_policy.item
}}
'@

$script:DATA_SOURCE_SECURITY_GROUP = @'
data "spa_security_group" "item" {{
  id = "{0}"
}}

output "item_data" {{
  value = data.spa_security_group.item
}}
'@

$script:DATA_SOURCE_ROUTING_DOMAIN = @'
data "spa_routing_domain" "item" {{
  fqdn = "{0}"
}}

output "item_data" {{
  value = data.spa_routing_domain.item
}}
'@

$script:DATA_SOURCE_APPLICATIONS_ALL = @'
data "spa_applications" "all" {{
{0}
}}

output "applications_data" {{
  value = data.spa_applications.all
}}
'@

$script:DATA_SOURCE_ACCESS_POLICIES_ALL = @'
data "spa_access_policies" "all" {{
{0}
}}

output "policies_data" {{
  value = data.spa_access_policies.all
}}
'@

$script:DATA_SOURCE_SECURITY_GROUPS_ALL = @'
data "spa_security_groups" "all" {{
{0}
}}

output "groups_data" {{
  value = data.spa_security_groups.all
}}
'@

$script:DATA_SOURCE_ROUTING_DOMAINS_ALL = @'
data "spa_routing_domains" "all" {{
{0}
}}

output "domains_data" {{
  value = data.spa_routing_domains.all
}}
'@

$script:DATA_SOURCE_CERTIFICATES_ALL = @'
data "spa_certificates" "all" {
}

output "certificates_data" {
  value = data.spa_certificates.all
}
'@

$script:DATA_SOURCE_BROWSER_MODE = @'
data "spa_browser_mode" "current" {
}

output "browser_mode_data" {
  value = data.spa_browser_mode.current
}
'@

$script:DATA_SOURCE_HYBRID_CONFIG = @'
data "spa_hybrid_config" "current" {
}

output "hybrid_config_data" {
  value = data.spa_hybrid_config.current
}
'@

$script:DATA_SOURCE_TERMINATE_MACHINE_ACCESS = @'
data "spa_terminate_machine_access" "all" {{
{0}
}}

output "terminate_machine_access_data" {{
  value = data.spa_terminate_machine_access.all
}}
'@

$script:DATA_SOURCE_TERMINATE_USER_ACCESS = @'
data "spa_terminate_user_access" "all" {{
{0}
}}

output "terminate_user_access_data" {{
  value = data.spa_terminate_user_access.all
}}
'@

$script:DATA_SOURCE_SESSION_POLICY = @'
data "spa_session_policy" "item" {{
  id = "{0}"
}}

output "item_data" {{
  value = data.spa_session_policy.item
}}
'@

$script:DATA_SOURCE_SESSION_POLICIES_ALL = @'
data "spa_session_policies" "all" {{
{0}
}}

output "session_policies_data" {{
  value = data.spa_session_policies.all
}}
'@

# Resource Templates
$script:RESOURCE_APPLICATION = @'
resource "spa_application" "{0}" {{
  name             = "{1}"
  type             = "{2}"{3}{4}
}}
'@

$script:RESOURCE_ACCESS_POLICY = @'
resource "spa_access_policy" "{0}" {{
  name        = "{1}"{2}
}}
'@

$script:RESOURCE_SECURITY_GROUP = @'
resource "spa_security_group" "{0}" {{
  name        = "{1}"
  app_ids     = {2}
  system = {{
    data_in  = "{3}"
    data_out = "{4}"
  }}
  unpublished_app = {{
    data_in  = "{5}"
    data_out = "{6}"
  }}{7}
}}
'@

$script:RESOURCE_BROWSER_MODE = @'
resource "spa_browser_mode" "{0}" {{
{1}
}}
'@

$script:RESOURCE_CERTIFICATE = @'
resource "spa_certificate" "{0}" {{
  certificate_name = "{1}"
  certificate      = "{2}"{3}
}}
'@

$script:RESOURCE_ROUTING_DOMAIN = @'
resource "spa_routing_domain" "{0}" {{
  fqdn        = "{1}"
  type        = "{2}"{3}
}}
'@

$script:RESOURCE_TERMINATE_MACHINE_ACCESS = @'
resource "spa_terminate_machine_access" "{0}" {{
    account_name  = "{1}"
    name          = "{2}"
    dns_host_name = "{3}"
    domain_name   = "{4}"
    object_id     = "{5}"
    idp_type      = "{6}"
    duration      = {7}
  }}
'@

$script:RESOURCE_TERMINATE_USER_ACCESS = @'
resource "spa_terminate_user_access" "{0}" {{
  account_name  = "{1}"
  email         = "{2}"
  domain_name   = "{3}"
  object_id     = "{4}"
  idp_type      = "{5}"
  duration      = {6}
}}
'@

# Import Templates
$script:IMPORT_BLOCKS_HEADER = @'
# Terraform Import Blocks
# Generated on {0}
# 
# This file contains import blocks for existing SPA resources.
# Run 'terraform plan' to generate configuration
# or run 'terraform apply' if you already have the resource configurations.
#
# To use these import blocks:
# 1. Run 'terraform plan' to auto-generate resource configs
# 2. Or ensure your spa_resources.tf file contains the matching resource configurations
# 3. Run 'terraform apply' to import the resources into state

{1}
'@

$script:IMPORT_BLOCK = @'
import {{
  to = {0}
  id = "{1}"
}}
'@

$script:MINIMAL_IMPORT_BLOCKS_HEADER = @'
# Terraform Import Blocks
# Generated on {0}
#
# NOTE: This workspace contains only data sources (discovery queries).
# Data sources do not require importing since they are read-only queries.
#
# Available data sources:{1}
#
# To use these data sources:
#   1. Run 'terraform validate' to verify configuration
#   2. Run 'terraform plan' to see what data will be queried
#   3. Run 'terraform apply' to execute the data queries
#
# No import blocks needed - this is a comment-only file.
'@

$script:SUMMARY_REPORT = @'

# SPA Resource Management Summary
Generated on: {0}

## Resource Counts:
- Applications: {1}
- Security Groups: {2}
- Access Policies: {3}
- Session Policies: {10}
- Certificates: {4}
- Browser Modes: {5}
- Routing Domains: {6}
- Hybrid Configs: {7}
- Terminate Machine Access: {8}
- Terminate User Access: {9}

Total Resources: {11}
Importable Resources: {12}
Total Unique Names: {13}

## Generated Files:
- spa_resources.tf ({11} total, {12} importable resources - configurable fields only)
- imports.tf ({12} import blocks)

## Data Processing:
- Uses in-memory data structures (no temporary JSON files)
- Avoids character encoding issues with newlines and special characters
- Data stored directly from Terraform queries without file I/O conversion

## Usage:
1. terraform validate                                  # Validate configuration
2. terraform plan                                      # Generate config if needed
3. terraform apply                                     # Import resources and review changes

## Notes:
- Resource blocks contain configurable fields only (read-only fields excluded)
- Terminate machine access records are fully manageable resources (create/read/update/delete)
- Terminate user access records are fully manageable resources (create/read/update/delete)
- Use 'terraform plan' to auto-generate resource configurations

## Platform Compatibility:
✓ Windows, macOS, Linux
✓ PowerShell 7+ required
✓ No external dependencies beyond terraform
'@

# =============================================================================
# SCRIPT-SCOPED STATE VARIABLES
# =============================================================================

# Working directory (resolved to absolute path)
$script:WorkDir = $null

# Main directory (parent of work dir)
$script:MainDir = $null

# Set of used names for uniqueness
$script:UsedNames = [System.Collections.Generic.HashSet[string]]::new()

# Resource counts
$script:ResourceCounts = @{
    applications            = 0
    security_groups         = 0
    access_policies         = 0
    certificates            = 0
    browser_modes           = 0
    routing_domains         = 0
    hybrid_configs          = 0
    terminate_machine_access = 0
    terminate_user_access   = 0
    session_policies        = 0
}

# In-memory data storage to replace temporary JSON files
$script:ResourceData = @{
    applications            = @{}
    access_policies         = @{}
    security_groups         = @{}
    routing_domains         = @{}
    certificates            = @{}
    browser_mode            = @{}
    hybrid_config           = @{}
    terminate_machine_access = @{}
    terminate_user_access   = @{}
    session_policies        = @{}
}

# Feature aliases for enable flag
$script:FeatureAliases = @{
    applications     = @('app', 'apps', 'application', 'applications')
    access_policies  = @('policy', 'policies', 'access_policy', 'access_policies')
    routing_domains  = @('routing', 'domain', 'domains', 'routing_domain', 'routing_domains')
    browser_mode     = @('browser', 'browser_mode')
    hybrid_config    = @('hybrid', 'hybrid_config')
    certificates     = @('cert', 'certificate', 'certificates')
    security_groups  = @('group', 'groups', 'security_group', 'security_groups')
    terminate_access = @('terminate', 'terminate_access', 'terminate_machine', 'terminate_user', 'terminate_machine_access', 'terminate_user_access')
    session_policies = @('session', 'session_policy', 'session_policies')
}

# List key mapping
$script:ListKeyMapping = @{
    applications            = 'applications'
    access_policies         = 'access_policies'
    security_groups         = 'security_groups'
    routing_domains         = 'routing_domains'
    certificates            = 'certificates'
    terminate_machine_access = 'machines'
    terminate_user_access   = 'users'
    session_policies        = 'session_policies'
}

# List type mapping
$script:ListTypeMapping = @{
    applications            = 'application'
    access_policies         = 'access_policy'
    security_groups         = 'security_group'
    routing_domains         = 'routing_domain'
    certificates            = 'certificate'
    terminate_machine_access = 'machine'
    terminate_user_access   = 'users'
    session_policies        = 'session_policy'
}

# Feature flags
$script:EnableApps = $true
$script:EnablePolicies = $true
$script:EnableRoutingDomains = $true
$script:EnableBrowserMode = $false  # Disabled for now
$script:EnableHybridConfig = $true
$script:EnableCertificates = $false  # Disabled for now
$script:EnableSecurityGroups = $true
$script:EnableTerminateAccess = $true

# Details flags for individual queries
$script:DetailsForApps = $true
$script:DetailsForPolicies = $true
$script:DetailsForRoutingDomains = $false
$script:DetailsForSecurityGroups = $false
$script:DetailsForCertificates = $false
$script:DetailsForSessionPolicies = $true

# Default limit (-1 means no limit)
$script:DefaultLimit = -1
$script:LimitValue = -1

# Debug settings
$script:DebugMode = $false
$script:DebugLevel = 'DEBUG'
$script:VerboseMode = $false

# Query enhancement settings
$script:QueryIndividualDetails = $true  # Default to enhanced behavior

# List details enabled (from provider config)
$script:ListDetailsEnabled = $true

# Plan output path
$script:PlanOutput = $null

# =============================================================================
# HELPER FUNCTIONS - OUTPUT
# =============================================================================

function Write-Status {
    <#
    .SYNOPSIS
        Print info message with color
    #>
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-WarningMessage {
    <#
    .SYNOPSIS
        Print warning message with color
    #>
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-ErrorMessage {
    <#
    .SYNOPSIS
        Print error message with color
    #>
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-SuccessMessage {
    <#
    .SYNOPSIS
        Print success message with color
    #>
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Header {
    <#
    .SYNOPSIS
        Print section header with color
    #>
    param([string]$Message)
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Blue
}

# =============================================================================
# HELPER FUNCTIONS - TEMPLATES
# =============================================================================

function Get-AuthConfig {
    <#
    .SYNOPSIS
        Get authentication configuration string
    #>
    param([bool]$HasAuthToken)
    
    if ($HasAuthToken) {
        return "auth_token = var.auth_token"
    }
    else {
        return "client_id = var.client_id`n  client_secret = var.client_secret"
    }
}

function Get-AuthVariables {
    <#
    .SYNOPSIS
        Get authentication variables string
    #>
    param([bool]$HasAuthToken)
    
    if ($HasAuthToken) {
        return $script:AUTH_TOKEN_VARIABLES
    }
    else {
        return $script:CLIENT_CREDENTIALS_VARIABLES
    }
}

function Get-RateLimitConfig {
    <#
    .SYNOPSIS
        Get rate limit configuration string
    #>
    param([bool]$HasRateLimit)
    
    if ($HasRateLimit) {
        return "rate_limit = var.rate_limit"
    }
    else {
        return ""
    }
}

function Get-RateLimitVariables {
    <#
    .SYNOPSIS
        Get rate limit variables string
    #>
    param([bool]$HasRateLimit)
    
    if ($HasRateLimit) {
        return $script:RATE_LIMIT_VARIABLES
    }
    else {
        return ""
    }
}

function Get-ListDetailsConfig {
    <#
    .SYNOPSIS
        Get list details config string
    #>
    param([bool]$Enabled)
    
    if ($Enabled) {
        return "fetch_details_on_list = var.fetch_details_on_list"
    }
    else {
        return ""
    }
}

function Get-ListDetailsVariables {
    <#
    .SYNOPSIS
        Get list details variables string
    #>
    param([bool]$Enabled)
    
    if ($Enabled) {
        return $script:LIST_DETAILS_VARIABLES
    }
    else {
        return ""
    }
}

function Format-TerraformProviderConfig {
    <#
    .SYNOPSIS
        Format complete provider configuration
    #>
    param(
        [bool]$HasAuthToken,
        [bool]$HasRateLimit,
        [bool]$ListDetails
    )
    
    $authConfig = Get-AuthConfig -HasAuthToken $HasAuthToken
    $rateLimitConfig = Get-RateLimitConfig -HasRateLimit $HasRateLimit
    $listDetailsConfig = Get-ListDetailsConfig -Enabled $ListDetails
    $authVariables = Get-AuthVariables -HasAuthToken $HasAuthToken
    $rateLimitVariables = Get-RateLimitVariables -HasRateLimit $HasRateLimit
    $listDetailsVariables = Get-ListDetailsVariables -Enabled $ListDetails
    
    return $script:TERRAFORM_PROVIDER_CONFIG -f $authConfig, $rateLimitConfig, $listDetailsConfig, $authVariables, $rateLimitVariables, $listDetailsVariables
}

function Format-DataSourceConfig {
    <#
    .SYNOPSIS
        Format data source configuration based on type
    #>
    param(
        [string]$DataSourceType,
        [hashtable]$Parameters = @{}
    )
    
    $templates = @{
        'application'              = $script:DATA_SOURCE_APPLICATION
        'access_policy'            = $script:DATA_SOURCE_ACCESS_POLICY
        'security_group'           = $script:DATA_SOURCE_SECURITY_GROUP
        'routing_domain'           = $script:DATA_SOURCE_ROUTING_DOMAIN
        'applications_all'         = $script:DATA_SOURCE_APPLICATIONS_ALL
        'access_policies_all'      = $script:DATA_SOURCE_ACCESS_POLICIES_ALL
        'security_groups_all'      = $script:DATA_SOURCE_SECURITY_GROUPS_ALL
        'routing_domains_all'      = $script:DATA_SOURCE_ROUTING_DOMAINS_ALL
        'certificates_all'         = $script:DATA_SOURCE_CERTIFICATES_ALL
        'browser_mode'             = $script:DATA_SOURCE_BROWSER_MODE
        'hybrid_config'            = $script:DATA_SOURCE_HYBRID_CONFIG
        'terminate_machine_access' = $script:DATA_SOURCE_TERMINATE_MACHINE_ACCESS
        'terminate_user_access'    = $script:DATA_SOURCE_TERMINATE_USER_ACCESS
        'session_policy'           = $script:DATA_SOURCE_SESSION_POLICY
        'session_policies_all'     = $script:DATA_SOURCE_SESSION_POLICIES_ALL
    }
    
    $template = $templates[$DataSourceType]
    if (-not $template) {
        return ''
    }
    
    # Check if template has format placeholders {0}, {1}, etc.
    # Templates without placeholders (like certificates_all, browser_mode, hybrid_config) are returned as-is
    if ($template -match '\{\d+\}') {
        # Has format placeholders - apply parameters
        switch ($DataSourceType) {
            'application' {
                return $template -f $Parameters['item_id']
            }
            'access_policy' {
                return $template -f $Parameters['item_id']
            }
            'security_group' {
                return $template -f $Parameters['item_id']
            }
            'routing_domain' {
                return $template -f $Parameters['item_fqdn']
            }
            'applications_all' {
                return $template -f $Parameters['limit_config']
            }
            'access_policies_all' {
                return $template -f $Parameters['limit_config']
            }
            'security_groups_all' {
                return $template -f $Parameters['limit_config']
            }
            'routing_domains_all' {
                return $template -f $Parameters['limit_config']
            }
            'terminate_machine_access' {
                return $template -f $Parameters['limit_config']
            }
            'terminate_user_access' {
                return $template -f $Parameters['limit_config']
            }
            'session_policy' {
                return $template -f $Parameters['item_id']
            }
            'session_policies_all' {
                return $template -f $Parameters['limit_config']
            }
            default {
                return $template
            }
        }
    }
    else {
        # No placeholders - return as-is
        return $template
    }
}

function Format-ResourceConfig {
    <#
    .SYNOPSIS
        Format resource configuration based on type
    #>
    param(
        [string]$ResourceType,
        [hashtable]$Parameters = @{}
    )
    
    $templates = @{
        'application'              = $script:RESOURCE_APPLICATION
        'access_policy'            = $script:RESOURCE_ACCESS_POLICY
        'security_group'           = $script:RESOURCE_SECURITY_GROUP
        'browser_mode'             = $script:RESOURCE_BROWSER_MODE
        'certificate'              = $script:RESOURCE_CERTIFICATE
        'routing_domain'           = $script:RESOURCE_ROUTING_DOMAIN
        'terminate_machine_access' = $script:RESOURCE_TERMINATE_MACHINE_ACCESS
        'terminate_user_access'    = $script:RESOURCE_TERMINATE_USER_ACCESS
    }
    
    $template = $templates[$ResourceType]
    if (-not $template) {
        return ''
    }
    
    # Apply parameters based on resource type
    switch ($ResourceType) {
        'application' {
            # {0}=safe_name, {1}=name, {2}=type, {3}=optional_fields, {4}=depends_on
            return $template -f $Parameters['safe_name'], $Parameters['name'], $Parameters['type'], $Parameters['optional_fields'], $Parameters['depends_on']
        }
        'access_policy' {
            # {0}=safe_name, {1}=name, {2}=optional_fields
            return $template -f $Parameters['safe_name'], $Parameters['name'], $Parameters['optional_fields']
        }
        'security_group' {
            # {0}=safe_name, {1}=name, {2}=app_ids, {3}=system_data_in, {4}=system_data_out, {5}=unpublished_app_data_in, {6}=unpublished_app_data_out, {7}=optional_fields
            return $template -f $Parameters['safe_name'], $Parameters['name'], $Parameters['app_ids'], $Parameters['system_data_in'], $Parameters['system_data_out'], $Parameters['unpublished_app_data_in'], $Parameters['unpublished_app_data_out'], $Parameters['optional_fields']
        }
        'browser_mode' {
            # {0}=safe_name, {1}=optional_fields
            return $template -f $Parameters['safe_name'], $Parameters['optional_fields']
        }
        'certificate' {
            # {0}=safe_name, {1}=name, {2}=certificate, {3}=optional_fields
            return $template -f $Parameters['safe_name'], $Parameters['name'], $Parameters['certificate'], $Parameters['optional_fields']
        }
        'routing_domain' {
            # {0}=safe_name, {1}=fqdn, {2}=type, {3}=optional_fields
            return $template -f $Parameters['safe_name'], $Parameters['fqdn'], $Parameters['type'], $Parameters['optional_fields']
        }
        'terminate_machine_access' {
            # {0}=safe_name, {1}=account_name, {2}=name, {3}=dns_host_name, {4}=domain_name, {5}=object_id, {6}=idp_type, {7}=duration
            return $template -f $Parameters['safe_name'], $Parameters['account_name'], $Parameters['name'], $Parameters['dns_host_name'], $Parameters['domain_name'], $Parameters['object_id'], $Parameters['idp_type'], $Parameters['duration']
        }
        'terminate_user_access' {
            # {0}=safe_name, {1}=account_name, {2}=email, {3}=domain_name, {4}=object_id, {5}=idp_type, {6}=duration
            return $template -f $Parameters['safe_name'], $Parameters['account_name'], $Parameters['email'], $Parameters['domain_name'], $Parameters['object_id'], $Parameters['idp_type'], $Parameters['duration']
        }
        default {
            return $template
        }
    }
}

# =============================================================================
# HELPER FUNCTIONS - UTILITIES
# =============================================================================

function Set-AllFeatures {
    <#
    .SYNOPSIS
        Set all feature flags to the same value
    #>
    param([bool]$Value)
    
    $script:EnableApps = $Value
    $script:EnablePolicies = $Value
    $script:EnableRoutingDomains = $Value
    $script:EnableBrowserMode = $false  # Disabled for now (matches Python)
    $script:EnableHybridConfig = $Value
    $script:EnableCertificates = $false  # Disabled for now (matches Python)
    $script:EnableSecurityGroups = $Value
    $script:EnableTerminateAccess = $Value
    $script:EnableSessionPolicies = $Value
}

function Test-Dependencies {
    <#
    .SYNOPSIS
        Check for required dependencies
    #>
    
    Write-Header "Checking Dependencies"
    
    $missingDeps = @()
    
    # Check PowerShell version
    $psVersion = $PSVersionTable.PSVersion
    if ($psVersion.Major -lt 7) {
        Write-ErrorMessage "PowerShell 7.0 or higher is required (current: $($psVersion.ToString()))"
        return $false
    }
    else {
        Write-Status "PowerShell $($psVersion.ToString()) ✓"
    }
    
    # Check for Terraform
    $terraformCmd = Get-Command terraform -ErrorAction SilentlyContinue
    if ($terraformCmd) {
        try {
            $tfVersionOutput = & terraform version 2>&1
            if ($LASTEXITCODE -eq 0) {
                $tfVersion = ($tfVersionOutput | Select-Object -First 1)
                Write-Status "$tfVersion ✓"
            }
            else {
                $missingDeps += "Terraform"
            }
        }
        catch {
            $missingDeps += "Terraform"
        }
    }
    else {
        $missingDeps += "Terraform"
    }
    
    if ($missingDeps.Count -gt 0) {
        Write-WarningMessage "Missing optional dependencies:"
        foreach ($dep in $missingDeps) {
            Write-WarningMessage "  - $dep"
        }
        Write-WarningMessage "Some features may be limited"
    }
    
    return $true
}

function Test-Credentials {
    <#
    .SYNOPSIS
        Check if credentials are configured
    #>
    
    $tfvarsPath = Join-Path $script:WorkDir "terraform.tfvars"
    
    if (-not (Test-Path $tfvarsPath)) {
        Write-WarningMessage "Credentials not found. Creating example file..."
        $examplePath = Join-Path $script:WorkDir "terraform.tfvars.example"
        if (Test-Path $examplePath) {
            Copy-Item $examplePath $tfvarsPath
        }
        Write-ErrorMessage "Please edit $tfvarsPath with your actual credentials"
        return $false
    }
    
    Write-Status "Credentials file found ✓"
    return $true
}

function Initialize-Terraform {
    <#
    .SYNOPSIS
        Initialize terraform for resource discovery
    #>
    
    Write-Header "Initializing Terraform"
    
    try {
        # Initialize terraform (assuming provider is already installed)
        Push-Location $script:WorkDir
        try {
            $result = & terraform init 2>&1
            $exitCode = $LASTEXITCODE
            
            if ($exitCode -ne 0) {
                Write-ErrorMessage "Terraform init failed: $result"
                return $false
            }
            
            Write-SuccessMessage "Terraform initialized successfully"
            return $true
        }
        finally {
            Pop-Location
        }
    }
    catch {
        Write-ErrorMessage "Failed to initialize terraform: $_"
        return $false
    }
}

function New-ProviderConfig {
    <#
    .SYNOPSIS
        Create base provider configuration
    #>
    
    Write-Header "Creating Provider Configuration"
    
    try {
        # Check authentication method
        $tfvarsPath = Join-Path $script:WorkDir "terraform.tfvars"
        $hasAuthToken = $false
        $hasClientCreds = $false
        $hasRateLimit = $false
        $listDetailsEnabled = $true  # Default to enabled
        
        if (Test-Path $tfvarsPath) {
            $lines = Get-Content $tfvarsPath
            foreach ($line in $lines) {
                $trimmedLine = $line.Trim()
                # Only check uncommented lines for authentication methods
                if ($trimmedLine -and -not $trimmedLine.StartsWith('#')) {
                    if ($trimmedLine -match 'auth_token' -and $trimmedLine -match '=') {
                        $hasAuthToken = $true
                    }
                    if ($trimmedLine -match 'client_id' -and $trimmedLine -match '=') {
                        $hasClientCreds = $true
                    }
                    if ($trimmedLine -match 'rate_limit' -and $trimmedLine -match '=') {
                        $hasRateLimit = $true
                    }
                    if ($trimmedLine -match 'fetch_details_on_list' -and $trimmedLine -match '=') {
                        $listDetailsEnabled = $trimmedLine -match 'true'
                    }
                }
            }
        }
        
        if ($hasAuthToken -and $hasClientCreds) {
            Write-ErrorMessage "Both auth_token and client_id/client_secret are configured. Please use only one authentication method."
            return $false
        }
        
        # Generate provider configuration
        $providerTfPath = Join-Path $script:WorkDir "provider.tf"
        
        if ($hasAuthToken) {
            Write-Status "Using auth_token authentication"
        }
        else {
            Write-Status "Using client credentials authentication"
        }
        
        $script:ListDetailsEnabled = $listDetailsEnabled
        if ($listDetailsEnabled) {
            $script:QueryIndividualDetails = $false  # Disable individual queries if list details are enabled
        }
        
        $providerConfig = Format-TerraformProviderConfig -HasAuthToken $hasAuthToken -HasRateLimit $hasRateLimit -ListDetails $listDetailsEnabled
        
        Set-Content -Path $providerTfPath -Value $providerConfig -Encoding UTF8
        
        # Format the file
        Push-Location $script:WorkDir
        try {
            & terraform fmt $providerTfPath 2>&1 | Out-Null
        }
        finally {
            Pop-Location
        }
        
        Write-SuccessMessage "Provider configuration created"
        return $true
    }
    catch {
        Write-ErrorMessage "Failed to create provider config: $_"
        return $false
    }
}

function Get-LimitConfig {
    <#
    .SYNOPSIS
        Generate limit parameter configuration
    #>
    
    if ($script:LimitValue -le 0) {
        return ""
    }
    return "  limit = $($script:LimitValue)"
}

function Invoke-TerraformQuery {
    <#
    .SYNOPSIS
        Run terraform with given configuration and capture output in memory
    #>
    param(
        [string]$ConfigContent,
        [string]$DataKey
    )
    
    # Define temp file path
    $tempTf = Join-Path $script:WorkDir "temp_$DataKey.tf"
    
    try {
        # Create temporary terraform file
        Set-Content -Path $tempTf -Value $ConfigContent -Encoding UTF8
        
        # Set up environment
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        if ($script:DebugMode) {
            $env:TF_LOG_PROVIDER = $script:DebugLevel
            $env:TF_LOG_PATH = Join-Path $script:WorkDir "debug-terraform.log"
        }
        
        # Run terraform apply and capture output
        Push-Location $script:WorkDir
        try {
            $applyResult = & terraform apply -auto-approve 2>&1
            $applyExitCode = $LASTEXITCODE
            
            if ($applyExitCode -ne 0) {
                Write-WarningMessage "Terraform apply for $DataKey failed: $applyResult"
                return $false
            }
            
            # Extract the actual data from terraform output
            $outputResult = & terraform output -json 2>&1
            $outputExitCode = $LASTEXITCODE
            
            if ($outputExitCode -ne 0) {
                Write-WarningMessage "Terraform output for $DataKey failed: $outputResult"
                return $false
            }
            
            # Parse the terraform output JSON (use -Depth 100 for nested structures)
            try {
                $tfOutputs = $outputResult | ConvertFrom-Json -Depth 100
                
                # Extract the specific output we're looking for
                $outputKey = $null
                if ($null -ne $tfOutputs.applications_data) { $outputKey = 'applications_data' }
                elseif ($null -ne $tfOutputs.policies_data) { $outputKey = 'policies_data' }
                elseif ($null -ne $tfOutputs.groups_data) { $outputKey = 'groups_data' }
                elseif ($null -ne $tfOutputs.domains_data) { $outputKey = 'domains_data' }
                elseif ($null -ne $tfOutputs.certificates_data) { $outputKey = 'certificates_data' }
                elseif ($null -ne $tfOutputs.browser_mode_data) { $outputKey = 'browser_mode_data' }
                elseif ($null -ne $tfOutputs.hybrid_config_data) { $outputKey = 'hybrid_config_data' }
                elseif ($null -ne $tfOutputs.terminate_machine_access_data) { $outputKey = 'terminate_machine_access_data' }
                elseif ($null -ne $tfOutputs.terminate_user_access_data) { $outputKey = 'terminate_user_access_data' }
                elseif ($null -ne $tfOutputs.session_policies_data) { $outputKey = 'session_policies_data' }
                
                if ($outputKey -and $null -ne $tfOutputs.$outputKey) {
                    # Extract the value from the terraform output structure
                    $extractedData = $tfOutputs.$outputKey.value
                    if ($null -eq $extractedData) {
                        $extractedData = @{}
                    }
                    
                    # Add metadata and store in memory
                    $finalData = @{
                        timestamp = (Get-Date).ToString('o')
                        source = 'terraform_query'
                    }
                    
                    # Merge extracted data into final data
                    if ($extractedData -is [PSCustomObject]) {
                        $extractedData.PSObject.Properties | ForEach-Object {
                            $finalData[$_.Name] = $_.Value
                        }
                    }
                    elseif ($extractedData -is [hashtable]) {
                        foreach ($key in $extractedData.Keys) {
                            $finalData[$key] = $extractedData[$key]
                        }
                    }
                    
                    # Store data in memory
                    $script:ResourceData[$DataKey] = $finalData
                    
                    if ($script:DebugMode) {
                        Write-Status "Stored data for $DataKey in memory"
                    }
                }
                else {
                    # Fallback to empty structure
                    $finalData = @{
                        timestamp = (Get-Date).ToString('o')
                        source = 'terraform_query'
                        note = "No data available for $DataKey"
                    }
                    
                    # Store empty data in memory
                    $script:ResourceData[$DataKey] = $finalData
                }
                
                if ($script:DebugMode -and $script:VerboseMode) {
                    Write-Status "Extracted: key=$DataKey"
                }
            }
            catch {
                Write-WarningMessage "Failed to parse terraform output JSON: $_"
                $finalData = @{
                    timestamp = (Get-Date).ToString('o')
                    source = 'terraform_query'
                    error = "JSON parse error: $_"
                }
                
                # Store error data in memory
                $script:ResourceData[$DataKey] = $finalData
            }
            
            return $true
        }
        finally {
            Pop-Location
        }
    }
    catch {
        Write-ErrorMessage "Failed to run terraform query: $_"
        return $false
    }
    finally {
        # Clean up temp file
        if (Test-Path $tempTf) { Remove-Item $tempTf -Force -ErrorAction SilentlyContinue }
        $planOutputPath = $script:PlanOutput
        if ($planOutputPath -and (Test-Path $planOutputPath)) { Remove-Item $planOutputPath -Force -ErrorAction SilentlyContinue }
        $tfstatePath = Join-Path $script:WorkDir "terraform.tfstate"
        if (Test-Path $tfstatePath) { Remove-Item $tfstatePath -Force -ErrorAction SilentlyContinue }
    }
}

function Get-IndividualItem {
    <#
    .SYNOPSIS
        Query individual item details using single-item data sources
    #>
    param(
        [string]$ItemType,
        [string]$ItemId,
        [string]$ItemName = $null
    )
    
    try {
        # Create temporary terraform file for individual item query
        $tempTf = Join-Path $script:WorkDir "temp_individual_${ItemType}_${ItemId}.tf"
        
        # Generate appropriate data source configuration based on item type
        $config = $null
        switch ($ItemType) {
            'application' {
                $config = Format-DataSourceConfig -DataSourceType 'application' -Parameters @{ item_id = $ItemId }
            }
            'access_policy' {
                $config = Format-DataSourceConfig -DataSourceType 'access_policy' -Parameters @{ item_id = $ItemId }
            }
            'security_group' {
                $config = Format-DataSourceConfig -DataSourceType 'security_group' -Parameters @{ item_id = $ItemId }
            }
            'routing_domain' {
                # Use fqdn for routing domains (that's what the data source expects)
                $fqdn = if ($ItemName) { $ItemName } else { $ItemId }
                $config = Format-DataSourceConfig -DataSourceType 'routing_domain' -Parameters @{ item_fqdn = $fqdn }
            }
            'session_policy' {
                $config = Format-DataSourceConfig -DataSourceType 'session_policy' -Parameters @{ item_id = $ItemId }
            }
            default {
                # Unsupported item type for individual queries
                if ($script:DebugMode) {
                    Write-WarningMessage "Individual query not supported for item type: $ItemType"
                }
                return $null
            }
        }
        
        Set-Content -Path $tempTf -Value $config -Encoding UTF8
        
        # Set up environment
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        if ($script:DebugMode) {
            $env:TF_LOG_PROVIDER = $script:DebugLevel
            $env:TF_LOG_PATH = Join-Path $script:WorkDir "debug-terraform-individual-$ItemType.log"
        }
        
        # Run terraform apply and capture output
        Push-Location $script:WorkDir
        try {
            $applyResult = & terraform apply -auto-approve 2>&1
            $applyExitCode = $LASTEXITCODE
            
            if ($applyExitCode -ne 0) {
                if ($script:DebugMode) {
                    Write-WarningMessage "Individual query for $ItemType $ItemId failed: $applyResult"
                }
                return $null
            }
            
            # Extract the actual data from terraform output
            $outputResult = & terraform output -json 2>&1
            $outputExitCode = $LASTEXITCODE
            
            if ($outputExitCode -ne 0) {
                if ($script:DebugMode) {
                    Write-WarningMessage "Individual query output for $ItemType $ItemId failed: $outputResult"
                }
                return $null
            }
            
            # Parse the terraform output JSON
            try {
                $tfOutputs = $outputResult | ConvertFrom-Json -Depth 100
                
                if ($null -ne $tfOutputs.item_data) {
                    $extractedData = $tfOutputs.item_data.value
                    if ($script:DebugMode) {
                        Write-Status "Successfully queried individual $ItemType`: $ItemId"
                    }
                    return $extractedData
                }
                else {
                    if ($script:DebugMode) {
                        Write-WarningMessage "No item_data in output for $ItemType`: $ItemId"
                    }
                    return $null
                }
            }
            catch {
                if ($script:DebugMode) {
                    $context = "$ItemType '$($ItemName ?? $ItemId)' (ID: $ItemId)"
                    Write-WarningMessage "Failed to parse individual query output JSON for $context`: $_"
                }
                return $null
            }
        }
        finally {
            Pop-Location
        }
    }
    catch {
        if ($script:DebugMode) {
            Write-WarningMessage "Error in individual query for $ItemType $ItemId`: $_"
        }
        return $null
    }
    finally {
        # Clean up
        $tempTf = Join-Path $script:WorkDir "temp_individual_${ItemType}_${ItemId}.tf"
        if (Test-Path $tempTf) { Remove-Item $tempTf -Force -ErrorAction SilentlyContinue }
        $tfstatePath = Join-Path $script:WorkDir "terraform.tfstate"
        if (Test-Path $tfstatePath) { Remove-Item $tfstatePath -Force -ErrorAction SilentlyContinue }
    }
}

function Update-ResourceDataWithIndividualQueries {
    <#
    .SYNOPSIS
        Enhance list data by querying individual items for complete details
    #>
    param([string]$DataKey)
    
    if (-not $script:QueryIndividualDetails) {
        return $true  # Skip enhancement if disabled
    }
    
    $data = $script:ResourceData[$DataKey]
    if (-not $data -or $data.ContainsKey('error')) {
        return $true  # Nothing to enhance
    }
    
    $listKey = $script:ListKeyMapping[$DataKey]
    if (-not $listKey -or -not $data.ContainsKey($listKey)) {
        return $true  # No list to enhance
    }
    
    $itemType = $script:ListTypeMapping[$DataKey]
    if (-not $itemType) {
        return $false
    }
    
    $itemsList = $data[$listKey]
    if (-not ($itemsList -is [array]) -or $itemsList.Count -eq 0) {
        return $true  # Nothing to enhance
    }
    
    $enhancedItems = @()
    $totalItems = $itemsList.Count
    
    Write-Status "Enhancing $DataKey data by querying $totalItems individual items..."
    
    for ($i = 0; $i -lt $itemsList.Count; $i++) {
        $item = $itemsList[$i]
        
        if (-not ($item -is [PSCustomObject]) -and -not ($item -is [hashtable])) {
            $enhancedItems += $item
            continue
        }
        
        # Get item properties (handle both PSCustomObject and hashtable)
        $itemId = $null
        $itemName = $null
        
        if ($item -is [PSCustomObject]) {
            $itemId = $item.id
            $itemName = $item.name
        }
        else {
            $itemId = $item['id']
            $itemName = $item['name']
        }
        
        # Special handling for routing domains - use fqdn as the identifier
        if ($itemType -eq 'routing_domain') {
            if ($item -is [PSCustomObject]) {
                $itemName = if ($item.fqdn) { $item.fqdn } else { $itemName }
                if (-not $itemId) { $itemId = $item.fqdn }
            }
            else {
                $itemName = if ($item['fqdn']) { $item['fqdn'] } else { $itemName }
                if (-not $itemId) { $itemId = $item['fqdn'] }
            }
        }
        
        if (-not $itemId) {
            if ($script:DebugMode) {
                Write-WarningMessage "No ID found for $itemType item"
            }
            $enhancedItems += $item
            continue
        }
        
        if ($script:VerboseMode) {
            Write-Status "  Querying $itemType $($i+1)/$totalItems`: $($itemName ?? $itemId)"
        }
        
        # Query individual item details
        $detailedItem = Get-IndividualItem -ItemType $itemType -ItemId $itemId -ItemName $itemName
        
        if ($detailedItem) {
            # Use the detailed data instead of the list data
            $enhancedItems += $detailedItem
            if ($script:DebugMode) {
                Write-Status "  Enhanced $itemType`: $($itemName ?? $itemId)"
            }
        }
        else {
            # Fall back to original item if individual query failed
            $enhancedItems += $item
            if ($script:DebugMode) {
                Write-WarningMessage "  Failed to enhance $itemType`: $($itemName ?? $itemId), using list data"
            }
        }
    }
    
    # Update the data with enhanced items
    $data[$listKey] = $enhancedItems
    $script:ResourceData[$DataKey] = $data
    
    Write-SuccessMessage "Enhanced $DataKey data with individual queries ($($enhancedItems.Count) items)"
    return $true
}

# =============================================================================
# RESOURCE LISTING FUNCTIONS
# =============================================================================

function Get-Applications {
    <#
    .SYNOPSIS
        List applications and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableApps) {
        return $true
    }
    
    Write-Header "Listing Applications"
    
    $success = $false
    
    if ($Ids -and $Ids.Count -gt 0) {
        $data = @()
        for ($i = 0; $i -lt $Ids.Count; $i++) {
            $data += @{ id = $Ids[$i]; name = "app_$($i+1)" }
        }
        $script:ResourceData['applications'] = @{
            timestamp = (Get-Date).ToString('o')
            source = 'manual_query'
            applications = $data
        }
        $success = $true
    }
    else {
        $limitConfig = Get-LimitConfig
        $config = Format-DataSourceConfig -DataSourceType 'applications_all' -Parameters @{ limit_config = $limitConfig }
        
        $success = Invoke-TerraformQuery -ConfigContent $config -DataKey 'applications'
    }
    
    if ($success -and $script:QueryIndividualDetails -and $script:DetailsForApps) {
        # Enhance with individual queries
        Update-ResourceDataWithIndividualQueries -DataKey 'applications'
    }
    
    return $success
}

function Get-AccessPolicies {
    <#
    .SYNOPSIS
        List access policies and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnablePolicies) {
        return $true
    }
    
    Write-Header "Listing Access Policies"
    
    $success = $false
    
    if ($Ids -and $Ids.Count -gt 0) {
        $data = @()
        for ($i = 0; $i -lt $Ids.Count; $i++) {
            $data += @{ id = $Ids[$i]; name = "policy_$($i+1)" }
        }
        $script:ResourceData['access_policies'] = @{
            timestamp = (Get-Date).ToString('o')
            source = 'manual_query'
            access_policies = $data
        }
        $success = $true
    }
    else {
        $limitConfig = Get-LimitConfig
        $config = Format-DataSourceConfig -DataSourceType 'access_policies_all' -Parameters @{ limit_config = $limitConfig }
        
        $success = Invoke-TerraformQuery -ConfigContent $config -DataKey 'access_policies'
    }
    
    if ($success -and $script:QueryIndividualDetails -and $script:DetailsForPolicies) {
        # Enhance with individual queries
        Update-ResourceDataWithIndividualQueries -DataKey 'access_policies'
    }
    
    return $success
}

function Get-SessionPolicies {
    <#
    .SYNOPSIS
        List session policies and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableSessionPolicies) {
        return $true
    }
    
    Write-Header "Listing Session Policies"
    
    $success = $false
    
    if ($Ids -and $Ids.Count -gt 0) {
        $data = @()
        for ($i = 0; $i -lt $Ids.Count; $i++) {
            $data += @{ id = $Ids[$i]; name = "session_policy_$($i+1)" }
        }
        $script:ResourceData['session_policies'] = @{
            timestamp        = (Get-Date).ToString('o')
            source           = 'manual_query'
            session_policies = $data
        }
        $success = $true
    }
    else {
        $limitConfig = Get-LimitConfig
        $config = Format-DataSourceConfig -DataSourceType 'session_policies_all' -Parameters @{ limit_config = $limitConfig }
        
        $success = Invoke-TerraformQuery -ConfigContent $config -DataKey 'session_policies'
    }
    
    if ($success -and $script:QueryIndividualDetails -and $script:DetailsForSessionPolicies) {
        # Enhance with individual queries to get full rule/condition detail
        Update-ResourceDataWithIndividualQueries -DataKey 'session_policies'
    }
    
    return $success
}

function Get-SecurityGroups {
    <#
    .SYNOPSIS
        List security groups and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableSecurityGroups) {
        return $true
    }
    
    Write-Header "Listing Security Groups"
    
    $success = $false
    
    if ($Ids -and $Ids.Count -gt 0) {
        $data = @()
        for ($i = 0; $i -lt $Ids.Count; $i++) {
            $data += @{ id = $Ids[$i]; name = "sg_$($i+1)" }
        }
        $script:ResourceData['security_groups'] = @{
            timestamp = (Get-Date).ToString('o')
            source = 'manual_query'
            security_groups = $data
        }
        $success = $true
    }
    else {
        $limitConfig = Get-LimitConfig
        $config = Format-DataSourceConfig -DataSourceType 'security_groups_all' -Parameters @{ limit_config = $limitConfig }
        
        $success = Invoke-TerraformQuery -ConfigContent $config -DataKey 'security_groups'
    }
    
    if ($success -and $script:QueryIndividualDetails -and $script:DetailsForSecurityGroups) {
        # Enhance with individual queries
        Update-ResourceDataWithIndividualQueries -DataKey 'security_groups'
    }
    
    return $success
}

function Get-RoutingDomains {
    <#
    .SYNOPSIS
        List routing domains and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableRoutingDomains) {
        return $true
    }
    
    Write-Header "Listing Routing Domains"
    
    $success = $false
    
    if ($Ids -and $Ids.Count -gt 0) {
        $data = @()
        for ($i = 0; $i -lt $Ids.Count; $i++) {
            $data += @{ id = $Ids[$i]; name = "routing_$($i+1)" }
        }
        $script:ResourceData['routing_domains'] = @{
            timestamp = (Get-Date).ToString('o')
            source = 'manual_query'
            routing_domains = $data
        }
        $success = $true
    }
    else {
        $limitConfig = Get-LimitConfig
        $config = Format-DataSourceConfig -DataSourceType 'routing_domains_all' -Parameters @{ limit_config = $limitConfig }
        
        $success = Invoke-TerraformQuery -ConfigContent $config -DataKey 'routing_domains'
    }
    
    if ($success -and $script:QueryIndividualDetails -and $script:DetailsForRoutingDomains) {
        # Enhance with individual queries
        Update-ResourceDataWithIndividualQueries -DataKey 'routing_domains'
    }
    
    return $success
}

function Get-Certificates {
    <#
    .SYNOPSIS
        List certificates and generate JSON data
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableCertificates) {
        return $true
    }
    
    Write-Header "Listing Certificates"
    
    $config = Format-DataSourceConfig -DataSourceType 'certificates_all'
    
    return Invoke-TerraformQuery -ConfigContent $config -DataKey 'certificates'
}

function Get-BrowserMode {
    <#
    .SYNOPSIS
        Get browser mode configuration
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableBrowserMode) {
        return $true
    }
    
    Write-Header "Getting Browser Mode Configuration"
    
    $config = Format-DataSourceConfig -DataSourceType 'browser_mode'
    
    return Invoke-TerraformQuery -ConfigContent $config -DataKey 'browser_mode'
}

function Get-HybridConfig {
    <#
    .SYNOPSIS
        Get hybrid configuration
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableHybridConfig) {
        return $true
    }
    
    Write-Header "Getting Hybrid Configuration"
    
    $config = Format-DataSourceConfig -DataSourceType 'hybrid_config'
    
    return Invoke-TerraformQuery -ConfigContent $config -DataKey 'hybrid_config'
}

function Get-TerminateMachineAccess {
    <#
    .SYNOPSIS
        List terminate machine access resources
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableTerminateAccess) {
        return $true
    }
    
    Write-Header "Listing Terminate Machine Access"
    
    $limitConfig = Get-LimitConfig
    $config = Format-DataSourceConfig -DataSourceType 'terminate_machine_access' -Parameters @{ limit_config = $limitConfig }
    
    return Invoke-TerraformQuery -ConfigContent $config -DataKey 'terminate_machine_access'
}

function Get-TerminateUserAccess {
    <#
    .SYNOPSIS
        List terminate user access resources
    #>
    param([string[]]$Ids = @())
    
    if (-not $script:EnableTerminateAccess) {
        return $true
    }
    
    Write-Header "Listing Terminate User Access"
    
    $limitConfig = Get-LimitConfig
    $config = Format-DataSourceConfig -DataSourceType 'terminate_user_access' -Parameters @{ limit_config = $limitConfig }
    
    return Invoke-TerraformQuery -ConfigContent $config -DataKey 'terminate_user_access'
}

function Save-TerraformEnv {
    <#
    .SYNOPSIS
        Save and clear environment variables that can corrupt terraform invocations.
    .DESCRIPTION
        TF_LOG, TF_LOG_CORE  - cause timestamped log lines to bleed into captured stderr,
                               corrupting 'terraform output -json' parsing.
        TF_CLI_ARGS_output   - can change output format away from JSON (e.g. -raw).
        TF_CLI_ARGS          - injects flags into every subcommand, including ones that
                               don't support those flags (e.g. -json injected into fmt).
        Returns a hashtable that must be passed to Restore-TerraformEnv when done.
    #>
    $saved = @{
        TF_LOG             = $env:TF_LOG
        TF_LOG_CORE        = $env:TF_LOG_CORE
        TF_CLI_ARGS_output = $env:TF_CLI_ARGS_output
        TF_CLI_ARGS        = $env:TF_CLI_ARGS
    }
    $env:TF_LOG             = $null
    $env:TF_LOG_CORE        = $null
    $env:TF_CLI_ARGS_output = $null
    $env:TF_CLI_ARGS        = $null
    return $saved
}

function Restore-TerraformEnv {
    <#
    .SYNOPSIS
        Restore environment variables previously saved by Save-TerraformEnv.
    #>
    param([hashtable]$Saved)
    $env:TF_LOG             = $Saved['TF_LOG']
    $env:TF_LOG_CORE        = $Saved['TF_LOG_CORE']
    $env:TF_CLI_ARGS_output = $Saved['TF_CLI_ARGS_output']
    $env:TF_CLI_ARGS        = $Saved['TF_CLI_ARGS']
}

function Find-Resources {
    <#
    .SYNOPSIS
        Main resource discovery function
    #>
    param([string[]]$Ids = @())
    
    Write-Header "Discovering SPA Resources"
    
    $discoveryFunctions = @(
        { param($ids) Get-Applications -Ids $ids },
        { param($ids) Get-AccessPolicies -Ids $ids },
        { param($ids) Get-SessionPolicies -Ids $ids },
        { param($ids) Get-SecurityGroups -Ids $ids },
        { param($ids) Get-RoutingDomains -Ids $ids },
        { param($ids) Get-Certificates -Ids $ids },
        { param($ids) Get-BrowserMode -Ids $ids },
        { param($ids) Get-HybridConfig -Ids $ids },
        { param($ids) Get-TerminateMachineAccess -Ids $ids },
        { param($ids) Get-TerminateUserAccess -Ids $ids }
    )
    
    $successCount = 0
    foreach ($func in $discoveryFunctions) {
        try {
            if (& $func $Ids) {
                $successCount++
            }
            else {
                Write-WarningMessage "Failed to discover resources"
            }
        }
        catch {
            Write-WarningMessage "Error in discovery: $_"
        }
    }
    
    Write-Status "Completed $successCount/$($discoveryFunctions.Count) discovery operations"
    return $successCount -gt 0
}

# =============================================================================
# RESOURCE GENERATION HELPER FUNCTIONS
# =============================================================================

function Get-DomainFromUrl {
    <#
    .SYNOPSIS
        Extract domain from a URL string
    .EXAMPLE
        Get-DomainFromUrl "https://testapp.com" # Returns: testapp.com
    #>
    param([string]$Url)
    
    if (-not $Url) {
        return ""
    }
    
    # If it's already just a domain pattern (with or without wildcard), return as-is
    if (-not $Url.StartsWith('http://') -and -not $Url.StartsWith('https://')) {
        return $Url
    }
    
    # Parse URL to extract domain
    try {
        $uri = [System.Uri]::new($Url)
        $netloc = $uri.Host
        return $netloc
    }
    catch {
        # Fallback: simple string manipulation
        $url = $Url -replace '^https?://', ''
        $domainWithPort = ($url -split '/')[0]
        # Strip port if present
        if ($domainWithPort -match ':') {
            $domainWithPort = ($domainWithPort -split ':')[0]
        }
        return $domainWithPort
    }
}

function Get-DestinationsFromApp {
    <#
    .SYNOPSIS
        Extract destination values from application's destination field
    #>
    param([object]$App)
    
    $destinations = @()
    
    # Get destination list from app (handle both PSCustomObject and hashtable)
    $destinationList = $null
    if ($App -is [PSCustomObject]) {
        $destinationList = $App.destination
    }
    elseif ($App -is [hashtable]) {
        $destinationList = $App['destination']
    }
    
    if ($destinationList -is [array]) {
        foreach ($destItem in $destinationList) {
            $destValue = $null
            if ($destItem -is [PSCustomObject]) {
                $destValue = $destItem.destination
            }
            elseif ($destItem -is [hashtable]) {
                $destValue = $destItem['destination']
            }
            
            if ($destValue -and $destValue -notin $destinations) {
                $destinations += $destValue
            }
        }
    }
    
    return $destinations
}

function Get-SafeTerraformName {
    <#
    .SYNOPSIS
        Convert name to safe terraform resource name and ensure uniqueness
    #>
    param(
        [string]$Name,
        [string]$Prefix = 'resource'
    )
    
    $safe = $Name -replace '\*', 'wildcard'
    # Replace non-alphanumeric with underscore
    $safe = $safe -replace '[^a-zA-Z0-9_]', '_'
    # Remove multiple consecutive underscores
    $safe = $safe -replace '_+', '_'
    # Strip leading/trailing underscores
    $safe = $safe.Trim('_')
    # Ensure it doesn't start with a number
    if ($safe -and $safe[0] -match '\d') {
        $safe = "${Prefix}_$safe"
    }
    
    # Handle duplicates by adding counter
    $originalSafe = $safe
    $counter = 1
    while ($script:UsedNames.ContainsKey($safe)) {
        $safe = "${originalSafe}_$counter"
        $counter++
    }
    
    $script:UsedNames[$safe] = $true
    return $safe
}

function ConvertTo-EscapedTerraformString {
    <#
    .SYNOPSIS
        Escape string for terraform configuration
    #>
    param([object]$Value)
    
    if ($null -eq $Value) {
        return ""
    }
    
    $s = [string]$Value
    # Escape backslashes first (before other escapes)
    $s = $s -replace '\\', '\\'
    # Escape quotes
    $s = $s -replace '"', '\"'
    # Escape newlines
    $s = $s -replace "`r`n", '\n'
    $s = $s -replace "`n", '\n'
    $s = $s -replace "`r", '\r'
    return $s
}

function ConvertTo-TerraformStringList {
    <#
    .SYNOPSIS
        Print list of strings as a single string using double quotes
    #>
    param([string[]]$Values)
    
    if ($null -eq $Values) {
        return "null"
    }
    
    $items = @()
    foreach ($item in $Values) {
        if ($item) {
            $items += "`"$item`""
        }
    }
    
    $joined = $items -join ','
    if ($joined) {
        return "[$joined]"
    }
    else {
        return "[]"
    }
}

function ConvertTo-HclDict {
    <#
    .SYNOPSIS
        Format a PowerShell hashtable as HCL syntax with proper double quotes
    #>
    param([object]$Dict)
    
    if ($null -eq $Dict) {
        return "null"
    }
    
    # Convert PSCustomObject to hashtable if needed
    $hashTable = @{}
    if ($Dict -is [PSCustomObject]) {
        $Dict.PSObject.Properties | ForEach-Object {
            $hashTable[$_.Name] = $_.Value
        }
    }
    elseif ($Dict -is [hashtable]) {
        $hashTable = $Dict
    }
    else {
        # If it's not a dict, try to escape it as a string
        return "`"$(ConvertTo-EscapedTerraformString $Dict)`""
    }
    
    if ($hashTable.Count -eq 0) {
        return "{}"
    }
    
    # Convert dictionary to HCL format with proper escaping
    $items = @()
    foreach ($key in $hashTable.Keys) {
        $value = $hashTable[$key]
        
        # Skip null/None values - they should not appear in Terraform configuration
        if ($null -eq $value) {
            continue
        }
        
        $escapedKey = ConvertTo-EscapedTerraformString $key
        
        # Handle different value types
        if ($value -is [bool]) {
            $escapedValue = $value.ToString().ToLower()
            $items += "`"$escapedKey`" = $escapedValue"
        }
        elseif ($value -is [int] -or $value -is [long] -or $value -is [double] -or $value -is [float]) {
            $items += "`"$escapedKey`" = $value"
        }
        elseif ($value -is [hashtable] -or $value -is [PSCustomObject]) {
            # Nested dictionary - recursively format
            $nestedHcl = ConvertTo-HclDict $value
            $items += "`"$escapedKey`" = $nestedHcl"
        }
        elseif ($value -is [array]) {
            # List values
            $listItems = @()
            foreach ($item in $value) {
                if ($item -is [hashtable] -or $item -is [PSCustomObject]) {
                    $listItems += ConvertTo-HclDict $item
                }
                else {
                    $listItems += "`"$(ConvertTo-EscapedTerraformString $item)`""
                }
            }
            $listStr = "[$($listItems -join ', ')]"
            $items += "`"$escapedKey`" = $listStr"
        }
        else {
            # String or other types - treat as string
            $escapedValue = ConvertTo-EscapedTerraformString $value
            $items += "`"$escapedKey`" = `"$escapedValue`""
        }
    }
    
    return "{ $($items -join ', ') }"
}

function ConvertTo-HclMap {
    <#
    .SYNOPSIS
        Format a PowerShell hashtable as HCL map syntax with unquoted keys
    #>
    param([object]$Dict)
    
    if ($null -eq $Dict) {
        return "null"
    }
    
    # Convert PSCustomObject to hashtable if needed
    $hashTable = @{}
    if ($Dict -is [PSCustomObject]) {
        $Dict.PSObject.Properties | ForEach-Object {
            $hashTable[$_.Name] = $_.Value
        }
    }
    elseif ($Dict -is [hashtable]) {
        $hashTable = $Dict
    }
    else {
        # If it's not a dict, try to escape it as a string
        return "`"$(ConvertTo-EscapedTerraformString $Dict)`""
    }
    
    if ($hashTable.Count -eq 0) {
        return "{}"
    }
    
    # Convert dictionary to HCL format with unquoted keys for map interpretation
    $items = @()
    foreach ($key in $hashTable.Keys) {
        $value = $hashTable[$key]
        
        # Skip null/None values - they should not appear in Terraform configuration
        if ($null -eq $value) {
            continue
        }
        
        # Use unquoted key for map interpretation (key must be valid identifier)
        $safeKey = [string]$key
        
        # Handle different value types
        if ($value -is [bool]) {
            $escapedValue = $value.ToString().ToLower()
            $items += "$safeKey = $escapedValue"
        }
        elseif ($value -is [int] -or $value -is [long] -or $value -is [double] -or $value -is [float]) {
            $items += "$safeKey = $value"
        }
        elseif ($value -is [hashtable] -or $value -is [PSCustomObject]) {
            # Nested dictionary - recursively format as map
            $nestedHcl = ConvertTo-HclMap $value
            $items += "$safeKey = $nestedHcl"
        }
        elseif ($value -is [array]) {
            # List values
            $listItems = @()
            foreach ($item in $value) {
                if ($item -is [hashtable] -or $item -is [PSCustomObject]) {
                    $listItems += ConvertTo-HclMap $item
                }
                else {
                    $listItems += "`"$(ConvertTo-EscapedTerraformString $item)`""
                }
            }
            $listStr = "[$($listItems -join ', ')]"
            $items += "$safeKey = $listStr"
        }
        else {
            # String or other types - treat as string
            $escapedValue = ConvertTo-EscapedTerraformString $value
            $items += "$safeKey = `"$escapedValue`""
        }
    }
    
    return "{ $($items -join ', ') }"
}

function ConvertTo-SsoNormalized {
    <#
    .SYNOPSIS
        Normalize SSO hashtable values: parse JSON-encoded strings into native arrays
        and recursively convert PSCustomObject items to hashtables for ConvertTo-HclMap.
        API already returns snake_case keys so no key renaming is needed.
    #>
    param([hashtable]$SsoDict)

    if ($null -eq $SsoDict -or $SsoDict.Count -eq 0) {
        return $SsoDict
    }

    $result = @{}
    foreach ($k in $SsoDict.Keys) {
        $val = $SsoDict[$k]

        # API quirk: custom_attributes (and potentially other fields) may arrive
        # as a JSON-encoded string like "[]" or "[{...}]" instead of a native array.
        # Parse it into a real array so ConvertTo-HclMap emits [] not "[]".
        if ($val -is [string] -and $val.TrimStart().StartsWith('[')) {
            try {
                $parsed = $val | ConvertFrom-Json -Depth 100
                # ConvertFrom-Json returns $null for "[]" in some PS versions; normalise to empty array
                if ($null -eq $parsed) { $val = @() } else { $val = @($parsed) }
            }
            catch {
                # Not valid JSON — keep the original string value
            }
        }

        # Recursively normalize array items (e.g. custom_attributes list entries)
        if ($val -is [array]) {
            $converted = @()
            foreach ($item in $val) {
                if ($item -is [hashtable]) {
                    $converted += ConvertTo-SsoNormalized $item
                }
                elseif ($item -is [PSCustomObject]) {
                    $h = @{}
                    $item.PSObject.Properties | ForEach-Object { $h[$_.Name] = $_.Value }
                    $converted += ConvertTo-SsoNormalized $h
                }
                else {
                    $converted += $item
                }
            }
            $result[$k] = $converted
        }
        elseif ($val -is [PSCustomObject]) {
            $h = @{}
            $val.PSObject.Properties | ForEach-Object { $h[$_.Name] = $_.Value }
            $result[$k] = ConvertTo-SsoNormalized $h
        }
        else {
            $result[$k] = $val
        }
    }
    return $result
}

# =============================================================================
# RESOURCE GENERATION FUNCTIONS
# =============================================================================

function New-ApplicationResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_application (without read-only fields)
    
    .DESCRIPTION
        This method handles different field types appropriately:
        - Retains empty strings (e.g., description = "")
        - Retains empty lists/maps when explicitly set to []/{} 
        - Excludes null/None values to let provider use defaults
        - Adds depends_on blocks for routing domain dependencies
    
    .PARAMETER App
        Application data dictionary/object
    
    .PARAMETER RoutingDomainFqdnToResource
        Mapping of routing domain FQDNs to Terraform resource names
    #>
    param(
        [object]$App,
        [hashtable]$RoutingDomainFqdnToResource = @{}
    )
    
    # Helper function to get property value from PSCustomObject or hashtable
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = Get-PropValue $App 'name' ''
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'app'
    
    # Required fields
    $appId = ConvertTo-EscapedTerraformString (Get-PropValue $App 'id' '')
    $appName = ConvertTo-EscapedTerraformString (Get-PropValue $App 'name' '')
    $appType = ConvertTo-EscapedTerraformString (Get-PropValue $App 'type' '')
    
    # Handle optional fields
    $optionalFields = @()
    
    # Special handling for description field
    $description = Get-PropValue $App 'description'
    if ($null -ne $description) {
        $optionalFields += "  description      = `"$(ConvertTo-EscapedTerraformString $description)`""
    }
    
    # Optional string fields
    $stringFields = @(
        @('url', 'url'),
        @('category', 'category'),
        @('template_name', 'template_name'),
        @('icon', 'icon'),
        @('state', 'state')
    )
    foreach ($field in $stringFields) {
        $fieldVal = Get-PropValue $App $field[0]
        if ($null -ne $fieldVal) {
            $tfName = $field[1].PadRight(15)
            $optionalFields += "  $tfName = `"$(ConvertTo-EscapedTerraformString $fieldVal)`""
        }
    }
    
    # Boolean attributes
    $boolFields = @(
        @('hidden', 'hidden'),
        @('agentless_access', 'agentless_access'),
        @('mobile_security', 'mobile_security'),
        @('sbs_only_launch', 'sbs_only_launch'),
        @('using_template', 'using_template')
    )
    foreach ($field in $boolFields) {
        $fieldVal = Get-PropValue $App $field[0]
        if ($null -ne $fieldVal) {
            $tfName = $field[1].PadRight(15)
            $optionalFields += "  $tfName = $($fieldVal.ToString().ToLower())"
        }
    }
    
    # List fields
    $listFields = @(
        @('related_urls', 'related_urls'),
        @('keywords', 'keywords')
    )
    foreach ($field in $listFields) {
        $fieldVal = Get-PropValue $App $field[0]
        if ($null -ne $fieldVal) {
            $tfName = $field[1].PadRight(15)
            $optionalFields += "  $tfName = $(ConvertTo-TerraformStringList $fieldVal)"
        }
    }
    
    # Handle locations
    $locations = Get-PropValue $App 'locations'
    # Force to array to handle single-element arrays from JSON
    $locationsArray = @($locations)
    if ($null -ne $locations -and $locationsArray.Count -gt 0) {
        $locationBlocks = @()
        foreach ($location in $locationsArray) {
            if ($location -is [PSCustomObject] -or $location -is [hashtable]) {
                $locName = ConvertTo-EscapedTerraformString (Get-PropValue $location 'name' '')
                $locUuid = ConvertTo-EscapedTerraformString (Get-PropValue $location 'uuid' '')
                $locationBlocks += "    {`n      name = `"$locName`"`n      uuid = `"$locUuid`"`n    }"
            }
            else {
                $locStr = ConvertTo-EscapedTerraformString $location
                $locationBlocks += "    {`n      name = `"$locStr`"`n      uuid = `"`"`n    }"
            }
        }
        $locationBlocksStr = $locationBlocks -join ",`n"
        $optionalFields += "  locations       = [`n$locationBlocksStr`n  ]"
    }
    
    # Handle SSO field
    $appSso = Get-PropValue $App 'sso'
    # Server-computed SSO fields that must NOT appear in generated Terraform config.
    # Including them causes "Provider produced inconsistent result after apply" errors
    # because the server overwrites them with computed values on every read.
    $ssoComputedFields = @('customer', 'application_id', 'saml_sso_login_url', 'saml_cert_issuer_name')
    if ($null -ne $appSso) {
        if ($appSso -is [hashtable] -or $appSso -is [PSCustomObject]) {
            # Filter out server-computed fields from sso
            $ssoFiltered = @{}
            if ($appSso -is [PSCustomObject]) {
                $appSso.PSObject.Properties | Where-Object { $_.Name -notin $ssoComputedFields } | ForEach-Object {
                    $ssoFiltered[$_.Name] = $_.Value
                }
            }
            else {
                foreach ($k in $appSso.Keys) {
                    if ($k -notin $ssoComputedFields) { $ssoFiltered[$k] = $appSso[$k] }
                }
            }
            $ssoFiltered = ConvertTo-SsoNormalized $ssoFiltered
            $ssoHcl = ConvertTo-HclMap $ssoFiltered
            $optionalFields += "  sso             = $ssoHcl"
        }
        elseif ($appSso -is [string] -and $appSso.Trim()) {
            if ($script:ListDetailsEnabled) {
                try {
                    $ssoDict = $appSso | ConvertFrom-Json -Depth 100
                    $ssoFiltered = @{}
                    if ($ssoDict -is [PSCustomObject]) {
                        $ssoDict.PSObject.Properties | Where-Object { $_.Name -notin $ssoComputedFields } | ForEach-Object {
                            $ssoFiltered[$_.Name] = $_.Value
                        }
                    }
                    $ssoFiltered = ConvertTo-SsoNormalized $ssoFiltered
                    $ssoHcl = ConvertTo-HclMap $ssoFiltered
                    $optionalFields += "  sso             = $ssoHcl"
                }
                catch {
                    $appIdVal = Get-PropValue $App 'id' 'unknown-id'
                    $appNameVal = Get-PropValue $App 'name' 'unknown-name'
                    Write-Host "[WARNING] Invalid JSON in SSO field for application '$appNameVal' (ID: $appIdVal), skipping: $_" -ForegroundColor Yellow
                }
            }
            else {
                $optionalFields += "  sso             = { type = `"$(ConvertTo-EscapedTerraformString $appSso)`" }"
            }
        }
    }
    
    # Handle destination field
    $appDestination = Get-PropValue $App 'destination'
    # Force to array to handle single-element arrays from JSON
    $destinationArray = @($appDestination)
    if ($null -ne $appDestination -and $destinationArray.Count -gt 0) {
        $destinationBlocks = @()
        foreach ($destItem in $destinationArray) {
            if ($destItem -is [PSCustomObject] -or $destItem -is [hashtable]) {
                $destItems = @()
                $props = if ($destItem -is [PSCustomObject]) { $destItem.PSObject.Properties } else { $destItem.GetEnumerator() }
                foreach ($prop in $props) {
                    $k = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Key } else { $prop.Name }
                    $v = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Value } else { $prop.Value }
                    $escapedKey = ConvertTo-EscapedTerraformString $k
                    $escapedValue = ConvertTo-EscapedTerraformString $v
                    $destItems += "      $escapedKey = `"$escapedValue`""
                }
                $destinationBlocks += "    {`n$($destItems -join "`n")`n    }"
            }
        }
        $destinationBlocksStr = $destinationBlocks -join ",`n"
        $optionalFields += "  destination     = [`n$destinationBlocksStr`n  ]"
    }
    
    # Handle custom properties map
    $customProperties = Get-PropValue $App 'custom_properties'
    if ($null -ne $customProperties) {
        if (($customProperties -is [hashtable] -and $customProperties.Count -gt 0) -or ($customProperties -is [PSCustomObject])) {
            $mapItems = @()
            $props = if ($customProperties -is [PSCustomObject]) { $customProperties.PSObject.Properties } else { $customProperties.GetEnumerator() }
            foreach ($prop in $props) {
                $k = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Key } else { $prop.Name }
                $v = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Value } else { $prop.Value }
                $escapedKey = ConvertTo-EscapedTerraformString $k
                if ($v -is [hashtable] -or $v -is [array] -or $v -is [PSCustomObject]) {
                    $escapedValue = ConvertTo-EscapedTerraformString ($v | ConvertTo-Json -Compress -Depth 100)
                }
                else {
                    $escapedValue = ConvertTo-EscapedTerraformString $v
                }
                $mapItems += "`"$escapedKey`" = `"$escapedValue`""
            }
            $optionalFields += "  custom_properties             = {`n    $($mapItems -join ', ')`n  }"
        }
        else {
            $optionalFields += "  custom_properties             = {}"
        }
    }
    
    # Handle customer domain fields map
    $customerDomainFields = Get-PropValue $App 'customer_domain_fields'
    if ($null -ne $customerDomainFields) {
        if (($customerDomainFields -is [hashtable] -and $customerDomainFields.Count -gt 0) -or ($customerDomainFields -is [PSCustomObject])) {
            $mapItems = @()
            $props = if ($customerDomainFields -is [PSCustomObject]) { $customerDomainFields.PSObject.Properties } else { $customerDomainFields.GetEnumerator() }
            foreach ($prop in $props) {
                $k = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Key } else { $prop.Name }
                $v = if ($prop -is [System.Collections.DictionaryEntry]) { $prop.Value } else { $prop.Value }
                $escapedKey = ConvertTo-EscapedTerraformString $k
                $escapedValue = ConvertTo-EscapedTerraformString $v
                $mapItems += "`"$escapedKey`" = `"$escapedValue`""
            }
            $optionalFields += "  customer_domain_fields             = {`n    $($mapItems -join ', ')`n  }"
        }
        else {
            $optionalFields += "  customer_domain_fields             = {}"
        }
    }
    
    # Detect routing domain dependencies
    $routingDomainDependencies = @()
    
    # Check related_urls field
    $relatedUrls = Get-PropValue $App 'related_urls' @()
    if ($relatedUrls -and $RoutingDomainFqdnToResource.Count -gt 0) {
        foreach ($url in $relatedUrls) {
            if ($RoutingDomainFqdnToResource.ContainsKey($url)) {
                $rdResourceName = $RoutingDomainFqdnToResource[$url]
                $routingDomainDependencies += "spa_routing_domain.$rdResourceName"
            }
        }
    }
    
    # Check url field
    $appUrl = Get-PropValue $App 'url'
    if ($appUrl -and $RoutingDomainFqdnToResource.Count -gt 0) {
        $domain = Get-DomainFromUrl $appUrl
        if ($domain -and $RoutingDomainFqdnToResource.ContainsKey($domain)) {
            $rdResourceName = $RoutingDomainFqdnToResource[$domain]
            $routingDomainDependencies += "spa_routing_domain.$rdResourceName"
        }
    }
    
    # Check destination field
    $destinations = Get-DestinationsFromApp $App
    if ($destinations -and $RoutingDomainFqdnToResource.Count -gt 0) {
        foreach ($dest in $destinations) {
            if ($RoutingDomainFqdnToResource.ContainsKey($dest)) {
                $rdResourceName = $RoutingDomainFqdnToResource[$dest]
                $routingDomainDependencies += "spa_routing_domain.$rdResourceName"
            }
        }
    }
    
    # Add depends_on block if we found routing domain dependencies
    $dependsOnBlock = ''
    if ($routingDomainDependencies.Count -gt 0) {
        $depsList = $routingDomainDependencies -join ",`n    "
        $dependsOnBlock = "`n`n  depends_on = [`n    $depsList`n  ]"
    }
    
    # Build the resource block
    $optionalPart = ''
    if ($optionalFields.Count -gt 0) {
        $optionalPart = "`n" + ($optionalFields -join "`n")
    }
    
    $resource = Format-ResourceConfig -ResourceType 'application' -Parameters @{
        safe_name = $safeName
        name = $appName
        type = $appType
        optional_fields = $optionalPart
        depends_on = $dependsOnBlock
    }
    
    return @($safeName, $resource)
}

function New-AccessRuleBlock {
    <#
    .SYNOPSIS
        Generate terraform block for an individual access rule
    #>
    param([object]$Rule)
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $ruleAttrs = @()
    
    # Required fields
    $name = Get-PropValue $Rule 'name' ''
    if ($name) {
        $ruleAttrs += "      name = `"$(ConvertTo-EscapedTerraformString $name)`""
    }
    
    $priority = Get-PropValue $Rule 'priority'
    if ($null -ne $priority) {
        $ruleAttrs += "      priority = $priority"
    }
    
    $active = Get-PropValue $Rule 'active'
    if ($null -ne $active) {
        $ruleAttrs += "      active = $($active.ToString().ToLower())"
    }
    
    $access = Get-PropValue $Rule 'access' ''
    if ($access) {
        $ruleAttrs += "      access = `"$access`""
    }
    
    $accessNative = Get-PropValue $Rule 'access_native' ''
    if ($accessNative) {
        $ruleAttrs += "      access_native = `"$accessNative`""
    }
    
    # Optional fields
    $description = Get-PropValue $Rule 'description'
    if ($null -ne $description) {
        $ruleAttrs += "      description = `"$(ConvertTo-EscapedTerraformString $description)`""
    }
    
    $accessNativeCamel = Get-PropValue $Rule 'accessNative'
    if ($accessNativeCamel) {
        $ruleAttrs += "      access_native = `"$accessNativeCamel`""
    }
    
    # Advanced settings
    $advancedSettings = Get-PropValue $Rule 'advanced_settings'
    if ($advancedSettings -and ($advancedSettings -is [PSCustomObject] -or $advancedSettings -is [hashtable])) {
        $domainOverrides = Get-PropValue $advancedSettings 'domain_overrides' @()
        if ($domainOverrides -and $domainOverrides.Count -gt 0) {
            $overrideBlocks = @()
            foreach ($override in $domainOverrides) {
                $locationIds = Get-PropValue $override 'location_ids' @()
                $locationIdsStr = ($locationIds | ForEach-Object { "`"$_`"" }) -join ', '
                $overrideBlock = @"
        {
          fqdn         = "$(Get-PropValue $override 'fqdn' '')"
          location_ids = [$locationIdsStr]
          type         = "$(Get-PropValue $override 'type' '')"
        }
"@
                $overrideBlocks += $overrideBlock
            }
            $domainOverridesContent = $overrideBlocks -join ",`n"
            $ruleAttrs += @"
      advanced_settings = {
        domain_overrides = [
$domainOverridesContent
        ]
      }
"@
        }
    }
    
    # Conditions
    $conditions = Get-PropValue $Rule 'conditions' @()
    if ($conditions -and $conditions.Count -gt 0) {
        $conditionBlocks = @()
        foreach ($condition in $conditions) {
            $conditionBlock = @"
        {
          platform_filter = "$(Get-PropValue $condition 'platform_filter' '')"
          user_and_groups = {}
        }
"@
            $conditionBlocks += $conditionBlock
        }
        $conditionsContent = $conditionBlocks -join ",`n"
        $ruleAttrs += @"
      conditions = [
$conditionsContent
      ]
"@
    }
    
    # Restrictions
    $restrictions = Get-PropValue $Rule 'restrictions'
    if ($restrictions -and ($restrictions -is [PSCustomObject] -or $restrictions -is [hashtable])) {
        $redirectSbs = Get-PropValue $restrictions 'redirect_sbs' $false
        $enhancedSecuritySettings = Get-PropValue $restrictions 'enhanced_security_settings' @{}
        $enhancedHcl = ConvertTo-HclDict $enhancedSecuritySettings
        $ruleAttrs += @"
      restrictions = {
        redirect_sbs                = $($redirectSbs.ToString().ToLower())
        enhanced_security_settings = $enhancedHcl
      }
"@
    }
    
    # Rules (required)
    $rules = Get-PropValue $Rule 'rules' @()
    if ($rules -and $rules.Count -gt 0) {
        $ruleBlocks = @()
        foreach ($r in $rules) {
            $values = Get-PropValue $r 'values' @()
            $valuesStr = ($values | ForEach-Object { "`"$_`"" }) -join ', '
            
            # Handle both camelCase (API) and snake_case (terraform) field names
            $tagSource = Get-PropValue $r 'tag_source'
            if (-not $tagSource) { $tagSource = Get-PropValue $r 'tagSource' '' }
            $tagKey = Get-PropValue $r 'tag_key'
            if (-not $tagKey) { $tagKey = Get-PropValue $r 'tagKey' '' }
            
            $metadata = Get-PropValue $r 'metadata' @{}
            $metadataHcl = ConvertTo-HclDict $metadata
            
            $ruleBlock = @"
        {
          type       = "$(Get-PropValue $r 'type' '')"
          operator   = "$(Get-PropValue $r 'operator' '')"
          tag_source = "$tagSource"
          tag_key    = "$tagKey"
          values     = [$valuesStr]
          metadata   = $metadataHcl
        }
"@
            $ruleBlocks += $ruleBlock
        }
        $rulesContent = $ruleBlocks -join ",`n"
        $ruleAttrs += @"
      rules = [
$rulesContent
      ]
"@
    }
    else {
        $ruleAttrs += '      rules = []'
    }
    
    # Build the complete rule block
    $ruleContent = $ruleAttrs -join ",`n"
    return @"
    {
$ruleContent
    }
"@
}

function New-AccessPolicyResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_access_policy (without read-only fields)
    
    .DESCRIPTION
        This method handles different field types appropriately:
        - Retains empty strings (e.g., description = "")
        - Excludes null/None values to let provider use defaults
    
    .PARAMETER Policy
        The access policy data from the API
    
    .PARAMETER AppIdToResourceName
        Optional mapping of application IDs to terraform resource names
        (e.g., @{"app-uuid-123" = "app_salesforce"})
        Used to replace hardcoded IDs with resource references
    #>
    param(
        [object]$Policy,
        [hashtable]$AppIdToResourceName = @{}
    )
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = Get-PropValue $Policy 'name' ''
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'ap'
    
    # Required fields
    $policyId = ConvertTo-EscapedTerraformString (Get-PropValue $Policy 'id' '')
    $policyName = ConvertTo-EscapedTerraformString (Get-PropValue $Policy 'name' '')
    
    # Handle optional fields
    $optionalFields = @()
    
    # Special handling for description field
    $description = Get-PropValue $Policy 'description'
    if ($null -ne $description) {
        $optionalFields += "  description = `"$(ConvertTo-EscapedTerraformString $description)`""
    }
    
    # Boolean field
    $active = Get-PropValue $Policy 'active'
    if ($null -ne $active) {
        $optionalFields += "  active     = $($active.ToString().ToLower())"
    }
    
    # Numeric field
    $priority = Get-PropValue $Policy 'priority'
    if ($null -ne $priority) {
        $optionalFields += "  priority    = $priority"
    }
    
    # Apps field - list of strings (use resource references if mapping available)
    $appsRaw = Get-PropValue $Policy 'apps'
    if ($null -ne $appsRaw) {
        $apps = @($appsRaw)  # force array - ConvertFrom-Json unwraps single-element JSON arrays to a scalar
        if ($apps.Count -gt 0) {
            $appsFormattedList = @()
            foreach ($app in $apps) {
                if ($AppIdToResourceName.ContainsKey($app)) {
                    $resourceName = $AppIdToResourceName[$app]
                    $appsFormattedList += "spa_application.$resourceName.id"
                }
                else {
                    $appsFormattedList += "`"$app`""
                }
            }
            $appsFormatted = $appsFormattedList -join ', '
            $optionalFields += "  apps = [$appsFormatted]"
        }
        else {
            $optionalFields += "  apps = []"
        }
    }
    else {
        $optionalFields += "  apps = []"
    }
    
    # Conditions field
    $conditions = Get-PropValue $Policy 'conditions'
    if ($null -ne $conditions) {
        if (($conditions -is [hashtable] -and $conditions.Count -gt 0) -or ($conditions -is [PSCustomObject])) {
            $conditionsHcl = ConvertTo-HclDict $conditions
            $optionalFields += "  conditions = $conditionsHcl"
        }
        elseif ($conditions -is [array] -and $conditions.Count -gt 0) {
            $conditionsList = @()
            foreach ($condition in $conditions) {
                if ($condition -is [hashtable] -or $condition -is [PSCustomObject]) {
                    $conditionHcl = ConvertTo-HclDict $condition
                    $conditionsList += "    $conditionHcl"
                }
            }
            if ($conditionsList.Count -gt 0) {
                $conditionsContent = $conditionsList -join ",`n"
                $optionalFields += "  conditions = [`n$conditionsContent`n  ]"
            }
        }
    }
    
    # Actions field
    $actions = Get-PropValue $Policy 'actions'
    if ($null -ne $actions) {
        if (($actions -is [hashtable] -and $actions.Count -gt 0) -or ($actions -is [PSCustomObject])) {
            $actionsHcl = ConvertTo-HclDict $actions
            $optionalFields += "  actions = $actionsHcl"
        }
        elseif ($actions -is [array] -and $actions.Count -gt 0) {
            $actionsList = @()
            foreach ($action in $actions) {
                if ($action -is [hashtable] -or $action -is [PSCustomObject]) {
                    $actionHcl = ConvertTo-HclDict $action
                    $actionsList += "    $actionHcl"
                }
            }
            if ($actionsList.Count -gt 0) {
                $actionsContent = $actionsList -join ",`n"
                $optionalFields += "  actions = [`n$actionsContent`n  ]"
            }
        }
    }
    
    # Access rules field - complex nested structure
    $accessRules = Get-PropValue $Policy 'access_rules'

    # PowerShell JSON parsing returns single items as PSCustomObject instead of array
    # Wrap single objects in array for consistent processing
    if ($null -ne $accessRules -and $accessRules -is [PSCustomObject]) {
        $accessRules = @($accessRules)
    }
    
    if ($null -ne $accessRules -and $accessRules -is [array]) {
        if ($accessRules.Count -gt 0) {
            $accessRulesBlocks = @()
            foreach ($rule in $accessRules) {
                $ruleBlock = New-AccessRuleBlock -Rule $rule
                $accessRulesBlocks += $ruleBlock
            }
            $accessRulesContent = $accessRulesBlocks -join ",`n"
            $optionalFields += "  access_rules = [`n$accessRulesContent`n  ]"
        }
        else {
            $optionalFields += "  access_rules = []"
        }
    }
    
    # Build the resource block
    $optionalPart = ''
    if ($optionalFields.Count -gt 0) {
        $optionalPart = "`n" + ($optionalFields -join "`n")
    }
    
    $resource = Format-ResourceConfig -ResourceType 'access_policy' -Parameters @{
        safe_name = $safeName
        name = $policyName
        optional_fields = $optionalPart
    }
    
    return @($safeName, $resource)
}

function New-SecurityGroupResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_security_group (without read-only fields)
    
    .DESCRIPTION
        This method handles different field types appropriately:
        - Retains empty strings (e.g., type = "")
        - Retains empty lists/maps when explicitly set to []/{} 
        - Excludes null/None values to let provider use defaults
    
    .PARAMETER Group
        The security group data from the API
    
    .PARAMETER AppIdToResourceName
        Optional mapping of application IDs to terraform resource names
        (e.g., @{"app-uuid-123" = "app_salesforce"})
        Used to replace hardcoded IDs with resource references
    #>
    param(
        [object]$Group,
        [hashtable]$AppIdToResourceName = @{}
    )
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = Get-PropValue $Group 'name' ''
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'sg'
    
    # Required fields
    $nameAttr = ConvertTo-EscapedTerraformString (Get-PropValue $Group 'name' '')
    
    # Required app_ids field - use resource references if mapping available
    $appIds = Get-PropValue $Group 'app_ids' @()
    if ($appIds -and $appIds.Count -gt 0) {
        $appIdsList = @()
        foreach ($appId in $appIds) {
            if ($AppIdToResourceName.ContainsKey($appId)) {
                $resourceName = $AppIdToResourceName[$appId]
                $appIdsList += "spa_application.$resourceName.id"
            }
            else {
                $appIdsList += "`"$appId`""
            }
        }
        $appIdsStr = '[' + ($appIdsList -join ', ') + ']'
    }
    else {
        $appIdsStr = '[]'
    }
    
    # Required system configuration
    $system = Get-PropValue $Group 'system' @{}
    $systemDataIn = Get-PropValue $system 'data_in' 'disabled'
    $systemDataOut = Get-PropValue $system 'data_out' 'disabled'
    
    # Required unpublished app configuration
    $unpublishedApp = Get-PropValue $Group 'unpublished_app' @{}
    $unpublishedAppDataIn = Get-PropValue $unpublishedApp 'data_in' 'disabled'
    $unpublishedAppDataOut = Get-PropValue $unpublishedApp 'data_out' 'disabled'
    
    # Build the resource block
    $optionalPart = ''
    
    $resource = Format-ResourceConfig -ResourceType 'security_group' -Parameters @{
        safe_name = $safeName
        name = $nameAttr
        app_ids = $appIdsStr
        system_data_in = $systemDataIn
        system_data_out = $systemDataOut
        unpublished_app_data_in = $unpublishedAppDataIn
        unpublished_app_data_out = $unpublishedAppDataOut
        optional_fields = $optionalPart
    }
    
    return @($safeName, $resource)
}

function New-BrowserModeResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_browser_mode (without read-only fields)
    #>
    param([string]$BrowserMode)
    
    $name = 'browser_mode'
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'browser_mode'
    
    $resource = Format-ResourceConfig -ResourceType 'browser_mode' -Parameters @{
        safe_name = $safeName
        optional_fields = "  browser_mode = `"$BrowserMode`""
    }
    
    return @($safeName, $resource)
}

function New-HybridConfigDataSource {
    <#
    .SYNOPSIS
        Generate terraform data source block for spa_hybrid_config
    #>
    param([object]$HybridConfigData)
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = 'hybrid_config'
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'hybrid_config'
    
    $firstTime = Get-PropValue $HybridConfigData 'first_time' 'unknown'
    $isHybrid = Get-PropValue $HybridConfigData 'is_hybrid' 'unknown'
    
    # Generate data source block (data sources are read-only, no configuration needed)
    $dataSource = @"
data "spa_hybrid_config" "$safeName" {
  # This data source retrieves current hybrid configuration
  # first_time: $firstTime
  # is_hybrid: $isHybrid
}
"@
    
    return @($safeName, $dataSource)
}

function New-CertificateResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_certificate (without read-only fields)
    #>
    param([object]$Certificate)
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = Get-PropValue $Certificate 'certificate_name' $Certificate
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'cert'
    
    # Required fields
    $certName = ConvertTo-EscapedTerraformString (Get-PropValue $Certificate 'certificate_name' '')
    $certData = ConvertTo-EscapedTerraformString (Get-PropValue $Certificate 'certificate_data' '')
    
    $resource = Format-ResourceConfig -ResourceType 'certificate' -Parameters @{
        safe_name = $safeName
        name = $certName
        certificate = $certData
        optional_fields = ''
    }
    
    return @($safeName, $resource)
}

function New-RoutingDomainResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_routing_domain (without read-only fields)
    
    .DESCRIPTION
        This method handles different field types appropriately:
        - Retains empty strings (e.g., comment = "")
        - Retains empty lists when explicitly set to []
        - Excludes null/None values to let provider use defaults
    #>
    param([object]$RoutingDomain)
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $name = Get-PropValue $RoutingDomain 'name' (Get-PropValue $RoutingDomain 'fqdn' 'routing_domain')
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'routing_domain'
    
    # Required fields
    $fqdn = ConvertTo-EscapedTerraformString (Get-PropValue $RoutingDomain 'fqdn' '')
    $type = ConvertTo-EscapedTerraformString (Get-PropValue $RoutingDomain 'type' '')
    
    # Handle optional fields
    $optionalFields = @()
    
    # Optional string fields
    $stringFields = @(
        @('flag', 'flag'),
        @('app_type', 'app_type'),
        @('comment', 'comment')
    )
    foreach ($field in $stringFields) {
        $fieldVal = Get-PropValue $RoutingDomain $field[0]
        if ($null -ne $fieldVal) {
            $tfName = $field[1].PadRight(11)
            $optionalFields += "  $tfName = `"$(ConvertTo-EscapedTerraformString $fieldVal)`""
        }
    }
    
    # Boolean field
    $ipVal = Get-PropValue $RoutingDomain 'ip'
    if ($null -ne $ipVal) {
        $optionalFields += "  ip          = $($ipVal.ToString().ToLower())"
    }
    
    # Required list field - always include (default to empty list if not present or null)
    $locationIds = Get-PropValue $RoutingDomain 'location_ids'
    if ($null -eq $locationIds) {
        $locationIds = @()
    }
    $optionalFields += "  location_ids = $(ConvertTo-TerraformStringList $locationIds)"
    
    # Build the resource block
    $optionalPart = ''
    if ($optionalFields.Count -gt 0) {
        $optionalPart = "`n" + ($optionalFields -join "`n")
    }
    
    $resource = Format-ResourceConfig -ResourceType 'routing_domain' -Parameters @{
        safe_name = $safeName
        fqdn = $fqdn
        type = $type
        optional_fields = $optionalPart
    }
    
    return @($safeName, $resource)
}

function New-SessionPolicyRuleBlock {
    <#
    .SYNOPSIS
        Generate HCL for one rule inside the rule = [...] list of a spa_session_policy resource
    #>
    param([object]$Rule)

    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) { $val = $obj.$propName; if ($null -ne $val) { return $val } }
        elseif ($obj -is [hashtable]) { if ($obj.ContainsKey($propName)) { return $obj[$propName] } }
        return $default
    }

    $ruleAttrs = @()

    $id = Get-PropValue $Rule 'id'
    if ($null -ne $id -and $id -ne '') {
        $ruleAttrs += "      id = `"$(ConvertTo-EscapedTerraformString $id)`""
    }

    $name = Get-PropValue $Rule 'name'
    if ($null -ne $name -and $name -ne '') {
        $ruleAttrs += "      name = `"$(ConvertTo-EscapedTerraformString $name)`""
    }

    $description = Get-PropValue $Rule 'description'
    if ($null -ne $description) {
        $ruleAttrs += "      description = `"$(ConvertTo-EscapedTerraformString $description)`""
    }

    $priority = Get-PropValue $Rule 'priority'
    if ($null -ne $priority) {
        $ruleAttrs += "      priority = $priority"
    }

    $active = Get-PropValue $Rule 'active'
    if ($null -ne $active) {
        $ruleAttrs += "      active = $($active.ToString().ToLower())"
    }

    # Actions block — only emit non-empty fields
    $actions = Get-PropValue $Rule 'actions'
    if ($null -ne $actions) {
        $actionFields = @()
        $routing = Get-PropValue $actions 'routing'
        if ($null -ne $routing -and $routing -ne '') {
            $actionFields += "        routing = `"$(ConvertTo-EscapedTerraformString $routing)`""
        }
        $disableSg = Get-PropValue $actions 'disable_security_groups'
        if ($null -ne $disableSg -and $disableSg -ne '') {
            $actionFields += "        disable_security_groups = `"$(ConvertTo-EscapedTerraformString $disableSg)`""
        }
        $localLan = Get-PropValue $actions 'local_lan_access'
        if ($null -ne $localLan -and $localLan -ne '') {
            $actionFields += "        local_lan_access = `"$(ConvertTo-EscapedTerraformString $localLan)`""
        }
        if ($actionFields.Count -gt 0) {
            $actionsContent = $actionFields -join "`n"
            $ruleAttrs += "      actions = {`n$actionsContent`n      }"
        }
    }

    # Condition blocks
    $conditions = Get-PropValue $Rule 'condition' @()
    if (-not $conditions) { $conditions = @() }
    $conditionsArray = @($conditions)
    if ($conditionsArray.Count -gt 0) {
        $conditionBlocks = @()
        foreach ($cond in $conditionsArray) {
            $condAttrs = @()

            $condType = Get-PropValue $cond 'type' ''
            $condAttrs += "          type = `"$condType`""

            $operator = Get-PropValue $cond 'operator' ''
            $condAttrs += "          operator = `"$operator`""

            $values = Get-PropValue $cond 'values' @()
            $valuesStr = ConvertTo-TerraformStringList @($values)
            $condAttrs += "          values = $valuesStr"

            $tagSource = Get-PropValue $cond 'tag_source'
            if ($null -eq $tagSource) { $tagSource = Get-PropValue $cond 'tagSource' }
            if ($null -ne $tagSource) {
                $condAttrs += "          tag_source = `"$(ConvertTo-EscapedTerraformString $tagSource)`""
            }

            $tagKey = Get-PropValue $cond 'tag_key'
            if ($null -eq $tagKey) { $tagKey = Get-PropValue $cond 'tagKey' }
            if ($null -ne $tagKey) {
                $condAttrs += "          tag_key = `"$(ConvertTo-EscapedTerraformString $tagKey)`""
            }

            $metadata = Get-PropValue $cond 'metadata'
            if ($null -ne $metadata) {
                $hasEntries = ($metadata -is [PSCustomObject] -and ($metadata.PSObject.Properties | Measure-Object).Count -gt 0) -or
                              ($metadata -is [hashtable] -and $metadata.Count -gt 0)
                if ($hasEntries) {
                    $metadataHcl = ConvertTo-HclDict $metadata
                    $condAttrs += "          metadata = $metadataHcl"
                }
            }

            $condContent = $condAttrs -join "`n"
            $conditionBlocks += "        {`n$condContent`n        }"
        }
        $conditionsContent = $conditionBlocks -join ",`n"
        $ruleAttrs += "      condition = [`n$conditionsContent`n      ]"
    }

    $ruleContent = $ruleAttrs -join "`n"
    return "    {`n$ruleContent`n    }"
}

function New-SessionPolicyResource {
    <#
    .SYNOPSIS
        Generate terraform resource block for spa_session_policy (without read-only fields)
    #>
    param([object]$Policy)

    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) { $val = $obj.$propName; if ($null -ne $val) { return $val } }
        elseif ($obj -is [hashtable]) { if ($obj.ContainsKey($propName)) { return $obj[$propName] } }
        return $default
    }

    $name = Get-PropValue $Policy 'name' ''
    $safeName = Get-SafeTerraformName -Name $name -Prefix 'sp'

    $policyName = ConvertTo-EscapedTerraformString $name
    $active = Get-PropValue $Policy 'active' $false

    $lines = @()
    $lines += "resource `"spa_session_policy`" `"$safeName`" {"
    $lines += "  name   = `"$policyName`""
    $lines += "  active = $($active.ToString().ToLower())"

    $description = Get-PropValue $Policy 'description'
    if ($null -ne $description) {
        $lines += "  description = `"$(ConvertTo-EscapedTerraformString $description)`""
    }

    $priority = Get-PropValue $Policy 'priority'
    if ($null -ne $priority -and [int]$priority -ne 0) {
        $lines += "  priority = $priority"
    }

    # Rules (Required field) — data source output uses 'generic_rules'
    $rules = Get-PropValue $Policy 'generic_rules' @()
    $rulesArray = @($rules)

    if ($rulesArray.Count -gt 0) {
        $ruleBlocks = @()
        foreach ($rule in $rulesArray) {
            $ruleBlocks += New-SessionPolicyRuleBlock -Rule $rule
        }
        $rulesContent = $ruleBlocks -join ",`n"
        $lines += "  generic_rules = [`n$rulesContent`n  ]"
    }
    else {
        $lines += "  generic_rules = []"
    }

    $lines += "}"
    $resource = ($lines -join "`n") + "`n"
    return @($safeName, $resource)
}

function Get-ResourceData {
    <#
    .SYNOPSIS
        Get resource data from memory (replaces the old JSON file approach)
    #>
    param([string]$DataKey)
    
    $data = $script:ResourceData[$DataKey]
    if (-not $data -or ($data -is [hashtable] -and $data.ContainsKey('error'))) {
        if ($script:VerboseMode) {
            Write-WarningMessage "No valid data available for $DataKey"
        }
        return $null
    }
    return $data
}

function Get-MemoryDataSummary {
    <#
    .SYNOPSIS
        Get summary of data stored in memory (for debugging/validation)
    #>
    
    $summary = @{}
    foreach ($key in $script:ResourceData.Keys) {
        $data = $script:ResourceData[$key]
        if ($data -and $data -is [hashtable]) {
            # Count items in the data based on expected structure
            switch ($key) {
                'applications' {
                    if ($data.ContainsKey('applications')) {
                        $summary[$key] = @($data['applications']).Count
                    } else { $summary[$key] = 0 }
                }
                'access_policies' {
                    if ($data.ContainsKey('access_policies')) {
                        $summary[$key] = @($data['access_policies']).Count
                    } else { $summary[$key] = 0 }
                }
                'security_groups' {
                    if ($data.ContainsKey('security_groups')) {
                        $summary[$key] = @($data['security_groups']).Count
                    } else { $summary[$key] = 0 }
                }
                'routing_domains' {
                    if ($data.ContainsKey('routing_domains')) {
                        $summary[$key] = @($data['routing_domains']).Count
                    } else { $summary[$key] = 0 }
                }
                'certificates' {
                    if ($data.ContainsKey('certificates')) {
                        $summary[$key] = @($data['certificates']).Count
                    } else { $summary[$key] = 0 }
                }
                'terminate_machine_access' {
                    if ($data.ContainsKey('machines')) {
                        $summary[$key] = @($data['machines']).Count
                    } else { $summary[$key] = 0 }
                }
                'terminate_user_access' {
                    if ($data.ContainsKey('users')) {
                        $summary[$key] = @($data['users']).Count
                    } else { $summary[$key] = 0 }
                }
                'session_policies' {
                    if ($data.ContainsKey('session_policies')) {
                        $summary[$key] = @($data['session_policies']).Count
                    } else { $summary[$key] = 0 }
                }
                { $_ -in @('browser_mode', 'hybrid_config') } {
                    $summary[$key] = if ($data) { 1 } else { 0 }
                }
                default {
                    $summary[$key] = if ($data) { $data.Count } else { 0 }
                }
            }
        }
        else {
            $summary[$key] = 0
        }
    }
    return $summary
}

function New-TerraformResources {
    <#
    .SYNOPSIS
        Generate complete terraform configuration and import blocks
    #>
    
    Write-Header "Generating Terraform Configuration"
    
    # Reset used names
    $script:UsedNames.Clear()
    
    $resources = @()
    $resources += "# Complete spa_resources.tf with all required attributes"
    $resources += "# Generated by SPA Manager"
    $resources += "# Generated on $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    $resources += ""
    
    $importCommands = @()
    
    # Build mapping of application IDs to terraform resource names
    $appIdToResourceName = @{}
    
    # Generate routing domains and build FQDN->resource name mapping
    $routingDomainFqdnToResource = @{}
    
    # Helper function to get property value
    function Get-PropValue($obj, $propName, $default = $null) {
        if ($obj -is [PSCustomObject]) {
            $val = $obj.$propName
            if ($null -ne $val) { return $val }
        }
        elseif ($obj -is [hashtable]) {
            if ($obj.ContainsKey($propName)) { return $obj[$propName] }
        }
        return $default
    }
    
    $rdData = Get-ResourceData 'routing_domains'
    if ($rdData -and $rdData.ContainsKey('routing_domains')) {
        $resources += "# Routing Domains"
        foreach ($domain in $rdData['routing_domains']) {
            $result = New-RoutingDomainResource -RoutingDomain $domain
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Build the FQDN mapping
            $fqdn = Get-PropValue $domain 'fqdn' ''
            if ($fqdn) {
                $routingDomainFqdnToResource[$fqdn] = $safeName
            }
            
            # Add import command
            $domainId = Get-PropValue $domain 'fqdn' (Get-PropValue $domain 'id' '')
            if ($domainId) {
                $importCommands += "terraform import spa_routing_domain.$safeName `"$domainId`""
            }
            
            $script:ResourceCounts['routing_domains']++
        }
    }
    
    # Generate applications
    $appsData = Get-ResourceData 'applications'
    if ($appsData -and $appsData.ContainsKey('applications')) {
        $resources += "# Applications"
        foreach ($app in $appsData['applications']) {
            $result = New-ApplicationResource -App $app -RoutingDomainFqdnToResource $routingDomainFqdnToResource
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Add import command
            $appId = Get-PropValue $app 'id' (Get-PropValue $app 'name' '')
            if ($appId) {
                $importCommands += "terraform import spa_application.$safeName `"$appId`""
                $appIdToResourceName[$appId] = $safeName
            }
            
            $script:ResourceCounts['applications']++
        }
    }
    
    # Generate access policies
    $apData = Get-ResourceData 'access_policies'
    if ($apData -and $apData.ContainsKey('access_policies')) {
        $resources += "# Access Policies"
        foreach ($policy in $apData['access_policies']) {
            $result = New-AccessPolicyResource -Policy $policy -AppIdToResourceName $appIdToResourceName
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Add import command
            $policyId = Get-PropValue $policy 'id' (Get-PropValue $policy 'name' '')
            if ($policyId) {
                $importCommands += "terraform import spa_access_policy.$safeName `"$policyId`""
            }
            
            $script:ResourceCounts['access_policies']++
        }
    }
    
    # Generate session policies
    $spData = Get-ResourceData 'session_policies'
    if ($spData -and $spData.ContainsKey('session_policies')) {
        $resources += "# Session Policies"
        foreach ($policy in $spData['session_policies']) {
            $result = New-SessionPolicyResource -Policy $policy
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Add import command
            $policyId = Get-PropValue $policy 'id' (Get-PropValue $policy 'name' '')
            if ($policyId) {
                $importCommands += "terraform import spa_session_policy.$safeName `"$policyId`""
            }
            
            $script:ResourceCounts['session_policies']++
        }
    }
    
    # Generate security groups
    $sgData = Get-ResourceData 'security_groups'
    if ($sgData -and $sgData.ContainsKey('security_groups')) {
        $resources += "# Security Groups"
        foreach ($group in $sgData['security_groups']) {
            $result = New-SecurityGroupResource -Group $group -AppIdToResourceName $appIdToResourceName
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Add import command
            $groupId = Get-PropValue $group 'id' (Get-PropValue $group 'name' '')
            if ($groupId) {
                $importCommands += "terraform import spa_security_group.$safeName `"$groupId`""
            }
            
            $script:ResourceCounts['security_groups']++
        }
    }
    
    # Generate certificates
    $certData = Get-ResourceData 'certificates'
    if ($certData -and $certData.ContainsKey('certificates')) {
        $resources += "# Certificates"
        foreach ($cert in $certData['certificates']) {
            $result = New-CertificateResource -Certificate $cert
            $safeName = $result[0]
            $resource = $result[1]
            $resources += $resource
            $resources += ""
            
            # Add import command
            $certId = Get-PropValue $cert 'certificate_id' ''
            if ($certId) {
                $importCommands += "terraform import spa_certificate.$safeName `"$certId`""
            }
            
            $script:ResourceCounts['certificates']++
        }
    }
    
    # Generate browser modes
    $browserModeData = Get-ResourceData 'browser_mode'
    if ($browserModeData -and $browserModeData.ContainsKey('browser_mode')) {
        $resources += "# Browser Modes"
        $browserMode = $browserModeData['browser_mode']
        $result = New-BrowserModeResource -BrowserMode $browserMode
        $safeName = $result[0]
        $resource = $result[1]
        $resources += $resource
        $resources += ""
        
        # Add import command
        $importCommands += "terraform import spa_browser_mode.$safeName `"browser_mode`""
        
        $script:ResourceCounts['browser_modes']++
    }
    
    # Generate hybrid config data source
    $hybridConfigData = Get-ResourceData 'hybrid_config'
    if ($hybridConfigData) {
        $resources += "# Hybrid Configuration (Data Source)"
        $result = New-HybridConfigDataSource -HybridConfigData $hybridConfigData
        $safeName = $result[0]
        $dataSource = $result[1]
        $resources += $dataSource
        $resources += ""
        
        $script:ResourceCounts['hybrid_configs']++
    }
    
    # Generate terminate machine access
    $tmaData = Get-ResourceData 'terminate_machine_access'
    if ($tmaData) {
        if ($script:DebugMode) {
            Write-Status "Terminate machine access data structure: $($tmaData.Keys -join ', ')"
        }
        
        if ($tmaData.ContainsKey('machines') -and $tmaData['machines']) {
            $resources += "# Terminate Machine Access Resources"
            
            $i = 0
            foreach ($access in $tmaData['machines']) {
                $machineName = Get-PropValue $access 'name' (Get-PropValue $access 'account_name' "machine_$i")
                $machineId = Get-PropValue $access 'id' ''
                $safeName = Get-SafeTerraformName -Name "tma_$machineName" -Prefix 'tma'
                
                $resourceConfig = Format-ResourceConfig -ResourceType 'terminate_machine_access' -Parameters @{
                    safe_name = $safeName
                    account_name = Get-PropValue $access 'account_name' ''
                    name = Get-PropValue $access 'name' ''
                    dns_host_name = Get-PropValue $access 'dns_host_name' ''
                    domain_name = Get-PropValue $access 'domain_name' ''
                    object_id = Get-PropValue $access 'object_id' ''
                    idp_type = Get-PropValue $access 'idp_type' ''
                    duration = Get-PropValue $access 'duration' 0
                }
                
                $resources += $resourceConfig
                $resources += ""
                
                if ($machineId) {
                    $importCommands += "terraform import spa_terminate_machine_access.$safeName `"$machineId`""
                }
                
                $script:ResourceCounts['terminate_machine_access']++
                $i++
            }
        }
    }
    
    # Generate terminate user access
    $tuaData = Get-ResourceData 'terminate_user_access'
    if ($tuaData) {
        if ($script:DebugMode) {
            Write-Status "Terminate user access data structure: $($tuaData.Keys -join ', ')"
        }
        
        if ($tuaData.ContainsKey('users') -and $tuaData['users']) {
            $resources += "# Terminate User Access Resources"
            
            $i = 0
            foreach ($access in $tuaData['users']) {
                $userEmail = Get-PropValue $access 'email' (Get-PropValue $access 'account_name' "user_$i")
                $userId = Get-PropValue $access 'id' ''
                $safeName = Get-SafeTerraformName -Name "tua_$userEmail" -Prefix 'tua'
                
                $resourceConfig = Format-ResourceConfig -ResourceType 'terminate_user_access' -Parameters @{
                    safe_name = $safeName
                    account_name = Get-PropValue $access 'account_name' ''
                    email = Get-PropValue $access 'email' ''
                    domain_name = Get-PropValue $access 'domain_name' ''
                    object_id = Get-PropValue $access 'object_id' ''
                    idp_type = Get-PropValue $access 'idp_type' ''
                    duration = Get-PropValue $access 'duration' 0
                }
                
                $resources += $resourceConfig
                $resources += ""
                
                if ($userId) {
                    $importCommands += "terraform import spa_terminate_user_access.$safeName `"$userId`""
                }
                
                $script:ResourceCounts['terminate_user_access']++
                $i++
            }
        }
    }
    
    # Write spa_resources.tf
    try {
        $spaResourcesPath = Join-Path $script:WorkDir 'spa_resources.tf'
        Set-Content -Path $spaResourcesPath -Value ($resources -join "`n") -Encoding UTF8
        Write-SuccessMessage "Generated spa_resources.tf"
        
        # Format the file
        Push-Location $script:WorkDir
        try {
            $fmtResult = & terraform fmt $spaResourcesPath 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-SuccessMessage "Terraform fmt completed successfully!"
            }
            else {
                Write-ErrorMessage "Terraform fmt failed: $fmtResult"
            }
        }
        finally {
            Pop-Location
        }
    }
    catch {
        Write-ErrorMessage "Error writing spa_resources.tf: $_"
        return $false
    }
    
    # Generate Terraform import blocks
    if ($importCommands.Count -gt 0) {
        New-TerraformImportBlocks -ImportCommands $importCommands
    }
    else {
        New-MinimalImportBlocks
    }
    
    return $true
}

function New-TerraformImportBlocks {
    <#
    .SYNOPSIS
        Generate Terraform import blocks instead of Python import script
    #>
    param([string[]]$ImportCommands)
    
    try {
        # Generate the import blocks
        $importBlocks = Format-ImportBlocks -ImportCommands $ImportCommands
        
        $timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
        $importContent = $script:IMPORT_BLOCKS_HEADER -f $timestamp, $importBlocks
        
        $importPath = Join-Path $script:WorkDir 'imports.tf'
        Set-Content -Path $importPath -Value $importContent -Encoding UTF8
        
        Write-SuccessMessage "Generated imports.tf (Terraform import blocks)"
        return $true
    }
    catch {
        Write-ErrorMessage "Error writing import blocks: $_"
        return $false
    }
}

function Format-ImportBlocks {
    <#
    .SYNOPSIS
        Format import commands as Terraform import blocks
    #>
    param([string[]]$ImportCommands)
    
    $importBlocks = @()
    foreach ($cmd in $ImportCommands) {
        # Parse "terraform import resource_type.resource_name 'resource_id'"
        $parts = $cmd -split ' ', 4
        if ($parts.Count -ge 4) {
            $resourceAddress = $parts[2]  # e.g., "spa_application.app_name"
            $resourceId = $parts[3].Trim('"''')
            
            # Generate import block
            $importBlock = $script:IMPORT_BLOCK -f $resourceAddress, $resourceId
            $importBlocks += $importBlock
        }
    }
    
    return ($importBlocks -join "`n`n")
}

function New-MinimalImportBlocks {
    <#
    .SYNOPSIS
        Generate a minimal import file when only data sources are found
    #>
    
    try {
        $tmaCount = $script:ResourceCounts['terminate_machine_access']
        $tuaCount = $script:ResourceCounts['terminate_user_access']
        
        # Build data sources list
        $dataSources = ""
        if ($tmaCount -gt 0) {
            $dataSources += "`n#   - spa_terminate_machine_access.discovered ($tmaCount machines)"
        }
        if ($tuaCount -gt 0) {
            $dataSources += "`n#   - spa_terminate_user_access.discovered ($tuaCount users)"
        }
        
        $timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
        $importContent = $script:MINIMAL_IMPORT_BLOCKS_HEADER -f $timestamp, $dataSources
        
        $importPath = Join-Path $script:WorkDir 'imports.tf'
        Set-Content -Path $importPath -Value $importContent -Encoding UTF8
        
        Write-SuccessMessage "Generated imports.tf (data sources info)"
        return $true
    }
    catch {
        Write-ErrorMessage "Failed to generate import file: $_"
        return $false
    }
}

function Get-SummaryReport {
    <#
    .SYNOPSIS
        Generate summary report
    #>
    
    $totalResources = 0
    foreach ($key in $script:ResourceCounts.Keys) {
        $totalResources += $script:ResourceCounts[$key]
    }
    
    # Only exclude terminate user access from import count since it's data source only
    $importableResources = $totalResources
    
    $timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
    
    $report = $script:SUMMARY_REPORT -f @(
        $timestamp,
        $script:ResourceCounts['applications'],
        $script:ResourceCounts['security_groups'],
        $script:ResourceCounts['access_policies'],
        $script:ResourceCounts['certificates'],
        $script:ResourceCounts['browser_modes'],
        $script:ResourceCounts['routing_domains'],
        $script:ResourceCounts['hybrid_configs'],
        $script:ResourceCounts['terminate_machine_access'],
        $script:ResourceCounts['terminate_user_access'],
        $script:ResourceCounts['session_policies'],
        $totalResources,
        $importableResources,
        $script:UsedNames.Count
    )
    
    return $report
}

function Remove-TempFiles {
    <#
    .SYNOPSIS
        Clean up temporary files and generated configurations
    #>
    
    Write-Header "Cleaning Up Temporary Files and Generated Configurations"
    
    # Temporary file patterns to remove
    $tempPatterns = @(
        'temp_*.tf',
        'terraform.tfstate*',
        '.terraform.lock.hcl',
        '.terraform.tfstate.lock.info',
        'tfplan',
        'test_plan',
        'debug*.log'
    )
    
    # Generated files to remove
    $generatedFiles = @(
        'management_summary.md',
        'provider.tf',
        'terraform-config-out',
        'spa_resources.tf',
        'imports.tf'
    )
    
    $removedCount = 0
    
    # Remove temporary files by pattern
    foreach ($pattern in $tempPatterns) {
        $matchingFiles = Get-ChildItem -Path $script:WorkDir -Filter $pattern -ErrorAction SilentlyContinue
        foreach ($file in $matchingFiles) {
            try {
                Remove-Item -Path $file.FullName -Force -ErrorAction Stop
                $removedCount++
                Write-Status "Removed temporary file: $($file.Name)"
            }
            catch {
                Write-WarningMessage "Could not remove $($file.FullName): $_"
            }
        }
    }
    
    # Remove generated files
    foreach ($filename in $generatedFiles) {
        $filePath = Join-Path $script:WorkDir $filename
        if (Test-Path $filePath) {
            try {
                Remove-Item -Path $filePath -Force -ErrorAction Stop
                $removedCount++
                Write-Status "Removed generated file: $filename"
            }
            catch {
                Write-WarningMessage "Could not remove $filename`: $_"
            }
        }
    }
    
    # Remove .terraform directory
    $terraformDir = Join-Path $script:WorkDir '.terraform'
    if (Test-Path $terraformDir) {
        try {
            Remove-Item -Path $terraformDir -Recurse -Force -ErrorAction Stop
            $removedCount++
            Write-Status "Removed .terraform directory"
        }
        catch {
            Write-WarningMessage "Could not remove .terraform directory: $_"
        }
    }
    
    if ($removedCount -gt 0) {
        Write-SuccessMessage "Successfully removed $removedCount files/directories"
    }
    else {
        Write-Status "No files to clean up"
    }
}

# =============================================================================
# TEST FUNCTIONS
# =============================================================================

function Invoke-Tests {
    <#
    .SYNOPSIS
        Run basic functionality tests
    #>
    
    Write-Header "Running Basic Functionality Tests"
    
    $tests = @(
        @{ Name = 'PowerShell version'; Test = { $PSVersionTable.PSVersion.Major -ge 7 } },
        @{ Name = 'Work directory exists'; Test = { Test-Path $script:WorkDir } },
        @{ Name = 'Can create temp file'; Test = { Test-FileCreation } },
        @{ Name = 'JSON parsing'; Test = { Test-JsonParsing } },
        @{ Name = 'Terraform name sanitization'; Test = { Test-NameSanitization } }
    )
    
    $passed = 0
    foreach ($testCase in $tests) {
        try {
            $result = & $testCase.Test
            if ($result) {
                Write-Status "✓ $($testCase.Name)"
                $passed++
            }
            else {
                Write-ErrorMessage "✗ $($testCase.Name)"
            }
        }
        catch {
            Write-ErrorMessage "✗ $($testCase.Name): $_"
        }
    }
    
    $success = $passed -eq $tests.Count
    if ($success) {
        Write-SuccessMessage "All $($tests.Count) tests passed!"
    }
    else {
        Write-ErrorMessage "Only $passed/$($tests.Count) tests passed"
    }
    
    return $success
}

function Test-FileCreation {
    <#
    .SYNOPSIS
        Test file creation capability
    #>
    
    $testFile = Join-Path $script:WorkDir 'test_temp_file.tmp'
    try {
        Set-Content -Path $testFile -Value 'test' -Encoding UTF8
        Remove-Item -Path $testFile -Force
        return $true
    }
    catch {
        return $false
    }
}

function Test-JsonParsing {
    <#
    .SYNOPSIS
        Test JSON parsing capability
    #>
    
    try {
        $testData = '{"test": "value"}'
        $parsed = $testData | ConvertFrom-Json -Depth 100
        return ($parsed.test -eq 'value')
    }
    catch {
        return $false
    }
}

function Test-NameSanitization {
    <#
    .SYNOPSIS
        Test terraform name sanitization
    #>
    
    try {
        # Reset used names for test
        $savedNames = $script:UsedNames.Clone()
        $script:UsedNames = @{}
        
        $testNames = @('test app', '123app', 'app-with-dashes', 'app.with.dots')
        foreach ($name in $testNames) {
            $safeName = Get-SafeTerraformName -Name $name
            # Check if name matches valid terraform identifier pattern
            if ($safeName -notmatch '^[a-zA-Z_][a-zA-Z0-9_]*$') {
                # Restore used names
                $script:UsedNames = $savedNames
                return $false
            }
        }
        
        # Restore used names
        $script:UsedNames = $savedNames
        return $true
    }
    catch {
        return $false
    }
}

# =============================================================================
# TERRAFORM OPERATION FUNCTIONS
# =============================================================================

function Invoke-TerraformPlan {
    <#
    .SYNOPSIS
        Run terraform plan
    #>
    
    Write-Header "Running Terraform Plan"
    
    Push-Location $script:WorkDir
    try {
        # Set environment variables for parallelism
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        if ($script:DebugMode) {
            $env:TF_LOG = $script:DebugLevel
            $env:TF_LOG_PROVIDER = $script:DebugLevel
            $env:TF_LOG_PATH = Join-Path $script:WorkDir "debug-terraform-plan.log"
        }
        
        # Build plan arguments with -out and -generate-config-out like Python
        $planArgs = @('plan', '-out', $script:PlanOutput, '-generate-config-out', 'terraform-config-out')
        
        $process = Start-Process -FilePath 'terraform' -ArgumentList $planArgs -Wait -NoNewWindow -PassThru -RedirectStandardOutput 'plan_out.txt' -RedirectStandardError 'plan_err.txt'
        
        $stdout = Get-Content 'plan_out.txt' -Raw -ErrorAction SilentlyContinue
        $stderr = Get-Content 'plan_err.txt' -Raw -ErrorAction SilentlyContinue
        
        Remove-Item 'plan_out.txt', 'plan_err.txt' -Force -ErrorAction SilentlyContinue
        
        if ($stdout) {
            Write-Host $stdout
        }
        
        if ($process.ExitCode -ne 0) {
            if ($stderr) {
                Write-ErrorMessage $stderr
            }
            return $false
        }
        
        Write-SuccessMessage "Terraform plan completed successfully!"
        return $true
    }
    catch {
        Write-ErrorMessage "Error running terraform plan: $_"
        return $false
    }
    finally {
        Pop-Location
    }
}

function Invoke-TerraformPlanResources {
    <#
    .SYNOPSIS
        Plan terraform resource for specific addresses
    #>
    param([string[]]$Addresses)
    
    Write-Header "Terraform Plan Resource"
    
    Push-Location $script:WorkDir
    try {
        # Set environment variables for parallelism
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        if ($script:DebugMode) {
            $env:TF_LOG_PROVIDER = $script:DebugLevel
            $env:TF_LOG_PATH = Join-Path $script:WorkDir "debug-terraform-plan.log"
        }
        
        # Build command arguments
        # Use platform-appropriate null device: NUL on Windows, /dev/null on Unix
        $nullDevice = if ($IsWindows -or ([System.Environment]::OSVersion.Platform -eq 'Win32NT')) { 'NUL' } else { '/dev/null' }
        $planArgs = @('plan', "-out=$nullDevice")
        foreach ($addr in $Addresses) {
            $planArgs += '-target'
            $planArgs += $addr
        }
        
        $process = Start-Process -FilePath 'terraform' -ArgumentList $planArgs -Wait -NoNewWindow -PassThru -RedirectStandardOutput 'plan_out.txt' -RedirectStandardError 'plan_err.txt'
        
        $stdout = Get-Content 'plan_out.txt' -Raw -ErrorAction SilentlyContinue
        $stderr = Get-Content 'plan_err.txt' -Raw -ErrorAction SilentlyContinue
        
        Remove-Item 'plan_out.txt', 'plan_err.txt' -Force -ErrorAction SilentlyContinue
        
        if ($stdout) {
            Write-Host $stdout
        }
        
        if ($process.ExitCode -ne 0) {
            Write-ErrorMessage "Terraform plan failed:"
            if ($stderr) {
                Write-ErrorMessage $stderr
            }
            return $false
        }
        
        Write-SuccessMessage "Terraform plan for $($Addresses -join ', ') completed successfully!"
        return $true
    }
    catch {
        Write-ErrorMessage "Error running terraform plan: $_"
        return $false
    }
    finally {
        Pop-Location
    }
}

function Show-TerraformUpdate {
    <#
    .SYNOPSIS
        Show terraform plan update
    #>
    param([bool]$Detail)
    
    Write-Header "Showing Terraform Plan Updates"
    
    # Check if plan output exists, create it if not
    if (-not $script:PlanOutput -or -not (Test-Path $script:PlanOutput)) {
        if (-not (Invoke-TerraformPlan)) {
            Write-ErrorMessage "Failed to create plan output file"
            return $false
        }
    }
    
    Push-Location $script:WorkDir
    try {
        # Set environment variables for parallelism
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        $process = Start-Process -FilePath 'terraform' -ArgumentList 'show', '-json', $script:PlanOutput -Wait -NoNewWindow -PassThru -RedirectStandardOutput 'show_out.txt' -RedirectStandardError 'show_err.txt'
        
        $stdout = Get-Content 'show_out.txt' -Raw -ErrorAction SilentlyContinue
        $stderr = Get-Content 'show_err.txt' -Raw -ErrorAction SilentlyContinue
        
        Remove-Item 'show_out.txt', 'show_err.txt' -Force -ErrorAction SilentlyContinue
        
        if ($process.ExitCode -ne 0) {
            Write-ErrorMessage "Terraform show failed:"
            if ($stderr) {
                Write-ErrorMessage $stderr
            }
            return $false
        }
        
        # Parse JSON and categorize changes like Python
        $created = @()
        $updated = @()
        $deleted = @()
        $noOps = @()
        
        try {
            $data = $stdout | ConvertFrom-Json -Depth 100
            
            if (-not $data.resource_changes) {
                return $false
            }
            
            foreach ($change in $data.resource_changes) {
                if ($change.change -and $change.change.actions) {
                    $actions = $change.change.actions
                    if ($actions -contains 'create') {
                        $created += $change.address
                    }
                    elseif ($actions -contains 'update') {
                        $updated += $change.address
                    }
                    elseif ($actions -contains 'delete') {
                        $deleted += $change.address
                    }
                    elseif ($actions -contains 'no-op') {
                        $noOps += $change.address
                    }
                }
            }
            
            # Display categorized changes
            foreach ($r in $created) {
                Write-Status "Resource to be created: $r"
            }
            foreach ($r in $deleted) {
                Write-Status "Resource to be deleted: $r"
            }
            
            if ($Detail) {
                # Show detailed plan for updated resources
                Invoke-TerraformPlanResources -Addresses $updated
            }
            else {
                foreach ($r in $updated) {
                    Write-Status "Resource to be updated: $r"
                }
            }
            
            return $true
        }
        catch {
            Write-ErrorMessage "Failed to parse terraform show output: $_"
            return $false
        }
    }
    catch {
        Write-ErrorMessage "Error showing terraform state: $_"
        return $false
    }
    finally {
        Pop-Location
    }
}

function Invoke-TerraformApplyPlan {
    <#
    .SYNOPSIS
        Apply terraform resource
    #>
    
    Write-Header "Terraform Apply Resource"
    
    # Check if plan output exists, create it if not
    if (-not $script:PlanOutput -or -not (Test-Path $script:PlanOutput)) {
        if (-not (Invoke-TerraformPlan)) {
            Write-ErrorMessage "Failed to create plan output file"
            return $false
        }
    }
    
    Push-Location $script:WorkDir
    try {
        # Set environment variables for parallelism
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        if ($script:DebugMode) {
            $env:TF_LOG_PROVIDER = $script:DebugLevel
            $env:TF_LOG_PATH = Join-Path $script:WorkDir "debug-terraform-apply.log"
        }
        
        $process = Start-Process -FilePath 'terraform' -ArgumentList 'apply', '-auto-approve', '-compact-warnings', $script:PlanOutput -Wait -NoNewWindow -PassThru -RedirectStandardOutput 'apply_out.txt' -RedirectStandardError 'apply_err.txt'
        
        $stdout = Get-Content 'apply_out.txt' -Raw -ErrorAction SilentlyContinue
        $stderr = Get-Content 'apply_err.txt' -Raw -ErrorAction SilentlyContinue
        
        Remove-Item 'apply_out.txt', 'apply_err.txt' -Force -ErrorAction SilentlyContinue
        
        if ($stdout) {
            Write-Host $stdout
        }
        
        if ($process.ExitCode -ne 0) {
            if ($stderr) {
                Write-ErrorMessage $stderr
            }
            return $false
        }
        
        Write-SuccessMessage "Terraform apply completed successfully!"
        
        # Clean up plan output file after successful apply
        if ($script:PlanOutput -and (Test-Path $script:PlanOutput)) {
            Remove-Item $script:PlanOutput -Force -ErrorAction SilentlyContinue
        }
        
        return $true
    }
    catch {
        Write-ErrorMessage "Error running terraform apply: $_"
        return $false
    }
    finally {
        Pop-Location
    }
}

function Test-TerraformConfiguration {
    <#
    .SYNOPSIS
        Validate generated terraform configuration
    #>
    
    Write-Header "Validating Terraform Configuration"
    
    Push-Location $script:WorkDir
    try {
        # Set environment variables for parallelism
        $env:TF_CLI_ARGS_apply = "-parallelism=$script:TERRAFORM_PARALLELISM"
        $env:TF_CLI_ARGS_plan = "-parallelism=$script:TERRAFORM_PARALLELISM"
        
        $terraformrcPath = Join-Path $script:WorkDir '.terraformrc'
        if (Test-Path $terraformrcPath) {
            $env:TF_CLI_CONFIG_FILE = $terraformrcPath
        }
        
        # Run terraform validate
        $process = Start-Process -FilePath 'terraform' -ArgumentList 'validate' -Wait -NoNewWindow -PassThru -RedirectStandardOutput 'validate_out.txt' -RedirectStandardError 'validate_err.txt'
        
        $stdout = Get-Content 'validate_out.txt' -Raw -ErrorAction SilentlyContinue
        $stderr = Get-Content 'validate_err.txt' -Raw -ErrorAction SilentlyContinue
        
        Remove-Item 'validate_out.txt', 'validate_err.txt' -Force -ErrorAction SilentlyContinue
        
        if ($process.ExitCode -eq 0) {
            Write-SuccessMessage "Terraform configuration is valid!"
            return $true
        }
        else {
            Write-ErrorMessage "Terraform validation failed:"
            if ($stderr) {
                Write-ErrorMessage $stderr
            }
            return $false
        }
    }
    catch {
        Write-ErrorMessage "Error running terraform validate: $_"
        return $false
    }
    finally {
        Pop-Location
    }
}

function Initialize-Credentials {
    <#
    .SYNOPSIS
        Setup credentials file
    #>
    
    Write-Header "Setting Up Credentials"
    
    $tfvarsPath = Join-Path $script:WorkDir 'terraform.tfvars'
    $examplePath = Join-Path $script:WorkDir 'terraform.tfvars.example'
    
    if (Test-Path $tfvarsPath) {
        Write-WarningMessage "terraform.tfvars already exists"
        return $true
    }
    
    if (-not (Test-Path $examplePath)) {
        Write-ErrorMessage "terraform.tfvars.example not found"
        return $false
    }
    
    try {
        Copy-Item -Path $examplePath -Destination $tfvarsPath -Force
        Write-SuccessMessage "Created $tfvarsPath"
        Write-Status "Please edit this file with your actual credentials"
        return $true
    }
    catch {
        Write-ErrorMessage "Error creating credentials file: $_"
        return $false
    }
}

function Initialize-TerraformSetup {
    <#
    .SYNOPSIS
        Complete terraform setup workflow
    #>
    
    # Check dependencies
    if (-not (Test-Dependencies)) {
        return $false
    }
    
    # Check credentials
    if (-not (Test-Credentials)) {
        return $false
    }
    
    # Create provider config
    if (-not (New-ProviderConfig)) {
        return $false
    }
    
    # Initialize terraform
    if (-not (Initialize-Terraform)) {
        return $false
    }
    
    Write-SuccessMessage "Terraform setup completed successfully!"
    return $true
}

function Invoke-TerraformList {
    <#
    .SYNOPSIS
        Main discovery and generation workflow
    #>
    param([string[]]$Ids = @())
    
    Write-Header "SPA Resource Discovery and Management"
    
    # Show enhancement mode
    if ($script:QueryIndividualDetails) {
        Write-Status "Enhanced mode: Will query individual items for complete field data"
    }
    else {
        Write-Status "Quick mode: Using list data only (use without -q for enhanced mode)"
    }
    
    if (-not (Initialize-TerraformSetup)) {
        Write-ErrorMessage "Terraform setup failed, aborting workflow"
        return $false
    }
    
    # Enable individual queries if IDs are provided
    if ($Ids.Count -gt 0) {
        $script:QueryIndividualDetails = $true
    }
    
    # Discover resources
    if (-not (Find-Resources -Ids $Ids)) {
        Write-WarningMessage "Resource discovery had issues, but continuing with generation..."
    }
    
    # Show in-memory data summary if in debug mode
    if ($script:DebugMode) {
        $memorySummary = Get-MemoryDataSummary
        Write-Status "In-memory data summary: $($memorySummary | ConvertTo-Json -Compress)"
    }
    
    # Generate terraform configuration
    if (-not (New-TerraformResources)) {
        return $false
    }
    
    # Generate summary
    $summary = Get-SummaryReport
    Write-Host $summary
    
    try {
        $summaryPath = Join-Path $script:WorkDir 'management_summary.md'
        Set-Content -Path $summaryPath -Value $summary -Encoding UTF8
        Write-SuccessMessage "Generated management_summary.md"
    }
    catch {
        Write-WarningMessage "Could not write summary file: $_"
    }
    
    Write-SuccessMessage "SPA resource management completed successfully!"
    Write-Status "Next steps:"
    Write-Status "  1. Review spa_resources.tf"
    Write-Status "  2. Run 'terraform validate' to verify configuration"
    Write-Status "  3. Run 'terraform plan' if needed"
    Write-Status "  4. Use 'terraform apply' to import resources and review changes"
    
    return $true
}

# =============================================================================
# INITIALIZATION FUNCTION
# =============================================================================

function Initialize-SPAManager {
    <#
    .SYNOPSIS
        Initialize the SPA Manager with provided parameters
    
    .DESCRIPTION
        This function initializes the SPA Manager class equivalent.
        It sets up the working directory, feature flags, and in-memory data structures.
        
        This script has been updated to use in-memory data structures instead of temporary JSON files
        to avoid character encoding issues (especially with newlines and special characters).
        
        Key features:
        - ResourceData hashtable stores all resource data in memory
        - Invoke-TerraformQuery stores data directly in memory instead of writing JSON files
        - Get-ResourceData replaces file loading for accessing stored data
        - No more temporary JSON files like temp_applications_data.json, etc.
    #>
    param(
        [string]$WorkDirectory,
        [string[]]$EnabledFeatures
    )
    
    # Set working directory
    if ([string]::IsNullOrEmpty($WorkDirectory) -or $WorkDirectory -eq '.') {
        $script:WorkDir = $PSScriptRoot
    }
    else {
        $script:WorkDir = $WorkDirectory
    }
    
    # Resolve to absolute path
    $script:WorkDir = (Resolve-Path $script:WorkDir -ErrorAction Stop).Path
    $script:MainDir = Split-Path $script:WorkDir -Parent
    
    # Initialize plan output path
    $script:PlanOutput = Join-Path $script:WorkDir 'tfplan'
    
    # Initialize used names tracker (hashtable used as a set)
    $script:UsedNames = @{}
    
    # Initialize resource counts
    $script:ResourceCounts = @{
        applications            = 0
        security_groups         = 0
        access_policies         = 0
        session_policies        = 0
        certificates            = 0
        browser_modes           = 0
        routing_domains         = 0
        hybrid_configs          = 0
        terminate_machine_access = 0
        terminate_user_access   = 0
    }
    
    # Initialize in-memory data storage to replace temporary JSON files
    $script:ResourceData = @{
        applications            = @{}
        access_policies         = @{}
        session_policies        = @{}
        security_groups         = @{}
        routing_domains         = @{}
        certificates            = @{}
        browser_mode            = @{}
        hybrid_config           = @{}
        terminate_machine_access = @{}
        terminate_user_access   = @{}
    }
    
    # Set default feature flags (enable all by default)
    Set-AllFeatures -Value $true
    
    # Handle feature flag filtering based on enabled list
    if ($EnabledFeatures.Count -gt 0) {
        # Check each feature alias and enable only those specified
        $script:EnableApps = $false
        $script:EnablePolicies = $false
        $script:EnableRoutingDomains = $false
        $script:EnableBrowserMode = $false
        $script:EnableHybridConfig = $false
        $script:EnableCertificates = $false
        $script:EnableSecurityGroups = $false
        $script:EnableTerminateAccess = $false
        $script:EnableSessionPolicies = $false
        
        foreach ($feature in $EnabledFeatures) {
            $featureLower = $feature.ToLower()
            if ($featureLower -in $script:FeatureAliases['applications']) { $script:EnableApps = $true }
            if ($featureLower -in $script:FeatureAliases['access_policies']) { $script:EnablePolicies = $true }
            if ($featureLower -in $script:FeatureAliases['routing_domains']) { $script:EnableRoutingDomains = $true }
            if ($featureLower -in $script:FeatureAliases['browser_mode']) { $script:EnableBrowserMode = $true }
            if ($featureLower -in $script:FeatureAliases['hybrid_config']) { $script:EnableHybridConfig = $true }
            if ($featureLower -in $script:FeatureAliases['certificates']) { $script:EnableCertificates = $true }
            if ($featureLower -in $script:FeatureAliases['security_groups']) { $script:EnableSecurityGroups = $true }
            if ($featureLower -in $script:FeatureAliases['terminate_access']) { $script:EnableTerminateAccess = $true }
            if ($featureLower -in $script:FeatureAliases['session_policies']) { $script:EnableSessionPolicies = $true }
        }
    }
    
    Write-Status "SPA Manager initialized"
    Write-Status "Working directory: $($script:WorkDir)"
}

# =============================================================================
# MAIN ENTRY POINT
# =============================================================================

function Main {
    <#
    .SYNOPSIS
        Main entry point
    #>
    
    # Suppress env vars that can disrupt terraform invocations for the entire
    # lifetime of the script. Restored at the end of Main.
    $savedEnv = Save-TerraformEnv

    # Initialize manager
    try {
        Initialize-SPAManager -WorkDirectory $WorkDir -EnabledFeatures $Enable
    }
    catch {
        Write-ErrorMessage "Failed to initialize SPA Manager: $_"
        Restore-TerraformEnv -Saved $savedEnv
        exit 1
    }
    
    # Set debug/verbose modes from parameters
    $script:DebugMode = $DebugOutput.IsPresent
    $script:DebugLevel = $Level
    $script:VerboseMode = $VerboseOutput.IsPresent
    $script:LimitValue = $Limit
    $script:QueryIndividualDetails = -not $Quick.IsPresent
    
    if ($Id.Count -gt 0) {
        $script:ListDetailsEnabled = $true
    }
    
    try {
        $success = $false
        
        # Handle specific operations based on parameter set
        if ($Clean) {
            Remove-TempFiles
            $success = $true
        }
        elseif ($Setup) {
            $success = Initialize-TerraformSetup
        }
        elseif ($Test) {
            $success = Invoke-Tests
        }
        elseif ($Validate) {
            $success = Test-TerraformConfiguration
        }
        elseif ($Plan) {
            $success = Invoke-TerraformPlan
        }
        elseif ($Update) {
            $success = Show-TerraformUpdate -Detail $false
        }
        elseif ($UpdateDetail) {
            $success = Show-TerraformUpdate -Detail $true
        }
        elseif ($Apply) {
            $success = Invoke-TerraformApplyPlan
        }
        elseif ($List) {
            $success = Invoke-TerraformList -Ids $Id
        }
        else {
            # No specific operation - show usage
            Get-Help $MyInvocation.MyCommand.Path -Detailed
            $success = $true
        }
        
        if ($success) {
            Restore-TerraformEnv -Saved $savedEnv
            exit 0
        }
        else {
            Restore-TerraformEnv -Saved $savedEnv
            exit 1
        }
    }
    catch {
        Write-ErrorMessage "Unexpected error: $_"
        if ($script:DebugMode) {
            Write-Host $_.ScriptStackTrace -ForegroundColor Red
        }
        Restore-TerraformEnv -Saved $savedEnv
        exit 1
    }
}

# Run main
Main
