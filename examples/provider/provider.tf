provider "dependencytrack" {
  host    = "http://localhost:8081"
  key     = "OS_ENV"
  headers = [{ name = "HEADER-NAME", value = "HEADER-VALUE" }]
}
