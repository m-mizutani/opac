package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/opac"
	"github.com/m-mizutani/zlog"
	"github.com/m-mizutani/zlog/filter"
	"github.com/urfave/cli/v2"
)

var logger = zlog.New()

var (
	errInvalidConfiguration = goerr.New("invalid configuration")

	// just to control exit code
	errExitWithNonZero = goerr.New("exit with non-zero")
)

type config struct {
	BaseURL       string
	path          string
	FailDefined   bool
	FailUndefined bool
	InputData     string
	InputFile     string
	AuthBearer    string `zlog:"secret"`

	LogLevel string
}

type authClient struct {
	auth string
}

func (x *authClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+x.auth)

	return (&http.Client{}).Do(req)
}

func cmd(args []string) error {
	var cfg config

	app := &cli.App{
		Name:  "opaq",
		Usage: "Query to OPA server",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "fail-defined",
				Usage:       "exits with non-zero exit code on undefined/empty result and errors",
				Destination: &cfg.FailDefined,
			},
			&cli.BoolFlag{
				Name:        "fail-undefined",
				Usage:       "exits with non-zero exit code on defined/non-empty result and errors",
				Destination: &cfg.FailUndefined,
			},
			&cli.StringFlag{
				Name:        "url",
				Aliases:     []string{"u"},
				EnvVars:     []string{"OPAQ_URL"},
				Required:    true,
				Usage:       "base URL of OPA server",
				Destination: &cfg.BaseURL,
			},

			&cli.StringFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Usage:       "inquiry path after v1/data",
				Destination: &cfg.path,
			},
			&cli.StringFlag{
				Name:        "input-data",
				Aliases:     []string{"d"},
				Usage:       "input data (string)",
				Destination: &cfg.InputData,
			},
			&cli.StringFlag{
				Name:        "input-file",
				Aliases:     []string{"f"},
				Usage:       "input data (file)",
				Destination: &cfg.InputFile,
			},
			&cli.StringFlag{
				Name:        "auth-bearer",
				EnvVars:     []string{"OPAQ_AUTH_BEARER"},
				Usage:       "Token for Authorization Bearer of HTTP header",
				Destination: &cfg.AuthBearer,
			},

			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "logging level [debug,info,warn,error]",
				Value:       "info",
				Destination: &cfg.LogLevel,
			},
		},

		Before: func(_ *cli.Context) error {
			l, err := zlog.NewWithError(
				zlog.WithLogLevel(cfg.LogLevel),
				zlog.WithFilters(filter.Tag()),
			)
			if err != nil {
				return err
			}
			logger = l

			logger.With("config", cfg).Debug("starting")

			return nil
		},
		After: func(_ *cli.Context) error {
			logger.Debug("exiting")
			return nil
		},

		Action: func(_ *cli.Context) error {
			return handler(&cfg)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		if errors.Is(errExitWithNonZero, err) {
			return err
		}
		logger.Error(err.Error())
		logger.With("config", cfg).Err(err).Debug("error detail")
		return err
	}

	return nil
}

func handler(cfg *config) error {
	logger.With("config", cfg).Debug("Starting inquiry")

	if cfg.InputData != "" && cfg.InputFile != "" {
		return goerr.Wrap(errInvalidConfiguration, "either one of input-data and input-file is allowed")
	}

	req := &opac.DataRequest{
		Path: cfg.path,
	}

	var data []byte
	if cfg.InputData != "" {
		data = []byte(cfg.InputData)
	}
	if cfg.InputFile != "" {
		raw, err := ioutil.ReadFile(cfg.InputFile)
		if err != nil {
			return goerr.Wrap(err)
		}
		data = raw
	}

	if len(data) > 0 {
		var obj interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return goerr.Wrap(err).With("data", data)
		}
		req.Input = obj
	}

	var options []opac.Option
	if cfg.AuthBearer != "" {
		options = append(options, opac.WithHTTPClient(&authClient{
			auth: cfg.AuthBearer,
		}))
	}

	opa, err := opac.New(cfg.BaseURL, options...)
	if err != nil {
		return goerr.Wrap(err)
	}

	logger.With("req", req).Debug("Sending API request")
	ctx := context.Background()
	var out interface{}
	if err := opa.GetData(ctx, req, &out); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return goerr.Wrap(err).With("out", out)
	}
	fmt.Println(string(raw))

	if cfg.FailDefined && !isEmpty(out) {
		return errExitWithNonZero
	}
	if cfg.FailUndefined && isEmpty(out) {
		return errExitWithNonZero
	}

	return nil
}

func isEmpty(out interface{}) bool {
	if out == nil {
		return true
	}
	switch reflect.TypeOf(out).Kind() {
	case reflect.Ptr:
		return reflect.ValueOf(out).IsNil()
	case reflect.Map, reflect.Array, reflect.Slice:
		return reflect.ValueOf(out).Len() == 0
	}
	return false
}

func main() {
	if err := cmd(os.Args); err != nil {
		os.Exit(1)
	}
}
