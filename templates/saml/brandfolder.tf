# Brandfolder — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_brandfolder_brandfolder_com" {
  fqdn         = "brandfolder.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Brandfolder"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_brandfolder_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Brandfolder"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_brandfolder" {
  name         = "Brandfolder"
  type         = "saas"
  state        = "complete"
  description  = "Digital asset management tool to store and share digital assets."
  url          = "https://brandfolder.com/organizations/<your-organization>/signin"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAIAAAACACAYAAADDPmHLAAAFRUlEQVR42u3d22scZRjH8f5BtqaKR6wieCGeUFT0RgviCS8U0QupKArihUWlQj30oBSL1nq6EkVBRTQKVsypTWySxqRNdzcJSd1usrNzenwem8USbWZmN9nsvPP9wdNSaGanO59533nnnXm7SUihs4mvAAAEAAQABAAEAAQABAAEAAQABAAEAAQABAAEAMRVADFV3GoCIAVvAQgACAAIAAgACAAIAAgAWs1CEEtvNaQy1mAtcgPAj3+FclFfncpY9x33AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEcB1EL3pyJemPIBAAAAAAAAAAAAAACwHgAWdTvj9Uh69bO/mg/l49lADswEsqfky5vn1f6yL5/OBfJTNZRTXixRDIDcAuirRXLHiCdXDbQ+9LpMf/Yu3cazk758ojCmvQgAeQHwwzrcb9ii9fxkQ043ojVrHQCQIwDNslbl3YoPgKICsNqs9dLJhnhtNgUAyCmAZh2oBG11BwDIOYAerd/OhgDII4Dbhj35QoeF354J5Wcd+n2vv3+mV/uvT/uyXb/wLSn385k/GwDII4CHxy584OyTv1MQV6QYRm4brGs3EAPAJQBZu5L+Fl/cAECXA7A9eHKikbitgzMBAFwEYPm9lrytV6d9ALgKYNKLpKd/9W29BgB3AUwpgK39dAGFBTCymLytQS4C3QWwu7T6Qbp+qC6t7ioAuhxA2Y/lyoR7ATZDyI0g124E6Uf7+svjJxrr1vwDoEsBzAeRHNKLutuHvcQZwf3lQGImg/IJwG7hPqZnuNWj4w3ZPurJjUc9uTjl/j010ZA608H5BdBOvXKKB0IKB8Amhl7UC76+Wighj4QVswW4XBE8rU3/1wuBzPsxAIraBVhd2l/Xf08EgKICaI4EntNuIeJGUP4A2ATPtqF/6xodFdhNn54WIDwx3tqIAAAbCOChMU8CHcSvLD+yN4lCeavky7WD6REcngsA4Nqt4CXF8KVe8F09kO6aYGwpAoBrk0GWgcUw1WtmL5/0AeAiAMuu6eSDdfMxDwCuArAHP9OMCuYy3B8AQI4AVHV/0uy3fTYAHARgs35p9vvQ7P+PBmyNAnvVfMd5ZV1Gmm3aMHXHip/9pRoCoJMAbJyfZr8/uMDzgTbEtEUpNq/BzScbogYRLUBHAYwuRan2+/OE+wHvlP2WD37P8kxkJ1bPAcCKvFcJUu33kYSmOdSW4EPtJlppCd447bf1AAoAWgQwof23PfyZ5gxdSrHfhmCvtgRpX0C1v/d2qTNnPgBW9NsDtTDVwbe6cyTb1freFN1Bs9mPOrxoWqEB2Jf9qzblD4x5ckmGJnpnxjeEDNjhhO5gdweb/UIBsGf/rtMz+9bhutytZ+69f3hyj9ZNx849+JG1fzYoJ+rZL82tO9hX+W93YH/eU+5ss18oAGu9gthHs0Fb340tPNXc3tbl9w2jDVwrEwAZ6hHtKvw2j5a1BLZQZXOcH2/wQqkAyLBEjLdGp6ohsOVqoi5YJRcACWUPixysBM7+B8sAWOViz5pom/WLY3cXtM41gKHFSO4/7sktw17iy51JF3c3HPXkQe3jd+pY/Jsz4T+zgkWIU8vF20TOlBfLkbPnVgy35eDenwlknw6z7PXvZtmVuC0MbSuEjeqQrl6Qg+08AAIAAgACAAIAkhMARahW3iYCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAuh/AQhBLbzWklqvUiIoFgBR8LoAAgACAAIAAgACAAIAAgHQtAAIAAgACAAIAAgACAAIAAgACAAIAAgACAAIAAgDiXP4GLLqfRwQMJRcAAAAASUVORK5CYII="

  using_template   = true
  template_name    = "Brandfolder"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://brandfolder.com/organizations/<your-organization>/saml"
    audience          = "https://brandfolder.com/organizations/<your-organization>/saml/metadata"
    sign_assertion    = "BOTH"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_brandfolder_brandfolder_com,
    citrixspa_routing_domain.rd_brandfolder_customer_fqdn,
  ]
}
