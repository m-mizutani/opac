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

type Local struct {
	compiler *ast.Compiler
	query    string
	logger   *zlog.Logger
	print    io.Writer
}

type LocalOption func(x *Local)

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

func NewLocal(path string, options ...LocalOption) (*Local, error) {
	policies := make(map[string]string)
	var loadedFiles []string
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return ErrReadRegoDir.Wrap(err).With("path", path)
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".rego" {
			return nil
		}

		raw, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return ErrReadRegoFile.Wrap(err).With("path", path)
		}

		policies[path] = string(raw)
		loadedFiles = append(loadedFiles, path)

		return nil
	})
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	compiler, err := ast.CompileModulesWithOpt(policies, ast.CompileOpts{
		EnablePrintStatements: true,
	})
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	client := &Local{
		compiler: compiler,
		query:    "data",
		logger:   zlog.New(),
	}
	for _, opt := range options {
		opt(client)
	}
	client.logger.
		With("query", client.query).
		With("files", loadedFiles).
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
