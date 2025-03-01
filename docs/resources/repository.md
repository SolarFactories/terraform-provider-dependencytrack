---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_repository Resource - dependencytrack"
subcategory: ""
description: |-
  Manages a Repository.
---

# dependencytrack_repository (Resource)

Manages a Repository.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `enabled` (Boolean) Whether the Repository Enabled.
- `identifier` (String) Identifier of the Repository.
- `internal` (Boolean) Whether the Repository is Internal.
- `password` (String, Sensitive) Password to use for Authentication to Repository.
- `type` (String) Type of the Repository. See DependencyTrack for valid enum values.
- `url` (String) URL of the Repository.
- `username` (String) Username to use for Authentication to Repository.

### Optional

- `precedence` (Number) Precedence / Resolution Order of the Repository.

### Read-Only

- `id` (String) UUID for the Repository as generated by DependencyTrack.
