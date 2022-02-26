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

type policyData struct {
	Name   string
	Policy string
}

type Local struct {
	flies    []string
	dirs     []string
	policies map[string]string

	compiler *ast.Compiler
	query    string
	logger   *zlog.Logger
	print    io.Writer
}

type LocalOption func(x *Local)

func WithFile(filePath string) LocalOption {
	return func(x *Local) {
		x.flies = append(x.flies, filepath.Clean(filePath))
	}
}

func WithDir(dirPath string) LocalOption {
	return func(x *Local) {
		x.dirs = append(x.dirs, filepath.Clean(dirPath))
	}
}

func WithPolicy(name, policy string) LocalOption {
	return func(x *Local) {
		x.policies[name] = policy
	}
}

func WithPackage(pkg string) LocalOption {
	return func(x *Local) {
		x.query = "data." + pkg
	}
}

func EnableLocalLogging() LocalOption {
	return func(x *Local) {
		x.logger = zlog.New(zlog.WithLogLevel("debug"))
	}
}

func WithRegoPrint(w io.Writer) LocalOption {
	return func(x *Local) {
		x.print = w
	}
}

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

func (x *Local) Query(ctx context.Context, in interface{}, out interface{}) error {
	x.logger.With("in", in).Debug("start Local.Query")

	q := rego.New(
		rego.Query(x.query),
		rego.PrintHook(&printLogger{
			w: x.print,
		}),
		rego.Compiler(x.compiler),
		rego.Input(in),
	)

	rs, err := q.Eval(ctx)

	if err != nil {
		return goerr.Wrap(err, "fail to eval local policy").With("input", in)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return goerr.Wrap(ErrNoEvalResult)
	}

	x.logger.With("rs", rs).Debug("got a result of rego.Eval")

	raw, err := json.Marshal(rs[0].Expressions[0].Value)
	if err != nil {
		return goerr.Wrap(err, "fail to marshal a result of rego.Eval").With("rs", rs)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return goerr.Wrap(err, "fail to unmarshal a result of rego.Eval to out").With("rs", rs)
	}

	x.logger.With("result set", rs).Debug("done Local.Query")

	return nil
}
