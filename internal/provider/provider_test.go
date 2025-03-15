package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
)

var (
	providerConfig = func() string {
		option := os.Getenv("DEPENDENCYTRACK_TEST_PROVIDER")
		if option == "rootCA" {
			rootCa, err := os.ReadFile("/opt/root_ca")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8082"
				key = "OS_ENV"
				root_ca = "` + string(rootCa) + `"
			}`
		}
		if option == "mtls" {
			return `provider "dependencytrack" {
				host = "https://localhost:8083"
				key = "OS_ENV"
				mtls = {
					key_path = "/opt/tls_key.pem",
					cert_path = "/opt/tls_cert.pem",
				}
			}`
		}
		if option == "rootCA+mtls" {
			rootCa, err := os.ReadFile("/opt/root_ca")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8084"
				key = "OS_ENV"
				root_ca = "` + string(rootCa) + `"
				mtls = {
					key_path = "/opt/tls_key.pem",
					cert_path = "/opt/tls_cert.pem",
				}
			}`
		}
		return `provider "dependencytrack" {
			host = "http://localhost:8081"
			key = "OS_ENV"
		}`
	}()
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(New("test")()),
	}
)
