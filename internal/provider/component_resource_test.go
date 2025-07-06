package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccComponentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_Project"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Component_Component"
	version = "v1.0"
	hashes = {}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
				/*					resource.TestCheckResourceAttrSet("dependencytrack_acl_mapping.test", "id"),
									resource.TestCheckResourceAttrPair(
										"dependencytrack_project.test", "id",
										"dependencytrack_acl_mapping.test", "project",
									),
									resource.TestCheckResourceAttrPair(
										"dependencytrack_team.test", "id",
										"dependencytrack_acl_mapping.test", "team",
									),*/
				),
			},
			// ImportState testing.
			/*{
				ResourceName:      "dependencytrack_component.test",
				ImportState:       true,
				ImportStateVerify: true,
			},*/
			// Update and Read testing.
			/*		{
									Config: providerConfig + `
					resource "dependencytrack_project" "test" {
						name = "Test_ACL_Project"
					}
					resource "dependencytrack_team" "test" {
						name = "Test_ACL_Team"
					}
					resource "dependencytrack_team" "test2" {
						name = "Test_ACL_Team_2"
					}
					resource "dependencytrack_acl_mapping" "test" {
						project = dependencytrack_project.test.id
						team = dependencytrack_team.test2.id
					}
					`,
									Check: resource.ComposeAggregateTestCheckFunc(
										resource.TestCheckResourceAttrSet("dependencytrack_acl_mapping.test", "id"),
										resource.TestCheckResourceAttrPair(
											"dependencytrack_project.test", "id",
											"dependencytrack_acl_mapping.test", "project",
										),
										resource.TestCheckResourceAttrPair(
											"dependencytrack_team.test2", "id",
											"dependencytrack_acl_mapping.test", "team",
										),
									),
								},*/
		},
	})
}
