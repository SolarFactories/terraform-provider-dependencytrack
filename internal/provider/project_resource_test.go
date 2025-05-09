package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
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
