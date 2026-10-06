terraform {
  required_providers {
    citrixspa = {
      source  = "registry.terraform.io/citrix/citrixspa"
      version = "1.2.0"
    }
  }
}

provider "citrixspa" {
  base_url       = var.base_url
  token_url      = var.token_url
  customer_id    = var.citrix_customer_id
  client_id      = var.citrix_client_id
  client_secret  = var.citrix_client_secret
  max_concurrent = var.max_concurrent
}
