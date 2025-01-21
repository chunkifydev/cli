package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/chunkifydev/cli/pkg/config"
	"github.com/chunkifydev/cli/pkg/formatter"
	"github.com/chunkifydev/cli/pkg/styles"
)

type Request struct {
	Config      *config.Config
	Path        string
	Method      string
	QueryParams url.Values
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
	if apiReq.Config.Debug {
		fmt.Println(styles.Debug.Render(apiReq.String()))
	}

	var zero T
	body := &bytes.Buffer{}

	if apiReq.Method == "POST" || apiReq.Method == "PATCH" || apiReq.Method == "PUT" {
		json, err := json.Marshal(apiReq.Body)
		if err != nil {
			return zero, fmt.Errorf("could not marshal request body: %w", err)
		}

		body = bytes.NewBuffer(json)

		if apiReq.Config.Debug {
			fmt.Println(styles.Debug.Render(string(json) + "\n"))
		}
	}

	req, err := http.NewRequest(apiReq.Method, apiUrl(apiReq.Config.ApiEndpoint, apiReq.Path), body)
	if err != nil {
		return zero, err
	}

	if len(apiReq.QueryParams) > 0 {
		req.URL.RawQuery = apiReq.QueryParams.Encode()
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "chunkify-cli")

	// for these endpoints, we need the account API key
	if strings.HasPrefix(apiReq.Path, "/api/tokens") || strings.HasPrefix(apiReq.Path, "/api/projects") {
		req.Header.Set("Authorization", "Bearer "+apiReq.Config.AccountToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiReq.Config.ProjectToken)
	}

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("could not read response body: %w", err)
	}

	if apiReq.Config.Debug {
		fmt.Println(formatter.HttpCode(resp.StatusCode))
		for h, v := range resp.Header {
			if strings.HasPrefix(h, "X-Amz") {
				continue
			}
			fmt.Println(styles.Debug.Render(fmt.Sprintf("%s: %s", h, v[0])))
		}
		fmt.Println(string(respBody))
	}

	if resp.StatusCode >= http.StatusBadRequest {
		if len(respBody) == 0 {
			return zero, fmt.Errorf("response status code %d", resp.StatusCode)
		}

		var errorResponse ApiError
		if err := json.Unmarshal(respBody, &errorResponse); err != nil {
			return zero, fmt.Errorf("response status code %d", resp.StatusCode)
		}
		return zero, fmt.Errorf("%s (%d)", errorResponse.Error.Message, errorResponse.Error.Code)
	}

	var apiResponse Response[T]
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &apiResponse); err != nil {
			return zero, fmt.Errorf("could not parse the response: %w", err)
		}
	}

	return apiResponse.Data, nil
}
