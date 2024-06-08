package opac_test

import (
	"context"
	"fmt"
	"os"

	"github.com/m-mizutani/opac"
)

func ExampleFiles() {
	client, err := opac.New(opac.Files("testdata/examples/authz.rego"))
	if err != nil {
		panic(err)
	}

	input := map[string]string{
		"user": "bob",
		"role": "admin",
	}
	var output struct {
		Allow bool `json:"allow"`
	}
	ctx := context.Background()
	if err := client.Query(ctx, "data.authz", input, &output); err != nil {
		panic(err)
	}
	fmt.Println("allow =>", output.Allow)
	//Output: allow => true
}

func ExampleData() {
	data := `package system.authz
	  allow {
	    input.user == "admin"
	  }
	`
	policies := map[string]string{
		"policy1.rego": data,
	}

	client, err := opac.New(opac.Data(policies))
	if err != nil {
		panic(err)
	}

	input := map[string]string{
		"user": "admin",
	}
	var output struct {
		Allow bool `json:"allow"`
	}
	ctx := context.Background()
	if err := client.Query(ctx, "data.system.authz", input, &output); err != nil {
		panic(err)
	}
	fmt.Println("allow =>", output.Allow)
	//Output: allow => true
}

func ExampleRemote() {
	// This test requires OPA server running with testdata/examples/authz.rego
	// You can run OPA server with following command:
	//
	// opa run -s testdata/examples/authz.rego
	//
	// And set OPA_SERVER_URL environment variable to the server URL.
	// For example: export OPA_SERVER_URL=http://localhost:8181/v1
	opaServerURL, ok := os.LookupEnv("OPA_SERVER_URL")
	if !ok {
		fmt.Println("allow => true") // dummy output
		return
	}

	client, err := opac.New(opac.Remote(opaServerURL))
	if err != nil {
		panic(err)
	}

	input := map[string]string{
		"user": "alice",
	}
	var output struct {
		Allow bool `json:"allow"`
	}
	ctx := context.Background()
	if err := client.Query(ctx, "data.authz", input, &output); err != nil {
		panic(err)
	}
	fmt.Println("allow =>", output.Allow)
	//Output: allow => true
}
