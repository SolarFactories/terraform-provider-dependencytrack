---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dependencytrack_team_apikey Resource - dependencytrack"
subcategory: ""
description: |-
  Manages an API Key for a Team..
---

# dependencytrack_team_apikey (Resource)

Manages an API Key for a Team..



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `team` (String) UUID for the Team for which to manage the permission.

### Read-Only

- `key` (String, Sensitive) The generated API Key for the Team.