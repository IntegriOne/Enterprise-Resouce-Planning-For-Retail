// Binary healthcheck melakukan probe kesiapan container Gateway terhadap endpoint tunggal `/api/graphql`.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type healthGraphQLResponse struct {
	Data struct {
		Health string `json:"health"`
	} `json:"data"`
	Errors []any `json:"errors"`
}

func main() {
	port := strings.TrimSpace(os.Getenv("GATEWAY_HTTP_PORT"))
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("http://127.0.0.1:%s/api/graphql?query=%s", port, "%7Bhealth%7D")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		os.Exit(1)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}

	var decoded healthGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		os.Exit(1)
	}
	if len(decoded.Errors) > 0 || decoded.Data.Health != "ok" {
		os.Exit(1)
	}
}
