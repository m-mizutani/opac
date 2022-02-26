package opac

import (
	"context"
	"encoding/json"
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
}

type LocalOption func(x *Local)

func WithPackage(pkg string) LocalOption {
	return func(x *Local) {
		x.query = "data." + pkg
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

	compiler, err := ast.CompileModules(policies)
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

	return client, nil
}

type printLogger struct {
	logger *zlog.Logger
}

func (x *printLogger) Print(ctx print.Context, msg string) error {
	x.logger.With("msg", msg).With("ctx", ctx).Debug("print")
	return nil
}

func (x *Local) Query(ctx context.Context, in interface{}, out interface{}) error {
	x.logger.With("in", in).Trace("start Local.Eval")
	rego := rego.New(
		rego.Query(x.query),
		rego.PrintHook(&printLogger{
			logger: x.logger,
		}),
		rego.Compiler(x.compiler),
		rego.Input(in),
	)

	rs, err := rego.Eval(ctx)

	if err != nil {
		return goerr.Wrap(err, "fail to eval local policy").With("input", in)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return goerr.Wrap(ErrNoEvalResult)
	}

	x.logger.With("rs", rs).Trace("got a result of rego.Eval")

	raw, err := json.Marshal(rs[0].Expressions[0].Value)
	if err != nil {
		return goerr.Wrap(err, "fail to marshal a result of rego.Eval").With("rs", rs)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return goerr.Wrap(err, "fail to unmarshal a result of rego.Eval to out").With("rs", rs)
	}

	x.logger.With("rs", rs).Trace("done Local.Eval")

	return nil
}
