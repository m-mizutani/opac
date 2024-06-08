# opac: Rego policy inquiry library with OPA

Unofficial Rego evaluation API for OPA server, local Rego file and in-memory Rego data.

## Motivation

[Rego](https://www.openpolicyagent.org/docs/latest/policy-language) is a versatile policy language, and the official documentation provides various methods for evaluating Rego policies. There are three primary ways to evaluate policies programmatically:

- Querying the OPA server
- Using local policy files
- Utilizing in-memory policy data (e.g., data from environment variables)

A software developer can choose the most suitable method based on their specific requirements. However, in many cases, end users also want to select the evaluation method depending on the runtime environment. Therefore, a unified policy evaluation approach can be beneficial for developers integrating Rego into their applications.

The `opac` library offers an abstracted API to evaluate Rego policies using an OPA server, local policy files, or in-memory text data. This allows developers to easily implement a mechanism to switch between evaluation methods based on the options chosen by end users.

## Example

### Query with local policy file(s)

```go
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
```

### Query to OPA server

```go
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
```

## Arguments

### Sources

`Source` specifies the source of the Rego policy data.

- `Files`: Read policies from local files. It can specify multiple files. If a directory is specified, it will be searched recursively.
- `Data`: Read policies from in-memory data.
- `Remote`: Use policies by inquiring the OPA server.

### Options

- `WithPrintHook`: Print the evaluation result to the standard output. It can be used for `Files` and `Data` sources.

## License

Apache License 2.0