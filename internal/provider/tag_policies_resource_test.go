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
				// Use Project to create the Tag, due to `dependencytrack_tag` being 4.13+.
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Policies_Resource_Project"
	tags = ["test_tag_policies_tag"]
}
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
	tag = "test_tag_policies_tag"
	policies = [
		dependencytrack_policy.test.id,
		dependencytrack_policy.test2.id,
	]
	depends_on = [dependencytrack_project.test]
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
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "tag", "test_tag_policies_tag"),
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
resource "dependencytrack_project" "test" {
	name = "Tag_Polcies_Resource_Project"
	tags = ["test_tag_policies_tag"]
}
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
	tag = "test_tag_policies_tag"
	policies = [
		dependencytrack_policy.test.id,
		dependencytrack_policy.test2.id,
	]
	depends_on = [dependencytrack_project.test]
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
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "tag", "test_tag_policies_tag"),
				),
			},
		},
	})
}

func TestAccTagPoliciesResourcePoliciesUnordered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				// Use Project to create the Tag, due to `dependencytrack_tag` being 4.13+.
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Policies_Resources_Unordered"
	tags = ["test_tag_policies_unordered"]
}
resource "dependencytrack_policy" "a" {
	name = "A"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "z" {
	name = "z"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_tag_policies" "test" {
	tag = "test_tag_policies_unordered"
	policies = [
		dependencytrack_policy.z.id,
		dependencytrack_policy.a.id,
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "policies.#", "2"),
					resource.TestCheckResourceAttrPair("dependencytrack_tag_policies.test", "policies.0", "dependencytrack_policy.z", "id"),
					resource.TestCheckResourceAttrPair("dependencytrack_tag_policies.test", "policies.1", "dependencytrack_policy.a", "id"),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Policies_Resources_Unordered"
	tags = ["test_tag_policies_unordered"]
}
resource "dependencytrack_policy" "a" {
	name = "A"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "z" {
	name = "Z"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "b" {
	name = "B"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_tag_policies" "test" {
	tag = "test_tag_policies_unordered"
	policies = [
		dependencytrack_policy.z.id,
		dependencytrack_policy.a.id,
		dependencytrack_policy.b.id
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_policies.test", "policies.#", "3"),
					resource.TestCheckResourceAttrPair("dependencytrack_tag_policies.test", "policies.0", "dependencytrack_policy.z", "id"),
					resource.TestCheckResourceAttrPair("dependencytrack_tag_policies.test", "policies.1", "dependencytrack_policy.a", "id"),
					resource.TestCheckResourceAttrPair("dependencytrack_tag_policies.test", "policies.2", "dependencytrack_policy.b", "id"),
				),
			},
		},
	})
}
