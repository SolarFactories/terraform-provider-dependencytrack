---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_config_property Resource - dependencytrack"
subcategory: ""
description: |-
  Manages a Config Property.
---

# dependencytrack_config_property (Resource)

Manages a Config Property.

## Example Usage

```terraform
resource "dependencytrack_config_property" "example" {
  group = "general"
  name  = "base.url"
  value = "http://localhost:8000"
  type  = "STRING"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group` (String) Group name of the Config Property.
- `name` (String) Property name of the Config Property.
- `type` (String) Type of the Config Property. See DependencyTrack for valid enum values.
- `value` (String) Value of the Config Property.

### Read-Only

- `description` (String) Description of the Config Property.
- `id` (String) ID used by provider. Has no meaning to DependencyTrack.

## Import

Import is supported using the following syntax:

```shell
terraform import dependencytrack_config_property.example general/base.url
```
