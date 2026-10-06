# CloudRanger — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_cloudranger_console_cloudranger_com" {
  fqdn         = "console.cloudranger.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CloudRanger"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_cloudranger_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "CloudRanger"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_cloudranger" {
  name         = "CloudRanger"
  type         = "saas"
  state        = "complete"
  description  = "CloudRanger is the world's easiest to use, Enterprise backup and disaster recovery platform for Amazon Web Services AWS EC2, RDS and Redshift"
  url          = "https://console.cloudranger.com/#/organizations/wixHg92G/accounts/j5JRQNSx/dashboard"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAACXBIWXMAAA7EAAAOxAGVKw4bAAAABmJLR0QA/wD/AP+gvaeTAAAAB3RJTUUH4gYFDRARB2y0SwAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxOC0wNi0wNVQxMzoxNjoxNyswMDowMMHns+UAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTgtMDYtMDVUMTM6MTY6MTcrMDA6MDCwugtZAAAG9ElEQVR4Xu1aeUhVaRT/PfX5zBQ1bbOiaY+arHQSbJmQFksqK6aFsqJssmaEIgomypj/WjAiCiSCrIyKoLINy2LabBVN22Syxco0NZc0ffr0Pef9jt+VKS2agZjl3h/o+d69537fOb+zXXjP1OQEdAwXJXULgwAldQuDACV1C4MAJXULgwAldQuDACV1C4MAJXULgwAldQuDACV1C4MAJXULgwAldQuDACV1C4MAJduEjf/qq9BkrUBTRT7gsMv1j1FVVaVWwNu3b2G3t633OdTU1KC6ulrW/K6moaFB1l8brQioqm/AT9fvwe/gZWw5tQfYEwbb7lCYcpKc2q5Kqxnnz59HWFgYFi9ejFGjRuHVq1eIi4vD8+fPlcaX4+jRo9i3bx/y8vIwdOhQxMbGYuzYsXjw4IHS+Dr4gIC0/CL47UlB4rk76FNbjPjGJDSYffAyIAzDqqOUVjPevHmDlStX4ubNmzh27BguXryIHj16wGKxwM3NTXSys7OFJA25ubkiGd2nT5/Kura2VnQo27dvL/dGjhyJvXv34uDBg1i1apXovX79GidPnoTVapXPL168QFFR0Qf7MwBnz55FaWmpZBRBu3JycmRdVlYm13lN+0awhYD88ipEnEmHo7jMaRWw5ZuHaHrrACzeiDVNRU5BPqJ/y1TaQEpKCpYvX64+Ae3atRPJjb29vZGYmIikpCSUlJRg6tSpcm/ZsmUiCwsLsW7dOlkzyg6HA9evX4e7uztcXFykFHiNBEyaNEn0Nm/eLNdmzJghn2fNmiVZc+PGDaxfv16cW716Nfz8/DB+/Hi8e/dOMrOurg6nTp3CmTNncOXKFdnv1q1bMJlMsk8LAasuZTnz38laYxMCOnlinPUOah0WmBprcNh2BK5NZhx58gpVtubapKPaJn+Gdu3AgQPYsWMHFixYAB8fH4mYr6+v3GOGcM3siYyMxOTJk8UxRtfV1VUiuWjRIjx+/Bhr1qyRZ7Zu3QovLy/JFIKEMwM3btyI+/fvS+Z07NgRAwcORK9evdC1a1fcvXtXMpWkXr58Wc6dP38+NmzYIHsQLQSkPXSmZLWTALsbdoe8BgLtcPV3g5u7GcFF02B3RtLeaEd2WXPDmzZtGnbt2iVrorGxUYwgATyQfxoqKyvh6ekp0SAYaUaTUktVOs9nucegQYOQnJyMR48eSbbQQfYWlgZJILQU1gjp0qWLNOMLFy5IdnIvkrR06VLJkISEBMkKrTw1tFhpc6hvyV0bEXenE2Kfx8DDWorV9T+gsMoGk60eKCh29sHmR7p16yab0tioqCgxjs2PjtGZ7du3Y8KECVi4cCGGDBki0QkODsbs2bMRHx8vDoSEhEgKs5TOnTsn0ecEef/+vZzB+o6IiJA1M4hlZbPJbEJ9vdMeJ3geSSNYbunp6Zg4caKso6OjJbPmzJmD1NRUcV7T1dDy+4DBu1Pw1Frn9N+O2sI65IVnoL7ajG+zQ2Ay16HJeRB8vFCzNhqeigQNNIbNry3QITqmoS3dj3X+Dnbu3ClZFhMTg7Vr10qtjxs3Tt39NFoI2JeZi8XXcuDt54sBFgcyvBLhXxSH8oKXsHi4w+brjR+H9MHu0UPlwX8j9u/fjydPnkgGjBkzRl39PD74hUhkcipSS624GPQ7jlX3wp5yH3i5OVBptWF0F39cm/690gROnDiBAQMGwN/fX0YaS+BLoaUyo87a/ifR6icyCel30d+Wjqic7s5u5UC3dhb88t1AxAX1VRrNOH78OK5evYoVK1agoKBAOjbB7sua58jx8PDAvHnzpKsfOnQIM2fORL9+/bBp0yYZg2lpadLEMjMz5Znw8HBcunRJ6nT48OHSPNnJhw0bJi9ap0+fltoeMWKEnB0YGChEFhcXS9Pr27evdH0+x76UkZEhNrEh892gd+/eUn4cmdOnT4fZbP7wRYhYM7o/poT/jDcxEShZFImCJVNaOU+wUW3btk0aWnl5uTQ5HsDxxgZIwyoqKkSXBHEUspsT2qtyfn6+GETH2bFJBg3lHpwCJJMZRhIIOsQav337tjROTgnawVEaGhoqE4PTaMmSJfKywz3v3bsnOhzHlFlZWWIXnSdcf3VCVhqa3IVNL4sZ7Z0j8FOgE5wEPLh79+6yKR2gMXyeEeM87tChg0SIhgUFBclnjkN2+MGDB0vpsJMzA/iyQh3WL7OHztMpTovOnTvLOcw8vgQdPnxYpkxAQIDco7NckxySOHfuXHGek6dnz54SIOqQBDZHBoRoVQL/dXBc0uFP4dmzZ5KpGv53BPxVtOoBeoNBgJK6hUGAkrqFQYCSuoVBgJK6hUGAkrqFQYCSuoVBgJK6hUGAkrqFQYCSuoVBgJK6hUGAkrqFzgkA/gD0Bj3axtJkpAAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "CloudRanger"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://login.druva.com/api/commonlogin/samlconsume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_cloudranger_console_cloudranger_com,
    citrixspa_routing_domain.rd_cloudranger_customer_fqdn,
  ]
}
