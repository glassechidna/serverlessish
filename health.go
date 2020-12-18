package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func waitForHealthy(ctx context.Context, port string) {
	path := strings.TrimPrefix(os.Getenv("LH_HEALTHCHECK_PATH"), "/")
	if path == "" {
		path = "ping"
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/%s", port, path)

	waitUntil(ctx, func() bool {
		resp, err := http.Get(url)
		return err == nil && resp != nil && resp.StatusCode == 200
	})
}

func waitUntil(ctx context.Context, condition func() bool) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		fmt.Println("sleep")
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
