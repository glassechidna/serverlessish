package lambdaruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type LambdaRuntime struct {
	Client *http.Client
}

func New(client *http.Client) *LambdaRuntime {
	return &LambdaRuntime{Client: client}
}

type ExtensionRegisterOutput struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
	Identifier      string
}

func (r *LambdaRuntime) ExtensionRegister(ctx context.Context, events ...string) (*ExtensionRegisterOutput, error) {
	c := http.DefaultClient
	if r != nil {
		c = r.Client
	}

	fullname, _ := os.Executable()
	name := filepath.Base(fullname)

	reqBody, _ := json.Marshal(map[string]interface{}{"events": events})
	u := os.ExpandEnv("http://${AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension/register")
	req, _ := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(reqBody))
	req.Header.Set("Lambda-Extension-Name", name)

	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output := ExtensionRegisterOutput{}
	err = json.Unmarshal(respBody, &output)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output.Identifier = resp.Header.Get("Lambda-Extension-Identifier")
	return &output, nil
}

type ExtensionNextOutput struct {
	EventType          string `json:"eventType"`
	DeadlineMs         int64  `json:"deadlineMs"`
	RequestID          string `json:"requestId"`
	InvokedFunctionArn string `json:"invokedFunctionArn"`
	Tracing            struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"tracing"`
}

func (r *LambdaRuntime) ExtensionNext(ctx context.Context, identifier string) (*ExtensionNextOutput, error) {
	c := http.DefaultClient
	if r != nil {
		c = r.Client
	}

	u := os.ExpandEnv("http://${AWS_LAMBDA_RUNTIME_API}/2020-01-01/extension/event/next")
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("Lambda-Extension-Identifier", identifier)

	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output := ExtensionNextOutput{}
	err = json.Unmarshal(respBody, &output)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &output, nil
}

type FunctionNextOutput struct {
	RequestId          string
	DeadlineMs         string
	InvokedFunctionArn string
	TraceId            string
	ClientContext      string
	CognitoIdentity    string
	Body               []byte
}

func (r *LambdaRuntime) FunctionNext(ctx context.Context) (*FunctionNextOutput, error) {
	c := http.DefaultClient
	if r != nil {
		c = r.Client
	}

	u := os.ExpandEnv("http://${AWS_LAMBDA_RUNTIME_API}/2018-06-01/runtime/invocation/next")
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	requestId := resp.Header.Get("Lambda-Runtime-Aws-Request-Id")
	deadline := resp.Header.Get("Lambda-Runtime-Deadline-Ms")
	invokedFunctionArn := resp.Header.Get("Lambda-Runtime-Invoked-Function-Arn")
	traceId := resp.Header.Get("Lambda-Runtime-Trace-Id")
	clientContext := resp.Header.Get("Lambda-Runtime-Client-Context")
	cognitoIdentity := resp.Header.Get("Lambda-Runtime-Cognito-Identity")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &FunctionNextOutput{
		RequestId:          requestId,
		DeadlineMs:         deadline,
		InvokedFunctionArn: invokedFunctionArn,
		TraceId:            traceId,
		ClientContext:      clientContext,
		CognitoIdentity:    cognitoIdentity,
		Body:               body,
	}, nil
}

func (r *LambdaRuntime) FunctionResponse(ctx context.Context, requestId string, body []byte) error {
	c := http.DefaultClient
	if r != nil {
		c = r.Client
	}

	u := fmt.Sprintf("http://%s/2018-06-01/runtime/invocation/%s/response", os.Getenv("AWS_LAMBDA_RUNTIME_API"), requestId)
	req, _ := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	_, err := c.Do(req)
	return errors.WithStack(err)
}
