---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "edgecenter_cdn_rule Resource - edgecenter"
subcategory: ""
description: |-
  Represent cdn resource rule
---

# edgecenter_cdn_rule (Resource)

Represent cdn resource rule

## Example Usage

```terraform
provider "edgecenter" {
  permanent_api_token = "251$d3361.............1b35f26d8"
}

resource "edgecenter_cdn_rule" "cdn_example_com_rule_1" {
  resource_id = edgecenter_cdn_resource.cdn_example_com.id
  name        = "All PNG images"
  rule        = "/folder/images/*.png"
  rule_type   = 0

  options {
    edge_cache_settings {
      default = "14d"
    }
    browser_cache_settings {
      value = "14d"
    }
    redirect_http_to_https {
      value = true
    }
    gzip_on {
      value = true
    }
    cors {
      value = [
        "*"
      ]
    }
    rewrite {
      body = "/(.*) /$1"
    }
    webp {
      jpg_quality = 55
      png_quality = 66
    }
    ignore_query_string {
      value = true
    }
  }
}

resource "edgecenter_cdn_rule" "cdn_example_com_rule_2" {
  resource_id     = edgecenter_cdn_resource.cdn_example_com.id
  name            = "All JS scripts"
  rule            = "/folder/images/*.js"
  rule_type       = 0
  origin_protocol = "HTTP"

  options {
    redirect_http_to_https {
      enabled = false
      value   = true
    }
    gzip_on {
      enabled = false
      value   = true
    }
    query_params_whitelist {
      value = [
        "abc",
      ]
    }
  }
}

resource "edgecenter_cdn_origingroup" "origin_group_1" {
  name     = "origin_group_1"
  use_next = true
  origin {
    source  = "example.com"
    enabled = true
  }
}

resource "edgecenter_cdn_resource" "cdn_example_com" {
  cname               = "cdn.example.com"
  origin_group        = edgecenter_cdn_origingroup.origin_group_1.id
  origin_protocol     = "MATCH"
  secondary_hostnames = ["cdn2.example.com"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Rule name
- `resource_id` (Number)
- `rule` (String) A pattern that defines when the rule is triggered. By default, we add a leading forward slash to any rule pattern. Specify a pattern without a forward slash.
- `rule_type` (Number) Type of rule. The rule is applied if the requested URI matches the rule pattern. It has two possible values: Type 0 — RegEx. Must start with '^/' or '/'. Type 1 — RegEx. Legacy type. Note that for this rule type we automatically add / to each rule pattern before your regular expression. Please use Type 0.

### Optional

- `options` (Block List, Max: 1) Each option in CDN resource settings. Each option added to CDN resource settings should have the following mandatory request fields: enabled, value. (see [below for nested schema](#nestedblock--options))
- `origin_group` (Number) ID of the Origins Group. Use one of your Origins Group or create a new one. You can use either 'origin' parameter or 'originGroup' in the resource definition.
- `origin_protocol` (String) This option defines the protocol that will be used by CDN servers to request content from an origin source. If not specified, it will be inherit from resource. Possible values are: HTTPS, HTTP, MATCH.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--options"></a>
### Nested Schema for `options`

Optional:

- `browser_cache_settings` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--browser_cache_settings))
- `cors` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--cors))
- `edge_cache_settings` (Block List, Max: 1) The cache expiration time for CDN servers. (see [below for nested schema](#nestedblock--options--edge_cache_settings))
- `gzip_on` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--gzip_on))
- `host_header` (Block List, Max: 1) Specify the Host header that CDN servers use when request content from an origin server. Your server must be able to process requests with the chosen header. If the option is in NULL state Host Header value is taken from the CNAME field. (see [below for nested schema](#nestedblock--options--host_header))
- `ignore_query_string` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--ignore_query_string))
- `query_params_blacklist` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--query_params_blacklist))
- `query_params_whitelist` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--query_params_whitelist))
- `redirect_http_to_https` (Block List, Max: 1) Sets redirect from HTTP protocol to HTTPS for all resource requests. (see [below for nested schema](#nestedblock--options--redirect_http_to_https))
- `rewrite` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--rewrite))
- `sni` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--sni))
- `static_headers` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--static_headers))
- `static_request_headers` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--static_request_headers))
- `tls_versions` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--tls_versions))
- `webp` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--webp))
- `websockets` (Block List, Max: 1) (see [below for nested schema](#nestedblock--options--websockets))

<a id="nestedblock--options--browser_cache_settings"></a>
### Nested Schema for `options.browser_cache_settings`

Optional:

- `enabled` (Boolean)
- `value` (String)


<a id="nestedblock--options--cors"></a>
### Nested Schema for `options.cors`

Required:

- `value` (Set of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--edge_cache_settings"></a>
### Nested Schema for `options.edge_cache_settings`

Optional:

- `custom_values` (Map of String) Caching time for a response with specific codes. These settings have a higher priority than the value field. Response code ('304', '404' for example). Use 'any' to specify caching time for all response codes. Caching time in seconds ('0s', '600s' for example). Use '0s' to disable caching for a specific response code.
- `default` (String) Content will be cached according to origin cache settings. The value applies for a response with codes 200, 201, 204, 206, 301, 302, 303, 304, 307, 308 if an origin server does not have caching HTTP headers. Responses with other codes will not be cached.
- `enabled` (Boolean)
- `value` (String) Caching time for a response with codes 200, 206, 301, 302. Responses with codes 4xx, 5xx will not be cached. Use '0s' disable to caching. Use custom_values field to specify a custom caching time for a response with specific codes.


<a id="nestedblock--options--gzip_on"></a>
### Nested Schema for `options.gzip_on`

Required:

- `value` (Boolean)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--host_header"></a>
### Nested Schema for `options.host_header`

Required:

- `value` (String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--ignore_query_string"></a>
### Nested Schema for `options.ignore_query_string`

Required:

- `value` (Boolean)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--query_params_blacklist"></a>
### Nested Schema for `options.query_params_blacklist`

Required:

- `value` (Set of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--query_params_whitelist"></a>
### Nested Schema for `options.query_params_whitelist`

Required:

- `value` (Set of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--redirect_http_to_https"></a>
### Nested Schema for `options.redirect_http_to_https`

Required:

- `value` (Boolean)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--rewrite"></a>
### Nested Schema for `options.rewrite`

Required:

- `body` (String)

Optional:

- `enabled` (Boolean)
- `flag` (String)


<a id="nestedblock--options--sni"></a>
### Nested Schema for `options.sni`

Optional:

- `custom_hostname` (String) Required to set custom hostname in case sni-type='custom'
- `enabled` (Boolean)
- `sni_type` (String) Available values 'dynamic' or 'custom'


<a id="nestedblock--options--static_headers"></a>
### Nested Schema for `options.static_headers`

Required:

- `value` (Map of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--static_request_headers"></a>
### Nested Schema for `options.static_request_headers`

Required:

- `value` (Map of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--tls_versions"></a>
### Nested Schema for `options.tls_versions`

Required:

- `value` (Set of String)

Optional:

- `enabled` (Boolean)


<a id="nestedblock--options--webp"></a>
### Nested Schema for `options.webp`

Required:

- `jpg_quality` (Number)
- `png_quality` (Number)

Optional:

- `enabled` (Boolean)
- `png_lossless` (Boolean)


<a id="nestedblock--options--websockets"></a>
### Nested Schema for `options.websockets`

Required:

- `value` (Boolean)

Optional:

- `enabled` (Boolean)


