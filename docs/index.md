---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack Provider"
subcategory: ""
description: |-
  Interact with DependencyTrack.
---

# dependencytrack Provider

Interact with DependencyTrack.

## Example Usage

```terraform
provider "dependencytrack" {
  host    = "http://localhost:8081"
  key     = "OS_ENV"
  headers = [{ name = "HEADER-NAME", value = "HEADER-VALUE" }]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `host` (String) URI for DependencyTrack API.
- `key` (String, Sensitive) API Key for authentication to DependencyTrack. Must have permissions for all attempted actions. Set to 'OS_ENV' to read from DEPENDENCYTRACK_API_KEY environment variable.

### Optional

- `headers` (Attributes List) Add additional headers to client API requests. Useful for proxy authentication. (see [below for nested schema](#nestedatt--headers))

<a id="nestedatt--headers"></a>
### Nested Schema for `headers`

Required:

- `name` (String) Name of the header to specify.
- `value` (String) Value of the header to specify.
