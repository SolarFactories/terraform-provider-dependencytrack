package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

func TestAccPolicyConditionResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Policy Condition.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_PolicyCondition_236"
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
			},
			// Duplicate reference to Policy Condition.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_PolicyCondition_236"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy_condition" "test" {
	policy = dependencytrack_policy.test.id
	subject = "AGE"
	operator = "NUMERIC_GREATER_THAN"
	value = "P2Y"
}
resource "dependencytrack_policy_condition" "test2" {
	policy = dependencytrack_policy.test.id
	subject = "AGE"
	operator = "NUMERIC_GREATER_THAN"
	value = "P2Y"
}
import {
	to = dependencytrack_policy_condition.test2
	id = dependencytrack_policy_condition.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy_condition.test", "id",
						"dependencytrack_policy_condition.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_PolicyCondition_236"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy_condition" "test2" {
	policy = dependencytrack_policy.test.id
	subject = "AGE"
	operator = "NUMERIC_GREATER_THAN"
	value = "P2Y"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_policy_condition.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
