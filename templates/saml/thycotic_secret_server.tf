# Thycotic Secret Server — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_thycotic_secret_server_your_organization_secretservercloud_com" {
  fqdn         = "<your-organization>.secretservercloud.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Thycotic Secret Server"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_thycotic_secret_server_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Thycotic Secret Server"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_thycotic_secret_server" {
  name         = "Thycotic Secret Server"
  type         = "saas"
  state        = "complete"
  description  = "Account management software tool to manage passwords."
  url          = "https://<your-organization>.secretservercloud.com/login.aspx?local=true"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAYAAADDPmHLAAARFUlEQVR42u2d95eUtRrH/ccEEQsICNJZlqU3sVAXEJYmvaoIFlY5LCC9L0uRfmgKKr2KFAVEsKBiAST3fuKde9ZxJnnfmSRvZibPOTlHf2B2JvnmKd+n5DEhxKOwSnc99r//CFKiEgAQABAAEAAQJAAgSABAkACAIAEARSoPHz4UP//8s7h165a4evWquHDhgjh9+rQ4duyY+Pzzz8XRo0f/sb744gtx8uRJce7cOfHVV1+JGzduiB9//FH8+eef4tGjRwEAvstff/0lbt++LU6dOiV2794tVq1aJd5++20xduxY8corr4iKigrRunVr8eyzz4pGjRqJxx9//P+rQYMG4umnnxYtWrQQnTp1En379hUjR44UM2fOFDU1NWLr1q0SJNeuXRO///57QQOiqADALb9+/brYtWuXeO+990RVVZXo3bu3eP7550XDhg3/ccj5rGeeeUaUl5eLIUOGiNmzZ4t169aJM2fOSDAEACQgv/32m9i/f788jH79+ok2bdr861bbWmiL5s2bi65du4rRo0dLMNy8ebNgtEJBA+CHH34Qa9asEb169ZKH8OSTTzo5dBUYmjRpItq3by/mzJkj/YwAAIPCrWLdv39frFixQtrwJA9ct9BC+Bw4kfgkPmqFggAAG/fgwQPx3XffifXr13t/8OkLR3PWrFnSacRc+QSEggAA4ReO3cCBA53ZdhurVatWMoq4fPmy1GIBABrh1hOnz5gxQ4Zkpg4CX6Fly5aic+fO0n948cUXxcsvvywGDRokQ8SXXnpJhn6Eim3btpU32OTffvXVV0VdXZ349ddfAwBUt3716tWiZ8+eeW86DuKAAQPE1KlTxaJFi0Rtba3Yu3evjOUhe86fPy++/PJLceXKFXHp0iVx8eJFGdZBBh04cEBs375d+hxvvfWWGD58uOjQoYN0+PL5Ti+88ILkFfAPAgDS5NtvvxXTpk2TB5frBhP7jxs3TvoMx48fl+zf999/L2N1HLK4Pggq++7du5IRhB0EQHPnzhXdunUTTzzxRM5OYp8+fcS+ffskhxEA8F/hoHr06JGTrecgUOfccBzGX375RR6cDacLEOHQoak+++wzMWXKFBkC5qoNYCqTIJK8AEDKy0fdQsHGjb2xq1C13PSkbhK/Aer5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"

  using_template   = true
  template_name    = "Thycotic Secret Server"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<your-organization>.secretservercloud.com/SAML/AssertionConsumerService.aspx"
    audience          = "<org-id>"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_thycotic_secret_server_your_organization_secretservercloud_com,
    citrixspa_routing_domain.rd_thycotic_secret_server_customer_fqdn,
  ]
}
