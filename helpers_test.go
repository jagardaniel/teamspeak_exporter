package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockServer(t *testing.T, statusCode int, responses map[string]string) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify header for api-key is set for the tests
		if r.Header.Get("x-api-key") == "" {
			t.Error("expected x-api-key header to be set, got empty string")
		}

		body, ok := responses[r.URL.Path]
		if !ok {
			t.Errorf("unexpected path request %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Server", "TeamSpeak Server 3.13.8")
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, body)
	}))

	t.Cleanup(ts.Close)

	return ts
}

// Embedded JSON fixtures

//go:embed testdata/version.json
var mockVersionJSON string

//go:embed testdata/serverlist_single.json
var mockServerListSingleJSON string

//go:embed testdata/serverlist_multi.json
var mockServerListMultiJSON string

//go:embed testdata/serverinfo_1.json
var mockServer1InfoJSON string

//go:embed testdata/serverinfo_2.json
var mockServer2InfoJSON string

//go:embed testdata/api_key_invalid.json
var mockErrInvalidAPIKeyJSON string

//go:embed testdata/api_key_out_of_scope.json
var mockErrOutOfScopeJSON string

//go:embed testdata/empty_body.json
var mockEmptyBodyJSON string

//go:embed testdata/serverinfo_bad_id.json
var mockServerInfoBadIDJSON string

//go:embed testdata/malformed.json
var mockMalformedJSON string
