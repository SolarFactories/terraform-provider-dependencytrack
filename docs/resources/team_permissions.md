---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_team_permissions Resource - dependencytrack"
subcategory: ""
description: |-
  Manages the attachment of Permissions to a Team. Conflicts with dependencytrack_team_permission.
---

# dependencytrack_team_permissions (Resource)

Manages the attachment of Permissions to a Team. Conflicts with `dependencytrack_team_permission`.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `permissions` (List of String) Alphabetically sorted Permissions for team. Conflicts with `dependencytrack_team_permission`. See DependencyTrack for allowed values.
- `team` (String) UUID for the Team for which to manage the permissions.
