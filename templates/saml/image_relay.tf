# Image Relay — SPA saas application template
# Source: SPA SaaS app catalog (internal), sourced 2026-09-17.
# Replace <placeholder> values (URLs, related URLs) before running terraform apply.

resource "citrixspa_routing_domain" "rd_image_relay_company_domain_imagerelay_com" {
  fqdn         = "<company-domain>.imagerelay.com"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Image Relay"
  ip           = false
  location_ids = []
}

resource "citrixspa_routing_domain" "rd_image_relay_customer_fqdn" {
  fqdn         = "<Customer FQDN>"
  type         = "external"
  app_type     = "saas"
  flag         = "enabled"
  comment      = "Image Relay"
  ip           = false
  location_ids = []
}

resource "citrixspa_application" "app_image_relay" {
  name         = "Image Relay"
  type         = "saas"
  state        = "complete"
  description  = "Digital asset management and brand management software to securely organize and share digital files."
  url          = "https://<company-domain>.imagerelay.com/welcome"
  related_urls = ["<Customer FQDN>"]
  icon         = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsQAAA7EAZUrDhsAAAkjSURBVHhe7ZsLcFTVGcf/9+4jm93sZpMYmJBoYxWByCAij47UamrEWFItlgR5DFbGoYjtiASmU8KrhaplGKZ2xBHDxNFREEOVIqgIaKutxEcy1tEp1AbsQxISyGMf2fe9/c69J5PdZHezZHezN2N/M2eSc+65d8/57jnf9z/n7AoysbLpz3j+q1b4ZRkQBKQFerZVp0Nd2XT8Yup0XhjJU78/gvo9R9DX54Moirw0tUiSDINBh8qqOdi160Hovr6zfOveM58jpNOnr/MMerZfknDi3JcozLZgduF4fkFl77NvYu0je5CdbUxb5xkCtSMUknDqr1/gUpcLgml/g+yVpfR2PpxQCJNy7ThdVc0LVObMqkVHeyeMRgMvSS8SvQyzmYwdBA370YZNtUHQTMQovQIFNhL8/hBE9k9cojR2xLBn+X24e8KVvGCAqqqZ6O5289zowPouGF5ukAM0HIZMAdbYYIBSMHVGCAVRUXotjt/xQ14QyYJ7tuHQ4SZYTVm8ZOSwFut0AiyWLORYTEP6x0acXq+LYQDWYa8Xq8umYfPUGbBRRSlJI5DzhTU7m+fiIcPhcJEjTG5CCHS/y+HHrp2H8OyeoygstClvvJ/4Bgj4UTtlGnbOvJkXjG3W1zZg/0snaTTQSOD0GyB6vAkE8KtpN/HM2Gfjpmr09vTxXCTRDUDWMYk6nhn7mHOyyPfSKI9C9CngdsF//2oY9CSOOIf+1Yplp96D2+eJrJsozIWQwNlAKvA3N85Ry0YJP0Ueq6kGpVdfwUuGmwKDOOd0YMFbf4RbCgFG8tAG4+UnIyX6wMeam/Dcl3/nT848CRngudbTII2qvEHl7Y840f2mbDzdeoY/OfMkZABJpsanUKfpmDE0QkIGiNtepg8STczXePqw9royfnPmScgAMWGLKJ9XVYvDJQqt7MN231KBmqsnqvdrgISiwKZPP8b2z5pVR9YPe6PUsfP3LkGRxcoLtUnSUSAqpBafmf1dzXd+OEZuABL31+Tk8EwkX7ucOE+jqC0ssbJeL2kIjZGUDwiyaRDGvrP/gLD3SZQ0Po9iShPCUknjC7C/VI+ZR//Aa2uD5JxgGG6aEktPvgFYaFRQrEcWLTzCk4mS2YLmjjZsaGnid2WelBmgqaOdlB7fV4yXSBUevdDG78o8KTOAAutgPNiUIY98b1ExL8g8KTOAEEspekkn9LnV5HGjvKQUW6bP5hczT2pHwGDI6x+vmI+LS1fiv/etgHPZKrwzL/p2WKZIrwFI/VWUfAsFWVkoJgeYEy6kNIIYa6+PKaV+YtUJL5czsb2eIDHbT1pGtEbb+SGJazQMHFBkKyc1gx4ihWBmp0kcI6vDJHU/7ENjfPBoY6Iw7JfJF4W5KfaCs0x6iA9NnKw6KNZ4lmjeTr+ylFdTWcVWb6xOKKTWYdvl1OHbwvb3bykqoXUuGYSGvfocL+Zdcx2/mnluv3UGui456d2GqIlBtLd1Y9HiciiHo3UtH+LJ1tMI0JBYSnO2YW45v22AjynOV71/Ah0Uxq632vCXiirYmeAJw0krw1tJDP3N0YMVV30b9Tffxq9og589/DROvt1Cazo9au4rR92mRaoB+PVvJOmNAmOA/xuA//3Gohjgt08cxA3Xr8LUySuxdk29ckFLeJ5ai96lU+Bc9wNIzm5eOoDUcxHOR25H76JSOH/6HQRPf8KvRNK3ex0cy2+A44Gb4DuwUykTtm97WX58+36MG5er6PlehxuzZk3Gm8e3KRUyjXvbMgSa3oZgtSu7ULLfDfvrnfyqSk+lHUJOPq1GSbuQPpHaWpF75CLE3AJeA3Ctq0TwzKcQLDZFn8jdHTA9sBHigX3voqgoDwaDnu7XoaDAhrdOfMRvyzz+k40Qcq+AQBpDMJkVnRH4/EN+lbIt7ymaRCCxI+jYaa8RYkExAsde5DVUCRf46BhEW776HDKUUFAE/4lXIPr9QVrFRq7kjIIJXrbbqwEE1rnw5rG3HPDxDBEKkgAb9LUaukFmgqwfn48MZOEZFdZn2dtHtotxDi8Ot7YfNcLkdT/hbYvRzoiXGqsOGVdzUUD2uCF1X1CcXaD5HcBMczaNaMoAro0L0bugGI4lZXDUXAv31sXkyAr51fSgGQN49u1AsOVPEPKLIOaPV5NtwItHkELxrhkDhD77QAlRgx3yEJSlS+osoBkD6MpmQe5z8m0EitPREltme5zQT7qR35U8mjFA9vI66CbPgNz5b8g9ndFT539g3vQCBHPqjuOEKRMflNkmQfjQ++ocSUvvK7RuTv77epeL5Oih+OweOhVoBIjjhm6nMyHk3lId4S/YSDJVr0HWklo17/ejd34+3X+Vkleg57ENHE1FAYZos0NHHRULJ0SmKJ1nSO3/VNTdSNGcAS4HtgjyNvya5k/0Q9pE0NQU8L1eD+8+WqV5HEym8dJosLZKitwVrKTvB23sjskp4H+3EZ7frVEblp0LmMjRxUzsAJZCpo0WSUl+n1E7Bjj+Iomg8epqjS2AEklxNUNiWkEzBoh5tjgSmGaIO4UG0IwB9N/7kRrrpRC1n+b3SBMtj6Wu8zB8fyF/cnw0Y4Csyp/AtHyDstaX3S5yZO7LTy6nsimS88Rh6MaHObw4RI8CZy/CFWiEQa+9w8yRQB1E7132xKOA3ijA6/Hz3NhHZl/OihEtohogL8+Mxx97lefGPr4DO2ilaee5SKJOAbby6ux04OGf34NHa++GyaSPecScKLIkw2Zjii2+tyc/Rj6ACaEkowKLAkEffI274HvtGQi5hZFhk0+BqAZgMCO43V64XT6E2A9+koQ93en14scL5uLgq3Vq4SBcm2sQeP81CEzopAIygmCyKMJpiD2HM0C66LjQg0fXL8TmLYt5iYqnYQt8B3dDzBuntC1VxOwWN4DI3vRoYs+z4I2jzTw3QPCTYxBy8pT/WaNTleLB+i6y39aNsg1AoXooyt7+6DWEtCIZSIR4511z0HXJAfZ9mXTDfq97ob0HK1ZU8JIBjPPuV46rlG2vNKMoxt4uGOZWQvfBqcNb29t7ce7seYRCzCrDjJsRwn6xnZNjwvpfLsJDq+fz0gH0k2bQ0taKUOsXtMz1p60drPOCqIfhjmqYV+3A/wBHjCIp8ko9qwAAAABJRU5ErkJggg=="

  using_template   = true
  template_name    = "Image Relay"
  hidden           = false
  agentless_access = false
  mobile_security  = false
  sbs_only_launch  = false

  sso = {
    type              = "saml"
    assertion_url     = "https://<company-domain>.imagerelay.com/sso/consume"
    sign_assertion    = "ASSERTION"
    name_id_source    = "email"
    name_id_format    = "emailAddress"
    saml_type         = "SP_IDP"
    sp_initiated_only = false
  }

  depends_on = [
    citrixspa_routing_domain.rd_image_relay_company_domain_imagerelay_com,
    citrixspa_routing_domain.rd_image_relay_customer_fqdn,
  ]
}
