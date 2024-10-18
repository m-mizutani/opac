package opac

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

type fileSource struct {
	cfg      *config
	paths    []string
	compiler *ast.Compiler
}

// AnnotationSet implements Source.
func (f *fileSource) AnnotationSet() *ast.AnnotationSet {
	return f.compiler.GetAnnotationSet()
}

// Configure implements Source.
func (f *fileSource) Configure(cfg *config) error {
	policies := map[string]string{}
	for _, dirPath := range f.paths {
		cfg.logger.Debug("Importing policy files/dirs", "path", dirPath)
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".rego" {
				return nil
			}

			fpath := filepath.Clean(path)
			cfg.logger.Debug("Reading policy file", "path", fpath)
			raw, err := os.ReadFile(fpath)
			if err != nil {
				return fmt.Errorf("failed to read policy file: %w", err)
			}

			policies[fpath] = string(raw)

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	}

	if len(policies) == 0 {
		return ErrNoPolicyData
	}
	cfg.logger.Debug("Policy files are loaded", "file count", len(policies))

	compiler, err := ast.CompileModulesWithOpt(policies, ast.CompileOpts{
		EnablePrintStatements: true,
		ParserOptions: ast.ParserOptions{
			ProcessAnnotation: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to compile policy: %w", err)
	}

	f.compiler = compiler
	f.cfg = cfg
	return nil
}

// Query implements Source.
func (f *fileSource) Query(ctx context.Context, query string, input any, output any, opt queryOptions) error {
	return queryLocal(ctx, f.cfg, f.compiler, query, input, output, opt)
}

var _ Source = (*fileSource)(nil)

// Files is an option to specify the file path to read rego files. If path is a directory, it reads all files with the .rego extension in the directory.
//
// Example:
//
//	client, err := opac.New(opac.Files(
//		"path/to/policy_file.rego",
//		"path/to/policy_dir",
//	))
func Files(paths ...string) Source {
	return &fileSource{
		paths: paths,
	}
}

// Data is an option to specify the policy data as a map. The key can be set any value as file path and the value is the policy content.
//
// Example:
//
//	data := `package system.authz
//	  allow {
//	    input.user == "admin"
//	  }
//	`
//	 policies := map[string]string{
//	   "policy1.rego": data,
//	 }
//
//	client, err := opac.New(opac.Data(policies))
func Data(policies map[string]string) Source {
	return &dataSource{
		policies: policies,
	}
}

type dataSource struct {
	cfg      *config
	policies map[string]string
	compiler *ast.Compiler
}

// AnnotationSet implements Source.
func (d *dataSource) AnnotationSet() *ast.AnnotationSet {
	return d.compiler.GetAnnotationSet()
}

// Configure implements Source.
func (d *dataSource) Configure(cfg *config) error {
	if len(d.policies) == 0 {
		return ErrNoPolicyData
	}
	cfg.logger.Debug("Policy data are loaded", "data count", len(d.policies))

	compiler, err := ast.CompileModulesWithOpt(d.policies, ast.CompileOpts{
		EnablePrintStatements: true,
		ParserOptions: ast.ParserOptions{
			ProcessAnnotation: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to compile policy: %w", err)
	}

	d.compiler = compiler
	d.cfg = cfg

	return nil
}

// Query implements Source.
func (d *dataSource) Query(ctx context.Context, query string, input any, output any, opt queryOptions) error {
	return queryLocal(ctx, d.cfg, d.compiler, query, input, output, opt)
}

var _ Source = (*dataSource)(nil)

func queryLocal(ctx context.Context, cfg *config, compiler *ast.Compiler, query string, input, output any, opt queryOptions) error {
	options := []func(r *rego.Rego){
		rego.Query(query),
		rego.Compiler(compiler),
		rego.Input(input),
	}

	if opt.printHook != nil {
		cfg.logger.Debug("Setting print hook")
		options = append(options, rego.PrintHook(opt.printHook))
	}

	q := rego.New(options...)

	rs, err := q.Eval(ctx)
	if err != nil {
		return fmt.Errorf("failed to evaluate query: %w", err)
	}

	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return ErrNoEvalResult
	}

	raw, err := json.Marshal(rs[0].Expressions[0].Value)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(raw, output); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}
