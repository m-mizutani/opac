package opac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/zlog"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown/print"
)

// Local loads and compile local policy data and evaluate input data with them.
type Local struct {
	flies    []string
	dirs     []string
	policies map[string]string

	compiler *ast.Compiler
	query    string
	logger   *zlog.Logger
	print    io.Writer
}

// LocalOption is Option of functional option pattern for Local
type LocalOption func(x *Local)

// WithFile specifies .rego policy file path
func WithFile(filePath string) LocalOption {
	return func(x *Local) {
		x.flies = append(x.flies, filepath.Clean(filePath))
	}
}

// WithDir specifies directory path of .rego policy. Import policy files recursively.
func WithDir(dirPath string) LocalOption {
	return func(x *Local) {
		x.dirs = append(x.dirs, filepath.Clean(dirPath))
	}
}

// WithPolicyData specifies raw policy data with name. If the `name` conflicts with file path loaded by WithFile or WithDir, the policy overwrites data loaded by WithFile or WithDir.
func WithPolicyData(name, policy string) LocalOption {
	return func(x *Local) {
		x.policies[name] = policy
	}
}

// WithPackage specifies using package name. e.g. "example.my_policy"
func WithPackage(pkg string) LocalOption {
	return func(x *Local) {
		x.query = "data." + pkg
	}
}

// WithLoggingLocal enables logger for debug
func WithLoggingLocal() LocalOption {
	return func(x *Local) {
		x.logger = zlog.New(zlog.WithLogLevel("debug"))
	}
}

// WithRegoPrint enables OPA print function and output to `w`
func WithRegoPrint(w io.Writer) LocalOption {
	return func(x *Local) {
		x.print = w
	}
}

// NewLocal creates a new Local client. It requires one or more WithFile, WithDir or WithPolicyData.
func NewLocal(options ...LocalOption) (*Local, error) {
	client := &Local{
		query:    "data",
		logger:   zlog.New(),
		policies: make(map[string]string),
	}
	for _, opt := range options {
		opt(client)
	}

	policies := make(map[string]string)
	var targetFiles []string
	for _, dirPath := range client.dirs {
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return ErrReadRegoDir.Wrap(err).With("path", path)
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".rego" {
				return nil
			}

			targetFiles = append(targetFiles, path)

			return nil
		})
		if err != nil {
			return nil, goerr.Wrap(err)
		}
	}
	targetFiles = append(targetFiles, client.flies...)

	for _, filePath := range targetFiles {
		raw, err := os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, ErrReadRegoFile.Wrap(err).With("path", filePath)
		}

		policies[filePath] = string(raw)
	}

	for k, v := range client.policies {
		policies[k] = v
	}

	if len(policies) == 0 {
		return nil, goerr.Wrap(ErrNoPolicyData)
	}

	compiler, err := ast.CompileModulesWithOpt(policies, ast.CompileOpts{
		EnablePrintStatements: true,
	})
	if err != nil {
		return nil, goerr.Wrap(err)
	}
	client.compiler = compiler

	client.logger.
		With("query", client.query).
		With("loaded files", targetFiles).
		Debug("created local client")

	return client, nil
}

type printLogger struct {
	w io.Writer
}

func (x *printLogger) Print(ctx print.Context, msg string) error {
	if x.w != nil {
		fmt.Fprintf(x.w, "%s:%d %s", ctx.Location.File, ctx.Location.Row, msg)
	}
	return nil
}

// Query evaluates policy with `input` data. The result will be written to `out`. `out` must be pointer of instance.
func (x *Local) Query(ctx context.Context, input interface{}, output interface{}) error {
	x.logger.With("in", input).Debug("start Local.Query")

	q := rego.New(
		rego.Query(x.query),
		rego.PrintHook(&printLogger{
			w: x.print,
		}),
		rego.Compiler(x.compiler),
		rego.Input(input),
	)

	rs, err := q.Eval(ctx)

	if err != nil {
		return goerr.Wrap(err, "fail to eval local policy").With("input", input)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return goerr.Wrap(ErrNoEvalResult)
	}

	x.logger.With("rs", rs).Debug("got a result of rego.Eval")

	raw, err := json.Marshal(rs[0].Expressions[0].Value)
	if err != nil {
		return goerr.Wrap(err, "fail to marshal a result of rego.Eval").With("rs", rs)
	}
	if err := json.Unmarshal(raw, output); err != nil {
		return goerr.Wrap(err, "fail to unmarshal a result of rego.Eval to out").With("rs", rs)
	}

	x.logger.With("result set", rs).Debug("done Local.Query")

	return nil
}
