package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagPoliciesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "A_TagPoliciesPolicy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "test2" {
	name = "Z_TagPoliciesPolicy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_tag_policies" "test" {
	tag = "testtag"
	policies = [
		dependencytrack_policy.test.id,
		dependencytrack_policy.test2.id,
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "policies.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_policies.test", "policies.0",
						"dependencytrack_policy.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_policies.test", "policies.1",
						"dependencytrack_policy.test2", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "tag", "testtag"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_tag_policies.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "A_TagPoliciesPolicy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "test2" {
	name = "Z_TagPoliciesPolicy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_tag_policies" "test" {
	tag = "testtag"
	policies = [
		dependencytrack_policy.test.id,
		dependencytrack_policy.test2.id,
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "policies.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_policies.test", "policies.0",
						"dependencytrack_policy.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_policies.test", "policies.1",
						"dependencytrack_policy.test2", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "tag", "testtag"),
				),
			},
		},
	})
}
