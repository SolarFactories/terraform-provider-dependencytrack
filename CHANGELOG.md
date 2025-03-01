## 1.5.0

#### FEATURES
- `dependencytrack_repository` Resource, to manage an external source repository.
- Added HTTP interception patching for select API requests, for which the SDK does not provide a working function.

#### MISC
- Added automated testing against Terraform `1.11.x`.
- Swapped `golangci-lint` rules from `enable` specific set, to `enable-all`, with specific `disable` to increase range of linters used.
- Reviewed linting rules, actioning, or identifying where not actioning.
- Added named import for `github.com/DependencyTrack/client-go` to resolve typecheck errors due to updated golang version.
- Removed secondary `Get` request when updating `dependencytrack_project`, instead using the return type of `Update` function.
- Added property tests for configuring properties within `dependencytrack_project`.

#### FIXES
- Fixed inability to delete a `dependencytrack_project_property`, as raised in `1.1.0`.
- Marked `type` as requiring replace when updated within `dependencytrack_project_property`.
- Fixed `active` not defaulting to `true` within `dependencytrack_project`.

#### DEPENDENCIES
- `actions/download-artifact` `4.1.8` -> `4.1.9`
- `github.com/hashicorp/terraform-plugin-framework` `v1.13.0` -> `v1.14.1`

## 1.4.0

#### FEATURES
- `dependencytrack_config_property` Resource, to manage a config property.
- `dependencytrack_config_property` DataSource, to retrieve a config property.
- `dependencytrack_config_properties` Resource, to manage multiple config properties more efficiently.

#### MISC
- Added automated testing against Terraform `1.10.x`.
- Disabled CDKTF binding generation, while it is not fully featured.
- Removed workflow to mark inactive issues as resolved.

#### DEPENDENCIES
- `golang/golangci-lint-action` `6.3.3` -> `6.4.0`

## 1.3.3

#### DEPENDENCIES
- `github.com/DependencyTrack/client-go` `0.14.0` -> `0.15.0`
- `golang/golangci-lint-action` `6.3.0` -> `6.3.3`
- `goreleaser/goreleaser-action` `6.1.0` -> `6.2.1`

## 1.3.2

#### DEPENDENCIES
- `actions/setup-go` `5.2.0` -> `5.3.0`
- `github.com/hashicorp/terraform-plugin-go` `0.25.0` -> `0.26.0`
- `actions/setup-node` `4.1.0` -> `4.2.0`
- `golangci/golangci-lint-action` `6.2.0` -> `6.3.0`

## 1.3.1

#### DEPENDENCIES
- `golang.org/x/net` `0.28.0` -> `0.33.0`
- `golangci/golangci-lint-action` `6.1.1` -> `6.2.0`

##### /tools
- `golang.org.net` `0.23.0` -> `0.33.0`

## 1.3.0

#### FEATURES
- `dependencytrack_team` Resource, to manage a team.
- `dependencytrack_team` DataSource, to retrieve a team.
- `dependencytrack_team_apikey` Resource, to manage an API Key for a team.
- `dependencytrack_team_permission` Resource, to manage the permissions of a team.

#### DEPENDENCIES
- `DependencyTrack/client-go` `v0.13.0` -> `v0.14.0`
- `hashicorp/terraform-plugin-testing` `v1.10.0` -> `v1.11.0`

## 1.2.0

#### FEATURES
- `dependencytrack` Provider - Added options for setting additional custom headers.

## 1.1.0

#### FEATURES
- `dependencytrack_project_property` Resource, to manage a project property.
- `dependencytrack_project_property` DataSource, to retrieve a singular property.

#### ISSUES
- [Fixed in `1.5.0`] Unable to delete project property within DependencyTrack, when using `dependencytrack_project_property` resource.
- [Fixed in `1.5.0`] Updating `type` on `dependencytrack_project_property` does not recreate the resource, which is required to change the `type`.

## 1.0.0

#### FEATURES
- Provider authentication via API Key, optionally reading from environment variable.
- `dependencytrack_project` Resource, for Projects, able to set minimal functionality.
- `dependencytrack_project` DataSource, to identify from a Project name and version, able to access properties.

#### ISSUES
- [Fixed in `1.5.0`] `dependencytrack_project.active` does not default to `true`.
