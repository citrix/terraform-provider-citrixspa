# Expensify — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_expensify_www_expensify_com" {
  fqdn         = "www.expensify.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Expensify"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_expensify_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Expensify"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_expensify" {
  name         = "Expensify"
  type         = "saas"
  state        = "complete"
  description  = "Expense management tool for expense report management, receipt tracking, and business travel."
  url          = "https://www.expensify.com/inbox"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAADwAAAA8CAIAAAC1nk4lAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAApSSURBVGhD7ZgJUFRHGscfDIOAB2pwAJVDUBSjRrnRxBgvSFzXym6ZuJty1yyrJpvy2EowtcJaMSGlxjUxiahELgcH8CAgcq7CACqgJegMM9wooHgNOMghCq69/379HG7ErcrUpmr+9dXz66+73/tNv+993ciRX6EM0PqSAVpfMkDrSwZofckArS8ZoPWlXzt0YyrRJBHN6WFYEmlME2YRUtuQUlUXV1V//MVWF19df+LZM2FiVOWD/WrND+rG4di/SjS1LU/YxOfQxd6kwI5ccupll51JEW+Xe8dhBZOIYhnm/bvg/WOpDvEZ02HHM2ck5rjqDJFT51xT8189lUV9ZrHpU49nzMTE9+T13E8KLrLkJSy0uL3zP5j7HDp/HCly67Yr80ixW9fleTczZt1Mn/WMb9JgzwEFkzFPljYl4Zznz1leidlexzM8/vap89+Dp20NmvZJoHNmoW9Y3Fx3H6sDR+ck53pjDLNjqZMIaZ5ysto8Rm0pKx3aRh9TcxFKLlI5hnfkDS146HPoAkk3UJkHueVNmn33b7PneP0U7EC0vjSodu+GLnTEvLh0FwadJPeKTfXw8LVatsJm6Ts2C5dKCivemD5rLMcZTbQflV08P+GcAC1Lswf09J+vAxoopkdVIwehHxWj5o4oVA868CD6WiKUF+60Up8SQzroCo/wHY6MlZcRb4ISv51KSj16Qsf3gI5L9ViwSLLi9xPfeXfikretL1UC2pLjTOydRsuvDgzNhStXy+slsaVcGFJFaXJUBdBu4jAFy+NzDa10vQeGRg5UegDuWryrNve1xuw5TXJqmuw5D/NeS/txKrpInddg0PHpHp5Y6d/YgBiLDegZs7HSJg7OA0MDIqG2mT2/rfNpbI12dXaduVTFfgCWtq6VEqfWt3CHr7H0GAi6zOO7z+xtrc3JbW9yza2v3fPhROIT3zgRtceA0AlZXvz7EHOcMV3pqhdAj5CqgHKorJEh6PSgowu1oqr5Mfy4mmbuECXG2g8EjWVumY9n5EW6EI0Pqffqa02+x79xmmBlRvO7qC80LKPQl+NEnMgU4Os3Tckve31oaKTHKH79RNEl0qoHDKSnwsqb2BqzhOkHXSghlZ6es8fgeSKx2NxcbNbPzC1MOWMTDFixcDyp9iCX+kJjsXOuLcCAd/8w+XzJgqyi+S+E7k7fCKWZVC2r1jIcpoALt4yjS9iYgaAVNoeCHPC8vAiXLz+ejCdxJqZ9jRNhTNJ3NLOT9zuTa06Y1xMaVS/1gg96V66elHnJd/jQzETRKtSKpo4uRgTl3G4DqG5AP2il7beBKHAiQt4g1Z5ISpHpCJMR3YYmgqTRh3S9DqzYXU5EMQB0Sm9oFDvcc2ho8VEVv3cox8SoGcsexT3mQFzYEOmhtP3xH4A2dnUeZWlphnXtScwMyWo9wdzJ3gLD6Of4Img0s68uEJuOsHMcNRg0atwfc+rja7SbChoIv73/Ne8m90PRhbtttEGI2+kqM6lqiJW2Aw1nbMp/SagAAxm6MIAzlg1jpRHJUy4wEY+YPBi0VI1lvny/nSFAH8jrULnxY2zjy1jk62v3uSghrftBF9tkHpw2fx72AjGW+c+rJO/5W/W09/2t1q6cwNCXzR+XHz2dFE/BvKGhcxRDQvM7IvLYJr6s8yldZ5u4MgsaLOXCFUm1DxEpfdCBAYNAo+RVeqpOzqCLzZmg9iGze1mNJ9GgokFGtDiWuvcvef8DtEATVfK7rDrcbXE6Ld6IYGPHe+DJiHFUCU4gg0BfdccJCY+cO2M0ue5J95o+dsPLRmJOoUvcSXHfsweDPnPeG3fwX2Wrg8abkdhaDACdUKODpusdqfy08Da+SEtZ92cnrcQ+8xQpNAj0xbEU66ob3URueJEr/YiZ1fK7DDvx5U/EPFmaow4ahv1lsuPowC9cUi54gxLoC5dZr15rl17go4OOScFErWPvUx648VHqmjCAWsWVYlfHBsQiSHd5Q09oxXKSb00K7YZr+RNI6RrMyyv6RHrGJjbNSbB055R8l4TsqbpI8nmXpNxpuuaxVLvELF9M3F50lzt4Ffk6lOFk19vnWXXQVPgRKDT4lp/bkxrSXjGAtZWT9kphEt50a02jVtXYrB6GqZqahUyFNI+6cIirb6WGkxGO9zfbOmF4NnxcWRez6w/pUYRJgJ402U5kOtJYbMGJzKmZIHe5Jq1QLPWg8WNRuASYD9asge+3dAlr9pcw7hUrlDPOyXnqDNeZLtNnwMaNf6WtTX/QgYGBa9f+CU5o6EGQLF6yNCkpiXX1lwAtsbbB0JZWmuY6Xb9+/Uh4eE5ODmselUqjoqLu3LkTERn55PFjqVS6YcPGhIQE1gt1dXV9FRLy0Ucfp6UJf/amZ2TExcfD2blz5+bNW5RKJYvX1tYGBQX/JSDg8OHDLCKXy8+ePatSqd5esQIkISEhmZmZeERnZycbgHuGhYUxvxd0SkpKZUUFZuLu1TXVtJtXS8tDoMA5GHpQJpPBMTUz43uo3lq8GCMVCgVrjhxNj4qLFr2F4BQnJxoyEtErr6qqqty8PDhWEomfvz8LYqSJmB5oN27YSNtGxmYWOCxQ7d27F70QfDNzc8Fn/1Bo4+5bQxKJNeLNWi18i5E493A+vvMRiY6Ogj933jx+HnFzd0fz/r17IhPxmLFjWXD37t0IVpSX+y7AAZ1jC3/gwAH4X3711boPP4Qjz5YjWFZWFhAQ8Ki93drGFkE6dw+dW1RUBB+OvYMDnJiYGPi619JrpY/JZPn5+bm5uXhZV/hpUGRkJLog1oyOotBYb9Y8ceIkmoGB23AdYznW3sFx4qRJEmtrNLdvD8JL0E28fPkS/M8CA8vLy+EwLffzZ+nHPio4wf8MhpOZkQEfr4sFF765iDlMgsegG5uaWLOnZs+egy7o5MlTaDLo6Oho1hsXF4dmUFAQrnPd3FLTM6Sy2NNnUrLkOffu3/f2obs64c9vFy9ehL9p8yY2EeuyZetWRKDzeXn4tXAQ7wmdlZUFf//33+O68rer+HlUvaDxCe/bt28Prx07diC+a9cuxLd9/jnLEERYTtvY0u0QQsFBE3VmFJ/KLBgbG2siFt+4ccPbl0I/4/9PiUF/sXOnn5+fbmRycjL8iIiIKU7OLNgTGoLPVFhYyCKQMNmod0IzhYaG4mpuMRIDiouL4SMBkhIT4bjOfJWO4LVu3ToMqK+vE9q8li1fjiAyEv7Tp/SPkezsbPgbNm681dDAD+GsJtCUwLeEXvZFwdmydQscXb1j79BIRMfoJEBXVqJsVOCqE5olJSVqtfoJ3a2oqqurEWffU0R4uFarRZKUqrt3OCg9PR0fDUoEa6JoKpUlzH/06BGK0r27d1kzLS01Kjo6JzeXNTEFj4Oj0WgwrL0dGyIV6iweFxwczJpMAvTwdeTIEdzl65AQof2LCRsCHsQkhJ7rpaGvXLmyfv0GVBih/Yupo6Nj85YtSI9GjUYIPddLQ/8/yACtLxmg9SUDtL5kgNaXDND6kgFaXzJA60eE/BeWvp0ev/NmIgAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Expensify"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://www.expensify.com/authentication/saml/loginCallback?domain=<customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_expensify_www_expensify_com,
    citrixspa_routing_domain.rd_expensify_customer_fqdn,
  ]
}
