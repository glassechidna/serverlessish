package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/glassechidna/serverlessish/lambdaruntime"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type httpResponseOutput struct {
	StatusCode        int                 `json:"statusCode"`
	StatusDescription string              `json:"statusDescription"`
	Headers           map[string]string   `json:"headers"`
	HeadersMV         map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
}

type httpRequestInput struct {
	HTTPMethod      string              `json:"httpMethod"`
	Path            string              `json:"path"`
	Headers         map[string]string   `json:"headers"`
	HeadersMV       map[string][]string `json:"multiValueHeaders"`
	Query           map[string]string   `json:"queryStringParameters"`
	QueryMV         map[string][]string `json:"multiValueQueryStringParameters"`
	Body            string              `json:"body"`
	IsBase64Encoded *bool               `json:"isBase64Encoded,omitempty"`
}

func lambdaResponseForHttpResponse(resp *http.Response) (*httpResponseOutput, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	encoded := base64.StdEncoding.EncodeToString(body)

	output := &httpResponseOutput{
		StatusCode:        resp.StatusCode,
		StatusDescription: resp.Status,
		HeadersMV:         resp.Header,
		Body:              encoded,
		IsBase64Encoded:   true,
	}

	return output, nil
}

func httpRequestForLambdaInvocation(input *lambdaruntime.FunctionNextOutput, port string) (*http.Request, bool, error) {
	base := fmt.Sprintf("http://127.0.0.1:%s", port)

	hrInput := &httpRequestInput{}
	err := json.Unmarshal(input.Body, &hrInput)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}

	var req *http.Request
	isHttpRequest := false

	if hrInput.IsBase64Encoded == nil {
		path := strings.TrimPrefix(os.Getenv("LH_INVOKE_PATH"), "/")
		if path == "" {
			path = "invoke"
		}

  		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/%s", base, path), bytes.NewReader(input.Body))
	} else {
		isHttpRequest = true
		var body io.Reader = strings.NewReader(hrInput.Body)
		if *hrInput.IsBase64Encoded {
			body = base64.NewDecoder(base64.StdEncoding, body)
		}

		q := url.Values{}
		for name, values := range hrInput.QueryMV {
			for _, value := range values {
				q.Add(name, value)
			}
		}

		if len(q) == 0 {
			for name, value := range hrInput.Query {
				q.Add(name, value)
			}
		}

		u, _ := url.Parse(base + hrInput.Path)
		u.RawQuery = q.Encode()
		req, _ = http.NewRequest(hrInput.HTTPMethod, u.String(), body)

		if len(hrInput.HeadersMV) > 0 {
			for name, values := range hrInput.HeadersMV {
				for _, value := range values {
					req.Header.Add(name, value)
				}
			}
		} else {
			for name, value := range hrInput.Headers {
				req.Header.Add(name, value)
			}
		}

		req.Host = req.Header.Get("Host")
	}

	req.Header.Set("Lambda-Runtime-Aws-Request-Id", input.RequestId)
	req.Header.Set("Lambda-Runtime-Deadline-Ms", input.DeadlineMs)
	req.Header.Set("Lambda-Runtime-Invoked-Function-Arn", input.InvokedFunctionArn)
	req.Header.Set("Lambda-Runtime-Trace-Id", input.TraceId)
	req.Header.Set("X-Amzn-Trace-Id", input.TraceId) // seems useful to add this too
	req.Header.Set("Lambda-Runtime-Client-Context", input.ClientContext)
	req.Header.Set("Lambda-Runtime-Cognito-Identity", input.CognitoIdentity)

	return req, isHttpRequest, nil
}
