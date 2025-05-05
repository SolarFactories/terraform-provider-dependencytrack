package provider

import (
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	providerConfig = func() string {
		option := os.Getenv("DEPENDENCYTRACK_TEST_PROVIDER")
		if option == "rootCA" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8082"
				key = "OS_ENV"
				root_ca = "` + strings.ReplaceAll(string(rootCa), "\n", "\\n") + `"
			}`
		}
		if option == "mtls" {
			return `provider "dependencytrack" {
				host = "http://localhost:8083"
				key = "OS_ENV"
				mtls = {
					key_path = "/opt/client_key.pem",
					cert_path = "/opt/client_cert.pem",
				}
			}`
		}
		if option == "rootCA+mtls" {
			rootCa, err := os.ReadFile("/opt/server_cert.pem")
			if err != nil {
				panic("Root CA file is unable to be read: " + err.Error())
			}
			return `provider "dependencytrack" {
				host = "https://localhost:8084"
				key = "OS_ENV"
				root_ca = "` + strings.ReplaceAll(string(rootCa), "\n", "\\n") + `"
				mtls = {
					key_path = "/opt/client_key.pem",
					cert_path = "/opt/client_cert.pem",
				}
			}`
		}
		return `provider "dependencytrack" {
			host = "http://localhost:8081"
			key = "OS_ENV"
		}`
	}()

	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(New("test")()),
	}
)
