provider "dependencytrack" {
  host    = "http://localhost:8081"
  key     = "OS_ENV"
  headers = [{ name = "HEADER-NAME", value = "HEADER-VALUE" }]
}

provider "dependencytrack" {
  host    = "https://localhost:8081"
  key     = "OS_ENV"
  root_ca = "-----BEGIN CERTIFICATE-----\n...\n...\n...\n-----END CERTIFICATE-----"
  mtls = {
    key_path  = "/opt/client_key.pem"
    cert_path = "/opt/client_cert.pem"
  }
}
