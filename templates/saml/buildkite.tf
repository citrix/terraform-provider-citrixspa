# Buildkite — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_buildkite_buildkite_com" {
  fqdn         = "buildkite.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Buildkite"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_buildkite_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Buildkite"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_buildkite" {
  name         = "Buildkite"
  type         = "saas"
  state        = "complete"
  description  = "Infrastructure tool for continuous integration software development."
  url          = "https://buildkite.com/<customer_domain>"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAubSURBVHhe7ZsLcFTlFcf/e/duzGOTkJCYEI0kPAJCURFCDRBDq6BmIgxSBkdxBAG1Ai04dYZCgaIoIqG+GIZ2hhmRl/IYlGqDZaoypUqhTrEVIRTyICEJ2WSTTbLJJvu4Pefb+y2bkJBNsiAJ/jJ37v0e9+49/++c83337gZzG/K0sy6rdiPzz/pibcLpzdpSxxGtxt2k1wYHwz22bVq51oixxgSsD8vESGMcbhQO157BL4s+xvnGS4gNjUbisFRYXHY8qg7CH8KzEG0I1Xt2H8O4uh1aM9xwah5YtCbcbYxHbmgW7lbj9S7Xn7ya7zG/YB/KWmyIUiNgMigwGVUkpQ2C2+NCs+ZGmacRU02DkUuDlqBE6Gd2HZ8ABvqDBrTQcbXmwAjyhA2hmRijJuhdrz37q/+LxUUfodxpQ7QaToYbRT15qk8Aj8ct6hiH5kIV3evDaireCstCfyVMbwmc1gL44SSVLXTxIUoM3g6bhLHXUIg91SdpxA+g3t1Ihof5DJd0JADDbXz/HMaPmgZhI3nvrUq43to5HQogII9wUhurPMIQi9fDJmKcKUlv7DnbKk/g18WfwEaG96MRV6HQHemNflxNAB90rw64xKDlkEfk3pKJOAqfzri6AH6wR1ipZ6oSRXGXhQy1+0JsvXQMi4oPwkHx3I9GXKUYvxoBCSChvg590HKMKXidwjhBNeuNVxKwABJOltWULDk0Xg2dgPtNt+stnfNW+d+xvOQQmjxOxJDhxk4Ml3RJAB0+h4WoJeseVlPwBiXLW9tJll0WQMIeUUNnJipmvE1xN/EqQuSWH8GyC3nkpRqijaEBGy7pjgASPpft8ybLFLxJ3hvnlyO6LYCA4s4FD128CbdRaKwLnYjJpjv0RuDVi3/D6tLD1ENDv24YLumJABIphJWFMKVSsrxfCNEzAfxwUWjwxZPUaMyoT8BvC/8Co2KkEb8FSjcNlwRDAIkQwkD3qrRglnNg8ASQeBQDKopL0GJvgkoCBINgCqCoRribnbCUVcBR38DzTnAxGgwwKyF04eAIGiwMRgWKB7AUlaL4zDk02RsRbiLv1Nv7LGy4kcysu1iJojP/g73BDiN5gUIDxUPUZwUwKGQ45Z6GimoUfZ8Pa02N1wuo3kDGS/qcAGw45x57pdfwqqoqiksSw2hsZbikzwjAhitkZIPFisJT+bBYyHD2gg4Ml/R+AWjWUclIu7UWxWR4VaVF1HVmuKRXCkDrLxpygxhxp81Orn6WprVLoi5QwyW9TwAyTlVVuO0OVOQX4GJJqRBEUVsnt0BReE1/o8MG8lsg1aTC09iMsvzzKCkqhsPtEmKw4d1dyCnDlViUeexoYSFoxXUjwSvARnpytDrrkGkeiEKaxy8UFKLZ5RKuznN5T1F2mXNw1Pw4PcxEolTThfBG2Q+GMNzdguqWWvzMnIpLY9bg47Q5sDRYhReIRUwQjGdEDkgzxuCv5hn4BwmRSs/5F8kj2OWut0ew4XZhuA1T+qWhKn0tPrlzHm4NiUSzu1nEf7BplQRZiE/M03A88gk6jhUe0ay5yB+urRCXDa/F1JgRqBn3Kg6kPY3+po7f5ASLdmeBFCUaH0VMxb9IiJ8Y41HuaRRvYIPtEWx4PY0sj/iM2FFovm8DPhj6pHg/eL246jQ4kITYE5EjhODvC9gjWAi+8Z5w2fAaPB13Lxp/uh7bhzyOEMWk97h+BLQOSFai8AEJ8Z+IJzFGSRCvoIVHdBE23OZ2oNpZj/nx6dAy3sIfB/0CYcYQvcf1JyABJAPUKOwwZ+PfkbNxjzEBhe66gDzCQ+11LjKcXP1XCRPI8FxsSp1+TZJaV+mSAJIkxYzdEdkoiJ6HceoAXKTQaBI5Qu+gw4bXuppgdTViyYBMaOPfRO7AHGr54Q2XdEsASawhFNvCH8b3kU9jonobijSvR7g1D2rdZLizASuTfg7tvjew/o5s/awbCwO5b9BSe72nGb9xHsXe88ewxJyOVXdM0Vt6Dq8DQo+vQFwIT40996AWjxtOWmUGVYBrSbAE4Gcfm7MRibf0w0Faa/QoBHoTHJZVNPuE0VT75+FzUX7v75BuTu77AngNp/WLx4WdQ54gw1ciJ2ak3trDJHgjwzNQNc0+DorzP9GU20jL6ydo0dWWPieAb+qlUd+YnE2Gv4YFCRl665X0GQE4l9tosWV12bGcp96MDVialKW3dkyvF4ANb+DniuYaLEwcT2uODViTHPj022sFYMPFI7TTRrE9GtqEt/HmwEf11sDpdQKQ3eI1GT9J5vS7E/Xp67B10Ey9tev0LgFoKuMHqixzCqzpr2FP2myY1Z79VrDXrASrnXXIOLUFecPmYnBY8H7D2GsEuFb0mWmwu/wogL6/aflRAH1/0/KjAPr+pkWZPXs2Zs6c2WpbvXo1bDab3iVwUlJSxJeWH374oShPmTJFlBctWiTKHWE2m0W/ffv2ibL4upu27du3i3J7PPbYY6LPihUr9JruwwuhdrfDhw/zOilg5HmTJk0S5QceeECUFy5cKModERISIvqRAKIsr0MCiHJ7yD4RERF6jaYdOHBA27x5s2axWPSazlHoAnQd4MUXX8TevXvx3nvvITzc+93c5MmTxT5QNm7ciBEjRmD37t2izD9J8993BH/Xz/CI+tO27A8Zi7S0NHzxxRd6DbB06VK88MILKCws1GsCgBWknXbo0CFdE00rKyvzKVxdXa3Xdh0SUFxj8eLFek37kOCi3/79+0VZfvaOHTtEOVDkeSUlJXpNaxwOh350Gd/QyFFgyGj9CIiNjcXWrVvFaAwfPlyv9cJ1vB09elSUefS5zF7UEXl5eYiMjPSdu379ekRFRemtreF2yeDBgxEfH4+kJO8/asyaNUu0v/TSS6Ls33fkyJGgsNJLwMGDB2EymRAaGir6JSQk4Ny5c6LNJ8C2bduQm5uLlStXYtSoUaJuzZo1Yt/S0iL2/sL4Q0KKvdVqFXvZvy07d+5EdnY2Ghoa9Bpg2bJlqKio0EutkaETFxeHgoIC8aNHeeMySTc1NYm9P3V1dXA6neKYB2PatGlwuVxITk4WCbeyshJDhw4V9+sTgNxNqLl27VpRZhVXrVoljqW6/IOk9mjb3jbmZZlnHCYjI0OIxpusaw8eqdGjR/uE5xuW+Ul6rBTfbreLPfPZZ595fyFKzJ07V+yPHz+OCxcuoL6+Hs8995yo43zhu9Nnn30WW7ZswaZNm3DXXXfh1KlTPsP8w6M7sDsWFRXpJeCrr77Sj3DVqe6hhx7CyZMnxTEbHxMTI479kQJIYRgOk/79++PLL78UZbbjxIkTePnll7Fu3TpfeBw7duyyADNmzBDK0JSFb7/91jdq77//vu/iUpCuwgKeOXNGHLc3I3SUA6QbM59++ql+1DnyPHZ1hkViu3h9s3z5crz77ruivri4+LIAHCP+cPJjeOSk4ew+3YGvPWbMGHHs8Xiu+CyO2Y7gcGGeeuqpduO9PcLCvP9AmZqaKvY84jLk5MZ5ivc+ATi5nD17Ft999x2ef/55XwzxWuD2273/EMXJixYa4phDRtKZZ/CHcQaXcIixmCzE+PHj9doreeedd1qFCyeuQOB8Vl5ejvT0dFHmz583b544ZubPny9E4VUvu6dv/my78UpO0l673D7//HPRh7K1KFMuEeWsrCxRJkFF+euvv251Xttt165dop8sy3UAhY+vbsmSJaLuwQcfFGUyRpSZzMxMXz/eGFootarz32jBpBmnT5/++yFDhog5ftiwYeBjdjtOFq+88gr188IzAucGzsg8ndDNiizNo8/TDCee06dPi2lm6tSp4jo8dbHLczLjEGBPYlfmlRotSjBgwADs2bNH1HO/Rx55RMz333zzDRITE8V1+fmCp0HOE2QP8vPzsWDBApHROdb52mPHjhX3yBmf457r+fOfeeYZcf6cOXNw/vx5lJaWirUAe/WRI0fIjmT8H2AD0yppjMjoAAAAAElFTkSuQmCC"

  using_template   = true
  template_name    = "Buildkite"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://buildkite.com/login/sso/saml/consume?organization=<customer_domain>"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_buildkite_buildkite_com,
    citrixspa_routing_domain.rd_buildkite_customer_fqdn,
  ]
}
