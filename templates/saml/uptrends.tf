# UPTRENDS — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_uptrends_app_uptrends_com" {
  fqdn         = "app.uptrends.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "UPTRENDS"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_uptrends_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "UPTRENDS"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_uptrends" {
  name         = "UPTRENDS"
  type         = "saas"
  state        = "complete"
  description  = "Website monitoring solution to track website uptime and performance."
  url          = "https://app.uptrends.com/Report"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAAlwSFlzAAAPXwAAD18B14rayQAAAAZiS0dEAP8A/wD/oL2nkwAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAxNS0xMi0xNlQwNTozMzo1My0wODowMPELhgUAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMTUtMTItMTZUMDU6MzM6NTMtMDg6MDCAVj65AAAAGHRFWHRTb2Z0d2FyZQBwYWludC5uZXQgNC4wLjb8jGPfAAAJ3klEQVR4Xs1baWxc1RU+Hns8mx1PPLZjx85iHGwSElKyNV1ICi0UQgVSC2kriqBFoKq/ItRWLYj+iCr40TZR1apUlYpEgSBoAi1CrQiBBiVhT5wmbRqHJDZ4mXgZ7zNjz+ae7857k1nee/PmLZl8kpW8O8937nfuved859zrigUGlQn45ooK6aFMsN0Al+YTdHB0lt6diNLJ6Tk6F56niVgy/aFMXhiiglo9VbSmxkUbFrlpe8BHtzfWSC/YB1sM0DM7T3/4dIKeG5ykiXiS3JUOqmaCVQ7+YdIOfgeEs4FhwCyJ1ALF+f8x/jeWXKAb6tz0YJufdq2sL/gdK2CpAV4cmqLHekaoLxKj2qpK8lRWUKWJQWNocR5dNJmiCBvy9uZa2rO6mVbzKrEKlhhgf3CaHjo9JAZax8SdPNNWA8Oc41Uxydtna72X/rqhjdrcTulT4zBlgHGela3HeukCz3hDdaWp2S4F82zocfYtD/O2+NO6pVKrMRg2wK8uhuin/w1SgGeh2oYZLwYMeyaRgv+k7puuoU6fsW1hyABffq+X3mev3sizbodjKgVJHv5INE5717bQrvaA1KofJRug+VAPRdg711TBl18dAAWE2/ta/fT851qlVn3QbYA4O6DAmz0ijCGsGYX8dXasnFAsQV9h/fDGlhVSS3HoNkDNG/8jF+91l8MY+flUigVQSkSIFH9litsWOyst9x/jHCW+1uCj1zcvl1q0ocsALW+dE57X6MxDDN1Q66LXNi2nJa4q0XaOxdK3TgxQXzStGazEKG+HH65YTL+7vkVqUUdRA2x/v4+6p+YM73ksy2801dKBjcukllxsOnaRejmMGl1ZagiyY3xpYxvtbKmTWpSh+a17e0N0dDxiivydGuSBI1tXimXLzkFqsQbN7ir69vF+mk5IeYcKVJlB5Dz6nyA1cagzApn8KxrkAQ9vqy6WtpC8VgJONuBy0vojF6UWZagaALG+nq1oxFvrJS8jwEa2mL8AHOzgXJx+yytZDYoGgLbvmTW2L0slD3zAosppk56q50iz68wl6akQigy/f2pIaPtSYYT8k+fHRHi1S1E6uF/4sB/xdlZCgQH+MjApYnapiY0R8gcuTdPjZ4dpkc2qsob9zNN949JTLgq++Wecz/tLjMsgj1BXKvl7PvpMeOv82U9wRID3hn6Y4X+h980A/fucDvoFc8tHjgG6p6IUZKeByo1eyOS1Ql0+MuQ9zgLyIIxBPbGqkf7Mqe6jnOBEOfcIsxAzAx+vgt/0FTrDHCH0wL8H6e/DM+TVqfisJo/+dnB/ryr0dzMLsuMmBBkwwgrx7a0raFu9T2rJWwEvDE2RW+fsW01+DP0tUSYP/IsFE0ps2B5G4WPjPdM/KT2lkTEAKrboG16zGOwgfxeTP7BBu7/dnU0UThjfCpjc/TyObGQM8PrIjLBwMdhFfn8R8sDGOjfFTKwARDYYEOJIRsYAh0NhUbrWQjnJAyh6VGYOE4wBGe3bzFVGxgAfTs5pVnMRkspJHnj60wl20OYM4ORxfMRcZWQMMMME1cTPHIegzXWespJ/lgXaP0ZmTVWjAJYDdGomzwDnwzFyFJn9g1v0VVgAq8kjN3mwe4CaXYUCDaU6nEfgJElHbUdM8gXmK0MY4BIPSm1lIewsYyJ6LW8H+Xs/LuwPZbXgXILqOWfZ0VhD1/mqhY+AjNcCTBjkccgQrFBCcmg4F/VPcnGlyOP8MMSr8sMvtVPP9lX0wo1tdPSL7ZTYsYaWupwU0VCNnHZRkn9fhjAArKkGVPA+i14OG2q40uRDt3bRZr9Hak0Dy/v0tg6R/OA9RaCrrI+EATR3Dndaxzn1N4/3Sw2FsJw896dFfuK2Ls4g1RO2P65t4WRKn2ASBkD40zICcgMIpftPDkotl7GnN0T3fNxvLXkFY+olD9zRVKPqC8RizxqmSIbenYiI6m9jdbpkrYZZtuos76+vN/rI43DQW2Nhocz8rLGvFvIAIoLrn2eoReH0GNsd9c7YHWvEs1gBLa4q/kA8awKZ2BL2uihhQTm6OXTgcCOfPBRjucgDrw5Pq0Yt8FwqnU0A4q12bzWl8IkOI2Bw6Bw/SsIJhQxUhspFHnjkVJDfVzYAwvq1WSfJmbcQT5Mma7MYMFZEKZUh4e0tJL/kUA+Bu5qqTTDF9YsUDLDF7xWqygxQtUElRy+KhToj5MX+16hm4/Ps8Jl585YGL3tOkyuAl1eHV9+1FbvI46BFC4gOtwYu3z7LvH0X79uISQMInR0pLprKRR7FVRzCYLvLyPwGjqc8rAfMVGBReHzqwpj0pIxykQdQXM0/LM35LdzHw0tGAUE1zHnFw6eHpJZc7BucKht5ABrmoWV+6SmNnKrwJ5wmdh7+RFFAlAKkz52cnT3Z1URra900MBenPRdD9AqHu2aOweUgj5UNJz9+23VSSxoF9wM6D58XBMze9cPgoBxx6xOpNrZH/mDFgK4AeWCSv2d3ZyP9+JoGqSWNAgO8NjwjEp+mLLVkBzAjII+szm7ykL/Ymgt3Xi+1XEZBL5CwDZwTYGnaBdgchxRnt62ynTyA+4SPq+gTxZ723dhKo6zn7QL0xlcbaqiD/YReGCWP2ccV21+yP1KCYm+3BHz0eVaGqLXZAfiFmwNe6ak4jJIHRmNJena9+nVa1R6PfmGluJisVS0yCpSl4Pn1wAx5VLPXcRT6Xmtu6MuGaq84Id63oU04D6uBE6jnBqakJ3WYIY+JQzTrvqldalGGZs/fXVpH97JyQgixEpDMSJx+cnZYaimEGfJwssPROL3DqxirTQtFe3+ZV8FyVm5alVYj8HPavJfF0c4TA8JLyzg2HiH/wbNCIxghD6A8/tSa5pxjcDUU6AA1YEYwKLMnM/mAYad4heFGV4ztgJMb1BTU8vliQHj9ActdvX9HoNsAQOObbAR+Xe8FilIgDyNbJpcKzPz9rXXs9fXfGC+JySirNtTTrPYJAIgbJQ/jBaMx+nlHQ0nkgZJWgIzvdA/QS0NT4uKzngsVdgKOMsTJ1t+2rKC7WcWWCkMGAERef6JfHJrYsSWKAcOeiKfEfcYz2ztKktTZMGwAGbi8dHgsTA28GsxmkHqA4aJyNc0ibTfL2yeu1V+DVIJpAwAfTEbovpOD4tgZHhxFSat3BoYZTi7QDOcoO6U/jbHC4JYYQAau1j/WM0xHQmHy8pJEHMeNDKPGwNCQyCBUYq8/snwx/Xp1M9Wq1PyNwFIDyMCAf983Ts+zozw9FSUnGwJxHsao4h9MXD4FxBXUCFC3h95Axgg5i0OWB9rSitQO2GKAfBwam6X3Mn88HaMgx2scn8mA6EGpDIqzi1PkTX4PbVvspXWL3NIb9uGKGODqBdH/AWGlvZOThHEHAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "UPTRENDS"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://app.uptrends.com/Account/Saml/<customer_id>"
    audience          = "https://app.uptrends.com/Account/Saml/<customer_id>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_uptrends_app_uptrends_com,
    citrixspa_routing_domain.rd_uptrends_customer_fqdn,
  ]
}
