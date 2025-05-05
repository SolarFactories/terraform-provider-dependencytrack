package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccPolicyProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "project" {
	name = "Test_Project"
}
resource "dependencytrack_policy" "test" {
	name = "Test_Policy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy_project" "test" {
	policy = dependencytrack_policy.test.id
	project = dependencytrack_project.project.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_project.test", "policy",
						"dependencytrack_policy.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_project.test", "project",
						"dependencytrack_project.project", "id",
					),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "project" {
	name = "Test_Project"
}
resource "dependencytrack_policy" "test" {
	name = "Test_Policy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "test_1" {
	name = "Test_Policy_1"
	operator = "ALL"
	violation = "FAIL"
}
resource "dependencytrack_policy_project" "test" {
	policy = dependencytrack_policy.test_1.id
	project = dependencytrack_project.project.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_project.test", "policy",
						"dependencytrack_policy.test_1", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_project.test", "project",
						"dependencytrack_project.project", "id",
					),
				),
			},
		},
	})
}
