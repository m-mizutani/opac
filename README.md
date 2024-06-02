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

### Query to OPA server

### Query with local policy file(s)



## Options

### for `NewRemote`

- `WithHTTPClient`: Replace `http.DefaultClient` with own `HTTPClient` instance.
- `WithHTTPHeader`: Add HTTP header. It can be added multiply.
- `WithLoggingRemote`: Enable debug logging

### for `NewLocal`

One ore more `WithFile`, `WithDir` or `WithPolicyData` is required.

- `WithFile`: Specify a policy file
- `WithDir`: Specify a policy file directory (search recursively)
- `WithPolicyData`: Specify a policy data
- `WithPackage`: Specify package name like "example.my_policy"
- `WithLoggingLocal`: Enable debug logging
- `WithRegoPrint`: Output `print()` result to `io.Writer`

## License

Apache License 2.0