package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

/*

curl --request POST \
    --url https://api.parallel.ai/v1/search \
    --header 'Content-Type: application/json' \
    --header 'x-api-key: <api-key>' \
    --data '{
    "objective": "Find latest information about Parallel Web Systems. Focus on new product releases, benchmarks, or company announcements.",
    "search_queries": ["Parallel Web Systems products", "Parallel Web Systems announcements"]
}'
*/

var httpClient = http.Client{Timeout: 20 * time.Second}

func search(term SearchInput, ctx context.Context) (string, error) {

	apiKey := os.Getenv("PARALLEL_API_KEY")

	if apiKey == "" {
		return "", errors.New("api key is not available, make sure PARALLEL_API_KEY is set")
	}
	data, err := json.Marshal(term)

	if err != nil {
		return "", fmt.Errorf("error occured when formatting value: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.parallel.ai/v1/search", bytes.NewReader(data))

	if err != nil {
		return "", fmt.Errorf("Error when creating request: %w", err)
	}

	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	req = req.WithContext(ctx)

	
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response returned with status code %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error occured when reading response body %w", err)
	}

	return string(responseBody), nil
}

// Directly from https://docs.parallel.ai/api-reference/search/search
type SearchInput struct {
	Objective     string   `json:"objective" jsonschema:"description=Natural-language description of the underlying question or goal driving the search. Used together with search_queries to focus results on the most relevant content. Should be self-contained with enough context to understand the intent of the search."`
	SearchQueries []string `json:"search_queries" jsonschema:"description=Concise keyword search queries, 3-6 words each. At least one query is required, provide 2-3 for best results. Used together with objective to focus results on the most relevant content."`
}


func FromString(s string) (SearchInput, error) {
	var searchInput SearchInput 
	err := json.Unmarshal([]byte(s), &searchInput)
	return searchInput, err
}