---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "edgecenter_region Data Source - edgecenter"
subcategory: ""
description: |-
  Represent region data
---

# edgecenter_region (Data Source)

Represent region data

## Example Usage

```terraform
provider "edgecenter" {
  permanent_api_token = "251$d3361.............1b35f26d8"
}

data "edgecenter_region" "rg" {
  name = "ED-10 Preprod"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Displayed region name

### Read-Only

- `id` (String) The ID of this resource.


