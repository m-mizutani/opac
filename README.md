# OPA Go Client

Unofficial [OPA](https://github.com/open-policy-agent/opa) HTTP Client in Go.

## Usage

```go
client, err := opaclient.New("http://localhost:8181")
if err != nil {
    panic(err)
}

req := opaclient.DataRequest{
    Path: "example/policy",
    Input: map[string]string{
        "user": "m-mizutani",
    },
}
resp := struct {
    Allowed bool `json:"allowed"`
}{}

if err := client.GetData(context.Background(), &req, &resp); err != nil {
    panic(err)
}

fmt.Println("allowed? =>", resp.Allowed)
```

## Testing

Run OPA server with following command.

```bash
$ opa run -s ./testdata/policy/
```

Then, run `go test`

```bash
$ env OPA_BASE_URL=http://localhost:8181 go test -v .
```

## License

MIT License
