package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/serverlist": mockServerListSingleJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	if err := client.Ping(); err != nil {
		t.Fatalf("expected Ping() to succeed, got error: %v", err)
	}
}

func TestPingInvalidAPIKey(t *testing.T) {
	ts := newMockServer(t, http.StatusUnauthorized, map[string]string{
		"/serverlist": mockErrInvalidAPIKeyJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	err := client.Ping()
	if err == nil {
		t.Fatalf("expected Ping() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 5122: invalid apikey"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestPingAPIKeyOutOfScope(t *testing.T) {
	ts := newMockServer(t, http.StatusUnauthorized, map[string]string{
		"/serverlist": mockErrOutOfScopeJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	err := client.Ping()
	if err == nil {
		t.Fatalf("expected Ping() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 5120: out of scope"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestVersion(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/version": mockVersionJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	versionInfo, err := client.Version()
	if err != nil {
		t.Fatalf("expected Version() to succeed, got error: %v", err)
	}

	if versionInfo.Build != 1779874471 {
		t.Errorf("expected build to be 1779874471, got %d", versionInfo.Build)
	}

	if versionInfo.Platform != "Linux" {
		t.Errorf("expected platform to be 'Linux', got %q", versionInfo.Platform)
	}

	if versionInfo.Version != "3.13.8" {
		t.Errorf("expected version to be '3.13.8', got %q", versionInfo.Version)
	}
}

func TestVirtualServerList(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/serverlist": mockServerListMultiJSON,
	})

	client := NewClient(ts.URL, "test-api-key")

	servers, err := client.VirtualServerList()
	if err != nil {
		t.Fatalf("expected VirtualServerList() to succeed, got error: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 entries in server list, got %d", len(servers))
	}

	if servers[0].ID != 1 {
		t.Errorf("expected server[0].ID to be 1, got %d", servers[0].ID)
	}
	if servers[0].Name != "First server" {
		t.Errorf("expected server[0].Name to be 'First server', got %q", servers[0].Name)
	}
	if servers[0].Status != "online" {
		t.Errorf("expected server[0].Status to be 'online', got %q", servers[0].Status)
	}

	if servers[2].ID != 6 {
		t.Errorf("expected server[2].ID to be 6, got %d", servers[2].ID)
	}
	if servers[2].Status != "offline" {
		t.Errorf("expected server[2].Status to be 'offline', got %q", servers[2].Status)
	}
}

func TestVirtualServerInfo(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/1/serverinfo": mockServer1InfoJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	info, err := client.VirtualServerInfo(1)
	if err != nil {
		t.Fatalf("expected VirtualServerInfo() to succeed, got error: %v", err)
	}

	if info.Uptime != 688902 {
		t.Errorf("expected Uptime to be 688902, got %d", info.Uptime)
	}

	if info.TotalPing != 0.000 {
		t.Errorf("expected TotalPing to be 0.000, got %.4f", info.TotalPing)
	}

	if info.BytesSentTotal != 2187625219 {
		t.Errorf("expected BytesSentTotal to be 2187625219, got %d", info.BytesSentTotal)
	}
}

func TestVirtualServerInfoBadID(t *testing.T) {
	ts := newMockServer(t, http.StatusBadRequest, map[string]string{
		"/999/serverinfo": mockServerInfoBadIDJSON,
	})

	client := NewClient(ts.URL, "test-api-key")

	_, err := client.VirtualServerInfo(999)
	if err == nil {
		t.Fatalf("expected VirtualServerInfo() error, got nil: %v", err)
	}

	expectedErr := "webquery api error 7: canceled"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestTrimURLTrailingSlash(t *testing.T) {
	client := NewClient("http://127.0.0.1:10080/", "test-api-key")

	if client.baseURL != "http://127.0.0.1:10080" {
		t.Errorf("expected baseURL to be 'http://127.0.0.1:10080', got %q", client.baseURL)
	}
}

func TestTrimApiKeyWhitespace(t *testing.T) {
	client := NewClient("http://127.0.0.1:10080", " test-api-key\n\r")

	if client.apiKey != "test-api-key" {
		t.Errorf("expected apiKey to be 'test-api-key', got %q", client.apiKey)
	}
}

func TestBadJSONResponse(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/version": mockMalformedJSON,
	})

	client := NewClient(ts.URL, "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatalf("expected Version() error, got nil: %v", err)
	}

	expectedErr := "failed to decode JSON response"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestServerUnreachable(t *testing.T) {
	client := NewClient("http://127.0.0.1:10090", "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}

	expectedErr := "unable to send HTTP request"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

func TestEmptyBody(t *testing.T) {
	ts := newMockServer(t, http.StatusOK, map[string]string{
		"/version": mockEmptyBodyJSON,
	})
	client := NewClient(ts.URL, "test-api-key")

	_, err := client.Version()
	if err == nil {
		t.Fatalf("expected Version() error, got nil: %v", err)
	}

	expectedErr := "expected exactly 1 version entry, got 0"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %q", expectedErr, err.Error())
	}
}
