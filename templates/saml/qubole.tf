# Qubole — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_qubole_us_qubole_com" {
  fqdn         = "us.qubole.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Qubole"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_qubole_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Qubole"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_qubole" {
  name         = "Qubole"
  type         = "saas"
  state        = "complete"
  description  = "Software for Big Data - Big Data Service"
  url          = "https://us.qubole.com"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAt9SURBVHhe7ZprjGRFFcf/99HP2XkvLC95qpCgqCQaAq6QQALEkBgTH8Fo+CLBDwYEImJEEyMvNRBBIkSNJqIfVDDGSJAIIrKiEWTBbMwSltequOzuzM67u+/L36l7e3ZmuhtmV7Jr7D4zp+vWvVWnTv3rnFN1q67nXf1Apj4mv0j7lgYAFGnf0gCAIu1bGgBQpH1LAwCKtG+pBwC2OMx5/1XOjjKuVnJHqddhV2fVncNKXQAwlXz0DLnyFGSJwjSGU+4m3AcUj2d+lbRGCmclyq6DszJ1kO0hK4tUSlvIj3l2+GDo8i5ANjMl7ToFgjTHBMXDbAkQYvIBtwKecZtniW/l1kcxwKVeCRktQEidaAPaA/B1C3kTqQMA11cU8/yWUqcU7DQDiKBKb7lOGnCSlzaTDoOi5hsRdVPqBRieX87tDyA8rEFpxbTJix1C6gqAj9l7fgODp8MeilrnhZKLkSYmx/SWiSGF8RK3zR18lRJzl/UBEJQTPbdzp6ZSXCccojHAzgD0fwoAum6jkuDf5g6BWkoWp3RO6V965I4vyCDJQTm4SWR7I9HZn75TsxMnKi7V5BELMs9c6tAD0L0HBIDMuklww2GVRi1VgkA/oPOuAgorxkKSSE2CY4NA1p2jgtt5zL+1R6dWI91311WK987gSuY+5maHhzoBwKdT/LOMFVRSFAxGNTnzsl696YN6W1FkgedpWFfAqPm+j6OYs3RyGZMuE0MqaqqazSkkYi6UNwJgSeeNZHrihvNV//dzaBFiBQb6etzozaUOANpGaAae4t9qLOj2L12m8ZExRitiJog1RCDz0wVKzKuULVC2soqzImXyxHk8xd4QnjTOfTpqwrGwVhTrrFOP03VXXKhsYUr+YfB/ox4ugBWgEN0FhXmdvmmYGDgFOpirRX1GVH4dHuHeEMoDyAr2itSsiEmPFUAObCmZUjmZBwyaNXCh956MnGieeGqQH3oQuq4DMnAJmJ5cEJyb1vNfv0SnjEaaz0ZUR0dWCXr4hd16ZMe0ymGiydTurFYei1YTsMr4/Zy/QXEz0pWbj9LEkM0YVUXEjnKQ6fcv7dJ5Nz4kf/ho6gDeIQah+zqAUQvTpqJwUprZpRe/caFOHA01qyEN27yNX1//q7/qlvu2S2UWRgnl1uptI1ziZjxPZkLaN6Mtd12ks49gBUnhFuuJStDSH17crQ/c+oRUx8piptq15NYMdmGLJZt22w2hqa3W2tkYC3LWuYJ8HlK/XaQb5Xa4hiwg5bUQSAlbpZkh2/TnYcKEciYI5u36kfTtJGkj0X1yDU8wv9dj1TdE8obpWH1Rm7Jp6u4zwYVSZm222GIqJHfmCaN65/GjetfxIy4986RxjdVQAAuztYmtOfIp2lae/NkyFM1EPDn1mGG9m/pnHD/m2K5P2Ih7xdRzbXWnDgAsFqeMcOSXFEbMAj5KxiHNxqrQLwWMFCZs7wZKZ2lgRl5U7mDMB7B8Ldm7gvk8C54GrmDWgKUDJ1OpLasTysZNDWFJT12zWc9eu1lbYUuf+tw5+tBp1om6ajFBl0VZSUsuENvUacbgWTyandNDV75PT1/zfj1zbc5PI+uuj79DWlqkbG/qbgEFZoa4o3y42nfdryNnjivKLDNlijptWUYrr5djRpEkFgMBhAU4aYs/Qxv9I8ubvLzgitYL4j7PphbMfVoGKZ5gpWLNLFmuaKAHdQWgn2gAQJH2LQ0AKNK+pQEARdq3NACgSPuWBgAUad/SAIAi7VvqAUD7Dap49+p8BVtB7TKky5zfymlV5iDJ9MnldL7b/XfyOwEw/T3bcDDBtmVteWvYc9tcbdp/HfC8k0106sNuRzAkG9g5CsSP1V2ubzdpzeVt18eOyyy1PK/JVjDl5ThDVcBNKWgl7JU3dVtxJp9XZicgr1RUXRd1sQAE+Q2V0oR39E1km8r8VGFU05yP6DTfw4/jRam5V5riPXxpX8EzOS/CjSmud0t7eSdfQMGFaSWJqZYqodXUz7dKE9mpUKi4EnEdqJTNaVpllWIDMdVcTLnp7Wo0qdSYU+RO5XbR9g5ptokiyIyeQx8DAJ1NN7d3YAe71l5+jNOLOjdFyWV+LIwAlMfoxHbtuPkTOnkkUxSgYFQBoEx/2blHj768W0Nehc4sFZX3kx0hpvyNg8+CF2o+burys0/TeCUAAFPKjldjPbxjty646TENb0w1c+ul8pJpLQWjqtg2GAOxa3pOTdW0GNac3IBcgnKVKFAZnrHTO2SfktZUnvCR7StIbde6qZ//bU4fufNP0nDdhrUrdQXA7M9tY6WBwplX9NQdl+mM+oKaDF3JH5KfMvpB3Qq2K8C9yLa+zGoKY8OyMtzByGo9uPM1XfzlB3TseKAdt31SlbghO5ONVVFoZ5TOnXKHzFuxHSIjkzcHu806RrsENCHjHctPsEAA++nfF/Wx2x6RRoZ7alhotZ8ycz7QN5+X30JYXT/68/M8GVHg19TieRKYErbdZLzHmZ355VqWHYXZwWfGiFDFoEh8F024si5J3/vNNmnDkKYXpC22reVXsT6sDanuyDyNlMHuOM5GNrFjOeSaO7mTag8dbasdxuTxVqhGvq5Ht+2USq/vAh0AGNlN9x2AH6laP1rf/PFW3fP0doWMRjUrRjTBN2k41ai8Fsq0UGYlRzZqgRpeTU3Y+luiW76Dwc6LPH3t/j/qvsdfkiqjWqwcrfOvu1cvzfAcoB2oqLAPa5nzfM3T9iyuNIv7zAehZgByV7BB0z51vZJmKR8kWI/5bljX9b98Rt/59TaFFXOdHOxu1BUA26y2+OHhAgsBqoxN6orbf6vHX7Cd2SrmZmbJqCRmbmW1yh5Mt1awxS7bxydquM40qZIK14kw2dSgaOmyi94qVemlAVSjUmVEF1z/Xe2hXh2Q/bShMaxoJIs1jOuMMEOM8GwDfRyFN6HjOLwBEWM0Uk6a8gDo24++qFvvf0b+5ITSMm6STzFdqQMAZ6AEOTfFMMJJycyXUdn4dn34K9/S1j1gbd8MBPSwjCXQsSrdWcvmi3YwYbHeooVtqQcx5o07NCy+YEHH1ce05ZZLpZlnAYZA6o1oR+lInfP5H4r5Q1nVvhmgh6ZlYCm9JkArQJ8A1/JgAqXKcAhXx3Tv1lf02bsflI48kamamSAygHsD0OVorAeBfGaBbwYAmI6cYpnNCHAIEGuleDRcngY8MwcUZ2o8f/Pp+sVnztWwdmENk2oBcBWgnv3HXp1545NKa0PuRF4JlpLME82Z/yvHIJv6b0iG0j8xtVT+2LFKG4E7fwioG4dYBn/d6AAAsH+sIcTsCUI+/hjb8JtLEI3XCrEznFJad+ZvPq9kSa25Pbr8vLN0z6feo1b0qmZLR2kjNX+2Zbc++pPHGEEARl77uwQPS6w1X8OCu3rqKjJ3WySWOJVaC/JCd8BHpgwQaNHDCtYNAPKZl1mj2chbK67LphiKRlhHh3ybk/FtK29xxMw3hPc2dcMlp+irF+P/KHjz77bpiz99Eqem8zwOWACFSYX2iOoEvsjm//VoSAcr8Sza2NcmEfUSm8W5jyW0DNi82FpaPwBwhenIvuxyHXdA2FSVaCnEzJ1LWMmCyNtnNnarRP/dV2U2ssSFFm5091XnanFqSld//1UNsVpqlGxmYWFL6I8B2oKoO9hsEWx7ab+GbGGVEn1SA56ls8+UGRKPWqw6e0k4sBhQLEr2U17VQOhGdsxupV0NZyLGjIwlTfzcvhgbOoJl7yLLVwPVHiAT8JYtykzvgMgq5i5jY2KOYMGwLW4trR8Aox7K7D+yXk22EuisUZS1B1zuP/Jul8yfOwM7aMor5yD27rzRG0eXlWSKdmF+u7L9dv4Vz/lxZZbrr36+fHFQzI+Tm8t7PTowAP4PaQBAkfYtDQAo0r6lAQBF2rc0AKBI+5b6HADpPxg6FgkW1rMFAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Qubole"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://us.qubole.com/saml/callback"
    audience          = "qubole"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_qubole_us_qubole_com,
    citrixspa_routing_domain.rd_qubole_customer_fqdn,
  ]
}
