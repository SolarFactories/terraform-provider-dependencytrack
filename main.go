// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

// Terraform Provider Communication for DependencyTrack Provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-dependencytrack/internal/provider"
)

var (
	// These will be set by the goreleaser configuration.
	// Set to appropriate values for the compiled binary.
	version = "dev"

	// `goreleaser` can pass other info to the main package, e.g. commit hash.
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false,
		"set to true to run the provider with support for debuggers like delve",
	)
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/solarfactories/dependencytrack",
		Debug:   debug,
	}

	err := providerserver.Serve(
		context.Background(),
		provider.New(version),
		opts,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
}
