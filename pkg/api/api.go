package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/level63/cli/pkg/config"
)

type QueryParams map[string]string

type Request struct {
	Config      *config.Config
	Path        string
	Method      string
	QueryParams QueryParams
	Body        any
}

type Response[T any] struct {
	Data T `json:"data"`
}

type ApiError struct {
	Error ApiErrorInfo `json:"error"`
}

type ApiErrorInfo struct {
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type EmptyResponse struct{}

func (r Request) String() string {
	return fmt.Sprintf("%s %s\nQuery: %v\n", r.Method, apiUrl(r.Config.ApiEndpoint, r.Path), r.QueryParams)
}

func apiUrl(baseUrl, path string) string {
	return fmt.Sprintf("%s%s", baseUrl, path)
}

func ApiRequest[T any](apiReq Request) (T, error) {
	var zero T
	body := &bytes.Buffer{}

	if apiReq.Method == "POST" || apiReq.Method == "PATCH" || apiReq.Method == "PUT" {
		json, err := json.Marshal(apiReq.Body)
		if err != nil {
			return zero, fmt.Errorf("could not marshal request body: %w", err)
		}

		body = bytes.NewBuffer(json)
	}

	req, err := http.NewRequest(apiReq.Method, apiUrl(apiReq.Config.ApiEndpoint, apiReq.Path), body)
	if err != nil {
		return zero, err
	}

	if len(apiReq.QueryParams) > 0 {
		q := url.Values{}
		for k, v := range apiReq.QueryParams {
			q.Add(k, v)
		}

		req.URL.RawQuery = q.Encode()
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "level63-cli")

	// for these endpoints, we need the account API key
	if strings.HasPrefix(apiReq.Path, "/api/projects") || strings.HasPrefix(apiReq.Path, "/api/storages") {
		req.Header.Set("Authorization", "Bearer "+apiReq.Config.AccountApiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiReq.Config.ProjectApiKey)
	}

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		fullBody, err := io.ReadAll(resp.Body)
		if err != nil || len(fullBody) == 0 {
			return zero, fmt.Errorf("response status code %d", resp.StatusCode)
		}

		var errorResponse ApiError
		if err := json.Unmarshal(fullBody, &errorResponse); err != nil {
			return zero, fmt.Errorf("response status code %d", resp.StatusCode)
		}
		return zero, fmt.Errorf("%s (%d)", errorResponse.Error.Message, errorResponse.Error.Code)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("could not read response body: %w", err)
	}

	var apiResponse Response[T]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &apiResponse); err != nil {
			return zero, fmt.Errorf("could not parse the response: %w", err)
		}
	}

	return apiResponse.Data, nil
}
