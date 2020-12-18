package main

import (
	"context"
	"encoding/json"
	"github.com/glassechidna/serverlessish/lambdaruntime"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()

	c := &http.Client{}
	if os.Getenv("LH_VERBOSE") != "" {
		c.Transport = &verboseTransport{http.DefaultTransport}
	}

	runtime := lambdaruntime.New(c)

	register, err := runtime.ExtensionRegister(ctx, "INVOKE")
	if err != nil {
		panic(err)
	}
	identifier := register.Identifier

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		for {
			_, err = runtime.ExtensionNext(ctx, identifier)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	waitForHealthy(ctx, port)

	group.Go(func() error {
		for {
			next, err := runtime.FunctionNext(ctx)
			if err != nil {
				return errors.WithStack(err)
			}

			req, isHttpRequest, err := httpRequestForLambdaInvocation(next, port)
			if err != nil {
				return errors.WithStack(err)
			}

			funcResp, err := c.Do(req)
			if err != nil {
				return errors.WithStack(err)
			}

			var responseBody []byte

			if isHttpRequest {
				lambdaResp, err := lambdaResponseForHttpResponse(funcResp)
				if err != nil {
				    return errors.WithStack(err)
				}

				responseBody, err = json.Marshal(lambdaResp)
				if err != nil {
				    return errors.WithStack(err)
				}
			} else {
				responseBody, err = ioutil.ReadAll(funcResp.Body)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			err = runtime.FunctionResponse(ctx, next.RequestId, responseBody)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	})

	err = group.Wait()
	if err != nil {
		panic(err)
	}
}
