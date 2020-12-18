package main

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
)

type verboseTransport struct {
	http.RoundTripper
}

func (l *verboseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(request, true)
	fmt.Println(string(dump))
	resp, err := l.RoundTripper.RoundTrip(request)
	if err != nil {
		fmt.Printf("ERROR: %+v\n", err)
		return nil, errors.WithStack(err)
	}
	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(dump))
	return resp, err
}
