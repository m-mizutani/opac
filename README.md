# opac: OPA/Rego inquiry library [![Test](https://github.com/m-mizutani/opac/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/opac/actions/workflows/test.yml) [![Vuln scan](https://github.com/m-mizutani/opac/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/opac/actions/workflows/trivy.yml) [![Sec Scan](https://github.com/m-mizutani/opac/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/opac/actions/workflows/gosec.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/opac.svg)](https://pkg.go.dev/github.com/m-mizutani/opac)

Unofficial Rego evaluation API for OPA server, local Rego file and in-memory Rego data.

## Motivation

[Rego](https://www.openpolicyagent.org/docs/latest/policy-language) is general policy language and various evaluation methods for Rego are provided by official. In programing way, there are three major to evaluate policy.

- Inquiry to OPA server
- Use local policy file(s)
- Use in-memory policy data (e.g. from environment variable)

A software developer can choose a appropriate way according to own requirements. However, in many case, an end user also want to choose evaluation method by each runtime environment. Therefore, unified policy evaluation method may be useful for developer to integrate with Rego.

`opac` provides abstracted API to evaluate Rego with OPA server, local policy file and in-memory text data. A developer can easily implement switching evaluation method mechanism by option chosen by end user.

## Example

### Query to OPA server

[source code](./examples/remote/)

```go
func main() {
	client, err := opac.NewRemote("https://opa-server-h6tk4k5hyq-an.a.run.app/v1/data/example")
	if err != nil {
		panic(err.Error())
	}

	input := struct {
		Color string `json:"color"`
	}{
		Color: "blue",
	}
	output := struct {
		Allow bool `json:"allow"`
	}{}

	if err := client.Query(context.Background(), input, &output); err != nil {
		panic(err.Error())
	}

	fmt.Println("result:", output)
}
```

### Query with local policy file(s)

[source code](./examples/local/)

```go
func main() {
	client, err := opac.NewLocal(
		opac.WithFile("./examples/local/policy.rego"),
		opac.WithPackage("example"),
	)
	if err != nil {
		panic(err.Error())
	}

	input := struct {
		Color string `json:"color"`
	}{
		Color: "blue",
	}
	output := struct {
		Allow bool `json:"allow"`
	}{}

	if err := client.Query(context.Background(), input, &output); err != nil {
		panic(err.Error())
	}

	fmt.Println("result:", output)
}
```

### Test with mock

Your package code
```go
package mock

import (
	"context"

	"github.com/m-mizutani/opac"
)

type Foo struct {
	client opac.Client
}

type Input struct{ User string }
type Result struct{ Allow bool }

func New(url string) *Foo {
	client, err := opac.NewRemote(url)
	if err != nil {
		panic(err)
	}

	return &Foo{
		client: client,
	}
}

func (x *Foo) IsAllow(user string) bool {
	input := &Input{User: user}
	var result Result
	if err := x.client.Query(context.Background(), input, &result); err != nil {
		panic(err)
	}

	return result.Allow
}
```

Then, create [export_test.go](./examples/mock/export_test.go) as following.

```go
package mock

import "github.com/m-mizutani/opac"

func NewWithMock(f opac.MockFunc) *Foo {
	return &Foo{
		client: opac.NewMock(f),
	}
}
```

After that, you can write [Foo's test](./examples/mock/main_test.go) as following.

```go
func TestWithMock(t *testing.T) {
	foo := mock.NewWithMock(func(input interface{}) (interface{}, error) {
		in, ok := input.(*mock.Input)
		require.True(t, ok)
		return &mock.Result{Allow: in.User == "blue"}, nil
	})

	assert.True(t, foo.IsAllow("blue"))
	assert.False(t, foo.IsAllow("orange"))
}
```

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