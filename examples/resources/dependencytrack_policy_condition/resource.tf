resource "dependencytrack_policy" "example" {
  name      = "Sample Policy"
  operator  = "ALL"
  violation = "ERROR"
}

resource "dependencytrack_policy_condition" "example" {
  policy   = dependencytrack_policy.example.id
  subject  = "AGE"
  operator = "NUMERIC_GREATER_THAN"
  value    = "P1Y"
}
