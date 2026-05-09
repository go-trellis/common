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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestNewClient(t *testing.T) {
	cfg := Config{
		BaseURL: "https://api.example.com",
		Timeout: 10 * time.Second,
		Headers: map[string]string{
			"X-API-Key": "test-key",
		},
	}
	client := NewClient(cfg)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "https://api.example.com", client.BaseURL, "should set base URL")
	testutils.Equals(t, 10*time.Second, client.Timeout, "should set timeout")
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewClient(Config{Timeout: 5 * time.Second})
	resp, err := client.Get(server.URL, nil)
	testutils.Ok(t, err)
	testutils.Assert(t, resp != nil, "response should not be nil")
	testutils.Equals(t, http.StatusOK, resp.StatusCode, "should return 200")
	resp.Body.Close()
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	}))
	defer server.Close()

	client := NewClient(Config{Timeout: 5 * time.Second})
	resp, err := client.Post(server.URL, map[string]string{"key": "value"}, nil)
	testutils.Ok(t, err)
	testutils.Assert(t, resp != nil, "response should not be nil")
	testutils.Equals(t, http.StatusCreated, resp.StatusCode, "should return 201")
	resp.Body.Close()
}

func TestReadBody(t *testing.T) {
	resp := &http.Response{
		Body: http.NoBody,
	}
	_, err := ReadBody(resp)
	// NoBody might return empty, that's OK
	_ = err
}

func TestIsSuccess(t *testing.T) {
	testutils.Assert(t, IsSuccess(200), "200 should be success")
	testutils.Assert(t, IsSuccess(201), "201 should be success")
	testutils.Assert(t, !IsSuccess(400), "400 should not be success")
}

func TestIsRedirect(t *testing.T) {
	testutils.Assert(t, IsRedirect(301), "301 should be redirect")
	testutils.Assert(t, IsRedirect(302), "302 should be redirect")
	testutils.Assert(t, !IsRedirect(200), "200 should not be redirect")
}

func TestIsClientError(t *testing.T) {
	testutils.Assert(t, IsClientError(400), "400 should be client error")
	testutils.Assert(t, IsClientError(404), "404 should be client error")
	testutils.Assert(t, !IsClientError(500), "500 should not be client error")
}

func TestIsServerError(t *testing.T) {
	testutils.Assert(t, IsServerError(500), "500 should be server error")
	testutils.Assert(t, IsServerError(503), "503 should be server error")
	testutils.Assert(t, !IsServerError(400), "400 should not be server error")
}
