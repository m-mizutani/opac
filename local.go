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

// Files is an option to specify the file path to read rego files. If path is a directory, it reads all files with the .rego extension in the directory.
//
// Example:
//
//	client, err := opac.New(opac.Files(
//		"path/to/policy_file.rego",
//		"path/to/policy_dir",
//	))
func Files(paths ...string) Source {
	return func(cfg *config) (queryFunc, error) {
		policies := map[string]string{}
		for _, dirPath := range paths {
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
				return nil, fmt.Errorf("failed to walk directory: %w", err)
			}
		}

		if len(policies) == 0 {
			return nil, ErrNoPolicyData
		}
		cfg.logger.Debug("Policy files are loaded", "file count", len(policies))

		compiler, err := ast.CompileModulesWithOpt(policies, ast.CompileOpts{
			EnablePrintStatements: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to compile policy: %w", err)
		}

		return func(ctx context.Context, query string, input, output any, opt *queryOptions) error {
			return queryLocal(ctx, cfg, compiler, query, input, output, opt)
		}, nil
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
	return func(cfg *config) (queryFunc, error) {
		if len(policies) == 0 {
			return nil, ErrNoPolicyData
		}
		cfg.logger.Debug("Policy data are loaded", "data count", len(policies))

		compiler, err := ast.CompileModulesWithOpt(policies, ast.CompileOpts{
			EnablePrintStatements: true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to compile policy: %w", err)
		}

		return func(ctx context.Context, query string, input, output any, opt *queryOptions) error {
			return queryLocal(ctx, cfg, compiler, query, input, output, opt)
		}, nil
	}
}

func queryLocal(ctx context.Context, cfg *config, compiler *ast.Compiler, query string, input, output any, opt *queryOptions) error {
	options := []func(r *rego.Rego){
		rego.Query(query),
		rego.Compiler(compiler),
		rego.Input(input),
	}

	if opt != nil && opt.printHook != nil {
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
