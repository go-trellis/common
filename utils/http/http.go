/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps http.Client with additional utilities
type Client struct {
	*http.Client
	BaseURL    string
	Headers    map[string]string
	Timeout    time.Duration
	RetryCount int
}

// Config configures HTTP client
type Config struct {
	BaseURL    string
	Headers    map[string]string
	Timeout    time.Duration
	RetryCount int
	Transport  http.RoundTripper
}

// NewClient creates a new HTTP client
func NewClient(cfg Config) *Client {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}
	if cfg.Timeout == 0 {
		client.Timeout = 30 * time.Second
	}
	if cfg.Transport != nil {
		client.Transport = cfg.Transport
	}

	return &Client{
		Client:     client,
		BaseURL:    cfg.BaseURL,
		Headers:    cfg.Headers,
		Timeout:    client.Timeout,
		RetryCount: cfg.RetryCount,
	}
}

// Request represents HTTP request options
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Context context.Context
}

// Do executes HTTP request
func (c *Client) Do(req Request) (*http.Response, error) {
	var body io.Reader
	if req.Body != nil {
		switch v := req.Body.(type) {
		case []byte:
			body = bytes.NewReader(v)
		case string:
			body = bytes.NewReader([]byte(v))
		case io.Reader:
			body = v
		default:
			data, err := json.Marshal(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(data)
		}
	}

	url := req.URL
	if c.BaseURL != "" && !isAbsoluteURL(url) {
		url = c.BaseURL + url
	}

	httpReq, err := http.NewRequestWithContext(getContext(req.Context), req.Method, url, body)
	if err != nil {
		return nil, err
	}

	// Set default headers
	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set request headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set content-type if body is JSON
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		if _, ok := req.Body.(string); !ok {
			httpReq.Header.Set("Content-Type", "application/json")
		}
	}

	return c.Client.Do(httpReq)
}

// Get performs GET request
func (c *Client) Get(url string, headers map[string]string) (*http.Response, error) {
	return c.Do(Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headers,
	})
}

// Post performs POST request
func (c *Client) Post(url string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.Do(Request{
		Method:  http.MethodPost,
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Put performs PUT request
func (c *Client) Put(url string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.Do(Request{
		Method:  http.MethodPut,
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Delete performs DELETE request
func (c *Client) Delete(url string, headers map[string]string) (*http.Response, error) {
	return c.Do(Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	})
}

// GetJSON performs GET request and unmarshals JSON response
func (c *Client) GetJSON(url string, headers map[string]string, result interface{}) error {
	resp, err := c.Get(url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// PostJSON performs POST request with JSON body and unmarshals JSON response
func (c *Client) PostJSON(url string, body interface{}, headers map[string]string, result interface{}) error {
	resp, err := c.Post(url, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// ReadBody reads response body and returns as bytes
func ReadBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("response or body is nil")
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ReadBodyString reads response body and returns as string
func ReadBodyString(resp *http.Response) (string, error) {
	data, err := ReadBody(resp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsSuccess checks if status code indicates success
func IsSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsRedirect checks if status code indicates redirect
func IsRedirect(statusCode int) bool {
	return statusCode >= 300 && statusCode < 400
}

// IsClientError checks if status code indicates client error
func IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerError checks if status code indicates server error
func IsServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}

func isAbsoluteURL(url string) bool {
	return len(url) > 0 && (url[0] == '/' || len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://"))
}

func getContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
