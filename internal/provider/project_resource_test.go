package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Project"
	active = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "name", "Test_Project"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "active", "true"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "description", ""),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "version", ""),
					resource.TestCheckNoResourceAttr("dependencytrack_project.test", "parent"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "classifier", "APPLICATION"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.#", "0"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Project_With_Change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "name", "Test_Project_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "active", "true"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "description", ""),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "version", ""),
					resource.TestCheckNoResourceAttr("dependencytrack_project.test", "parent"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "classifier", "APPLICATION"),
				),
			},
		},
	})
}

func TestAccProjectNestedResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "parent" {
	name = "Parent_Project"
}
resource "dependencytrack_project" "child" {
	name = "Child_Project"
	parent = dependencytrack_project.parent.id
}
`,
				Check: resource.TestCheckResourceAttrPair(
					"dependencytrack_project.parent", "id",
					"dependencytrack_project.child", "parent",
				),
			},
			// ImportState.
			{
				ResourceName:      "dependencytrack_project.child",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "parent" {
	name = "Parent_Project"
}
resource "dependencytrack_project" "child" {
	name = "Child_Project_With_Change"
	parent = dependencytrack_project.parent.id
}
`,
				Check: resource.TestCheckResourceAttrPair(
					"dependencytrack_project.parent", "id",
					"dependencytrack_project.child", "parent",
				),
			},
		},
	})
}

func TestAccProjectIdentity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project Identity Project"
	group = "TestGroup"
	purl = "pkg:npm/namespace/name@v1.0?k=v#subpath"
	cpe = "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
	swid = "Test_SWID"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "group", "TestGroup"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "purl", "pkg:npm/namespace/name@v1.0?k=v#subpath"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "cpe", "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "swid", "Test_SWID"),
				),
			},
			// ImportState.
			{
				ResourceName:      "dependencytrack_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project Identity Project"
	group = "TestGroup With Change"
	purl = "pkg:npm/namespace/name@v1.0?k=v#subpath"
	cpe = "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"
	swid = "Test_SWID"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "group", "TestGroup With Change"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "purl", "pkg:npm/namespace/name@v1.0?k=v#subpath"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "cpe", "cpe:2.3:a:ntp:ntp:4.2.8:p3:*:*:*:*:*:*"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "swid", "Test_SWID"),
				),
			},
		},
	})
}

func TestAccProjectVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project With Version"
	version = "Test_Version"
}
data "dependencytrack_project" "data" {
	name = dependencytrack_project.test.name
	version = dependencytrack_project.test.version
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project.test", "id",
						"data.dependencytrack_project.data", "id",
					),
					resource.TestCheckResourceAttr("data.dependencytrack_project.data", "name", "Test Project With Version"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.data", "version", "Test_Version"),
				),
			},
		},
	})
}

func TestAccProjectTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project With Tags"
	version = "Test_Tags"
	tags = ["test_project_tags_tag1", "test_project_tags_tag2"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.0", "test_project_tags_tag1"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.1", "test_project_tags_tag2"),
				),
			},
			// ImportState.
			{
				ResourceName:      "dependencytrack_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project With Tags"
	version = "Test_Tags"
	tags = ["test_project_tags_tag1", "test_project_tags_tag2_with_change"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.0", "test_project_tags_tag1"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "tags.1", "test_project_tags_tag2_with_change"),
				),
			},
		},
	})
}

func TestAccProjectTagsRead(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "project1" {
	name = "Test_Project_Tags_Read_1"
	tags = ["test_project_tags_read_tag"]
}
resource "dependencytrack_project" "project2" {
	name = "Test_Project_Tags_Read_2"
	version = "v1"
}
resource "dependencytrack_tag_projects" "projects" {
	tag = "test_project_tags_read_tag"
	projects = [
		dependencytrack_project.project1.id,
		dependencytrack_project.project2.id,
	]
}
data "dependencytrack_project" "project2" {
	name = dependencytrack_project.project2.name
	version = dependencytrack_project.project2.version
	depends_on = [dependencytrack_tag_projects.projects]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// State obtained prior to it having the tag applied, so expect to not be aware of tags.
					resource.TestCheckResourceAttr("dependencytrack_project.project2", "tags.#", "0"),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.projects", "projects.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.projects", "projects.1",
						"dependencytrack_project.project2", "id",
					),
					resource.TestCheckResourceAttr("data.dependencytrack_project.project2", "tags.#", "1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.project2", "tags.0", "test_project_tags_read_tag"),
				),
			},
			// ImportState.
			{
				ResourceName:      "dependencytrack_project.project2",
				ImportState:       true,
				ImportStateVerify: true,
				// Configuration of tags on `project2`, validated by `data.dependencytrack_project.project2`.
				ImportStateVerifyIgnore: []string{"tags"},
			},
			// Update and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "project1" {
	name = "Test_Project_Tags_Read_1"
	tags = ["test_project_tags_read_tag"]
}
resource "dependencytrack_project" "project2" {
	name = "Test_Project_Tags_Read_2"
	version = "v1"
}
resource "dependencytrack_tag_projects" "projects" {
	tag = "test_project_tags_read_tag"
	projects = [
		dependencytrack_project.project1.id,
		dependencytrack_project.project2.id,
	]
}
data "dependencytrack_project" "project2" {
	name = dependencytrack_project.project2.name
	version = dependencytrack_project.project2.version
	depends_on = [dependencytrack_tag_projects.projects]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.project2", "tags.#", "1"),
					resource.TestCheckResourceAttr("dependencytrack_project.project2", "tags.0", "test_project_tags_read_tag"),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.projects", "projects.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.projects", "projects.1",
						"dependencytrack_project.project2", "id",
					),
					resource.TestCheckResourceAttr("data.dependencytrack_project.project2", "tags.#", "1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.project2", "tags.0", "test_project_tags_read_tag"),
				),
			},
		},
	})
}

// API 4.13+.
func TestAccProjectCollection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project Collection Direct"
	version = "Test_Collection"
	collection = {
		logic = "AGGREGATE_DIRECT_CHILDREN"
	}
}
resource "dependencytrack_project" "test2" {
	name = "Test Project Collection Tag"
	version = "Test_Collection_Tag"
	collection = {
		logic = "AGGREGATE_DIRECT_CHILDREN_WITH_TAG",
		tag = "some_tag",
	}
}
resource "dependencytrack_project" "test3" {
	name = "Test Project Collection None"
	version = "Test_Collection_None"
	collection = {
		logic = "NONE"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "collection.logic", "AGGREGATE_DIRECT_CHILDREN"),
					resource.TestCheckResourceAttr("dependencytrack_project.test2", "collection.logic", "AGGREGATE_DIRECT_CHILDREN_WITH_TAG"),
					resource.TestCheckResourceAttr("dependencytrack_project.test2", "collection.tag", "some_tag"),
					resource.TestCheckResourceAttr("dependencytrack_project.test3", "collection.logic", "NONE"),
				),
			},
			// ImportState.
			{
				ResourceName:      "dependencytrack_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "dependencytrack_project.test2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "dependencytrack_project.test3",
				ImportState:       true,
				ImportStateVerify: true,
				// Importing when the collection is "NONE" results in a nil value being stored.
				ImportStateVerifyIgnore: []string{"collection"},
			},
			// Update and Read.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test Project Collection Direct With Change"
	version = "Test_Collection"
	collection = {
		logic = "AGGREGATE_DIRECT_CHILDREN"
	}
}
resource "dependencytrack_project" "test2" {
	name = "Test Project Collection Tag"
	version = "Test_Collection_Tag"
	collection = {
		logic = "AGGREGATE_DIRECT_CHILDREN_WITH_TAG",
		tag = "some_tag_with_change",
	}
}
resource "dependencytrack_project" "test3" {
	name = "Test Project Collection None With Change"
	version = "Test_Collection_None"
	collection = {
		logic = "NONE"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project.test", "collection.logic", "AGGREGATE_DIRECT_CHILDREN"),
					resource.TestCheckResourceAttr("dependencytrack_project.test2", "collection.logic", "AGGREGATE_DIRECT_CHILDREN_WITH_TAG"),
					resource.TestCheckResourceAttr("dependencytrack_project.test2", "collection.tag", "some_tag_with_change"),
					resource.TestCheckResourceAttr("dependencytrack_project.test3", "collection.logic", "NONE"),
				),
			},
		},
	})
}
