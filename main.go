package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"code.citrite.net/csvc/spa-sdk/terraform-provider-spa/internal/provider"
)

//go:generate terraform fmt -recursive ./examples/

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/citrix/spa",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
