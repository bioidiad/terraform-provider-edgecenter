provider "edgecenter" {
  permanent_api_token = "251$d3361.............1b35f26d8"
}

resource "edgecenter_dns_zone" "example_zone" {
  name = "example_zone.com"
}
