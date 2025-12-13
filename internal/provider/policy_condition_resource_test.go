package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicyConditionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy_condition" "test" {
	policy = dependencytrack_policy.test.id
	subject = "AGE"
	operator = "NUMERIC_GREATER_THAN"
	value = "P1Y"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_policy_condition.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_condition.test", "policy",
						"dependencytrack_policy.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "subject", "AGE"),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "operator", "NUMERIC_GREATER_THAN"),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "value", "P1Y"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_policy_condition.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy_condition" "test" {
	policy = dependencytrack_policy.test.id
	subject = "AGE"
	operator = "NUMERIC_GREATER_THAN"
	value = "P2Y"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_policy_condition.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_condition.test", "policy",
						"dependencytrack_policy.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "subject", "AGE"),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "operator", "NUMERIC_GREATER_THAN"),
					resource.TestCheckResourceAttr("dependencytrack_policy_condition.test", "value", "P2Y"),
				),
			},
		},
	})
}
