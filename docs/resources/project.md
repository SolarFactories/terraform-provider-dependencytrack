---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_project Resource - dependencytrack"
subcategory: ""
description: |-
  Manages a Project.
---

# dependencytrack_project (Resource)

Manages a Project.

## Example Usage

```terraform
resource "dependencytrack_project" "example" {
  name        = "Example"
  description = "Example project"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the Project.

### Optional

- `active` (Boolean) Whether the Project is active. Defaults to true.
- `classifier` (String) Classifier of the Project. Defaults to APPLICATION. See DependencyTrack for valid options.
- `cpe` (String) Common Platform Enumeration of the Project. Standardised format v2.2 / v2.3 from MITRE / NIST.
- `description` (String) Description of the Project.
- `group` (String) Namespace / group / vendor of the Project.
- `parent` (String) UUID of a parent project, to allow for nesting. Available in API 4.7+.
- `purl` (String) Package URL of the Project. MUST be in standardised format to be saved. See DependencyTrack for format.
- `swid` (String) SWID Tag ID. ISO/IEC 19770-2:2015.
- `tags` (List of String) Tags to assign to a project. If unset, retains existing tags on project. If set, and `dependencytrack_tag_projects` is used with any of the tags, it must include this project's `id`.
- `version` (String) Version of the project.

### Read-Only

- `id` (String) UUID for the Project as generated by DependencyTrack.

## Import

Import is supported using the following syntax:

```shell
terraform import dependencytrack_project.example c82d6f01-a7a4-41d6-9b03-4f06497f575b
```
