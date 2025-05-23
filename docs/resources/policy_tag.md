---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_policy_tag Resource - dependencytrack"
subcategory: ""
description: |-
  Manages an application of a Policy to a Tag.
---

# dependencytrack_policy_tag (Resource)

Manages an application of a Policy to a Tag.

## Example Usage

```terraform
resource "dependencytrack_policy" "example" {
  name      = "Sample Policy"
  operator  = "ALL"
  violation = "ERROR"
}

resource "dependencytrack_policy_tag" "example" {
  policy = dependencytrack_policy.example.id
  tag    = "DemoTag"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `policy` (String) UUID for the Policy to apply to the Tag.
- `tag` (String) Name of the Tag to which to apply Policy.
