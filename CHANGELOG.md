## 1.1.0
FEATURES:
	- `dependencytrack_project_property` Resource, to manage a project property.
	- `dependencytrack_project_property` DataSource, to retrieve a singular property.

ISSUES:
	- Unable to delete project property within DependencyTrack, when using `dependencytrack_project_property` resource.

FIXES:
	- Removed erroneous configuration of attributes on `dependencytrack_project_property` from being labelled as not changing.

## 1.0.0

FEATURES:
	- Provider authentication via API Key, optionally reading from environment variable.
	- `dependencytrack_project` Resource, for Projects, able to set minimal functionality.
	- `dependencytrack_project` DataSource, to identify from a Project name and version, able to access properties.
