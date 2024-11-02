---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_project Data Source - dependencytrack"
subcategory: ""
description: |-
  Fetch an existing Project by name and version. Requires the project to have a version defined on DependencyTrack.
---

# dependencytrack_project (Data Source)

Fetch an existing Project by name and version. Requires the project to have a version defined on DependencyTrack.

## Example Usage

```terraform
data "dependencytrack_project" "example" {
  name    = "Example"
  version = "v1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the project to find.
- `version` (String) Version of the project to find.

### Read-Only

- `id` (String) UUID of the project located.
- `properties` (Attributes List) Existing properties within the Project. (see [below for nested schema](#nestedatt--properties))

<a id="nestedatt--properties"></a>
### Nested Schema for `properties`

Read-Only:

- `description` (String) Description for the project Property.
- `group` (String) Group Name for the project Property.
- `name` (String) Property Name for the project Property.
- `type` (String) Property Type for the project Property as a string enum.
- `value` (String) Property Value for the project Property.