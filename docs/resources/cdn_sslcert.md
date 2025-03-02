---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "edgecenter_cdn_sslcert Resource - edgecenter"
subcategory: ""
description: |-
  
---

# edgecenter_cdn_sslcert (Resource)



## Example Usage

```terraform
provider "edgecenter" {
  permanent_api_token = "251$d3361.............1b35f26d8"
}

variable "cert" {
  type      = string
  sensitive = true
}

variable "private_key" {
  type      = string
  sensitive = true
}

resource "edgecenter_cdn_sslcert" "cdnopt_cert" {
  name        = "Test cert for cdnopt_bookatest_by"
  cert        = var.cert
  private_key = var.private_key
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cert` (String, Sensitive) The public part of the SSL certificate. All chain of the SSL certificate should be added.
- `name` (String) Name of the SSL certificate. Must be unique.
- `private_key` (String, Sensitive) The private key of the SSL certificate.

### Read-Only

- `automated` (Boolean) The way SSL certificate was issued.
- `has_related_resources` (Boolean) It shows if the SSL certificate is used by a CDN resource.
- `id` (String) The ID of this resource.


