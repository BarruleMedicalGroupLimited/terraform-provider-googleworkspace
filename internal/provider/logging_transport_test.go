// Copyright (c) Barrule Medical Group Limited
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"strings"
	"testing"
)

const testAccessTokenSecret = "ya29.super-secret-value"

func TestObfuscateValues_ScrubsAccessToken(t *testing.T) {
	m := map[string]interface{}{
		"accessToken": testAccessTokenSecret,
	}

	result := obfuscateValues(m)

	if got := result["accessToken"]; got != "********" {
		t.Fatalf("expected accessToken to be scrubbed, got %q", got)
	}
	if strings.Contains(result["accessToken"].(string), testAccessTokenSecret) {
		t.Fatalf("scrubbed accessToken still contains the secret value")
	}
}

func TestObfuscateValues_LeavesOtherFieldsUntouched(t *testing.T) {
	m := map[string]interface{}{
		"accessToken":  testAccessTokenSecret,
		"primaryEmail": "user@example.com",
		"kind":         "admin#directory#user",
	}

	result := obfuscateValues(m)

	if got := result["primaryEmail"]; got != "user@example.com" {
		t.Fatalf("expected primaryEmail to be untouched, got %q", got)
	}
	if got := result["kind"]; got != "admin#directory#user" {
		t.Fatalf("expected kind to be untouched, got %q", got)
	}
}

func TestObfuscateValues_NoScrubbableFieldPresent(t *testing.T) {
	m := map[string]interface{}{
		"primaryEmail": "user@example.com",
	}

	result := obfuscateValues(m)

	if len(result) != 1 {
		t.Fatalf("expected map to be unchanged, got %v", result)
	}
	if _, ok := result["accessToken"]; ok {
		t.Fatalf("did not expect an accessToken key to be added")
	}
}

func TestObfuscateValues_ScrubsNestedMap(t *testing.T) {
	m := map[string]interface{}{
		"credentials": map[string]interface{}{
			"accessToken": testAccessTokenSecret,
			"tokenType":   "Bearer",
		},
	}

	result := obfuscateValues(m)

	creds := result["credentials"].(map[string]interface{})
	if got := creds["accessToken"]; got != "********" {
		t.Fatalf("expected nested accessToken to be scrubbed, got %q", got)
	}
	if got := creds["tokenType"]; got != "Bearer" {
		t.Fatalf("expected sibling field to be untouched, got %q", got)
	}
}

func TestObfuscateValues_ScrubsWithinArray(t *testing.T) {
	m := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"accessToken": testAccessTokenSecret, "id": float64(1)},
			map[string]interface{}{"id": float64(2)},
		},
	}

	result := obfuscateValues(m)

	items := result["items"].([]interface{})
	first := items[0].(map[string]interface{})
	if got := first["accessToken"]; got != "********" {
		t.Fatalf("expected accessToken in array element to be scrubbed, got %q", got)
	}
	if got := first["id"]; got != float64(1) {
		t.Fatalf("expected sibling field to be untouched, got %v", got)
	}

	second := items[1].(map[string]interface{})
	if _, ok := second["accessToken"]; ok {
		t.Fatalf("did not expect an accessToken key to be added to an element that never had one")
	}
}

func TestObfuscateValues_NilMap(t *testing.T) {
	// obfuscateValues should not panic when given a nil map (e.g. an empty
	// JSON body of "{}" unmarshals to a non-nil empty map, but guard against
	// nil defensively since m[v] lookups on a nil map are legal in Go).
	result := obfuscateValues(nil)
	if result != nil {
		t.Fatalf("expected nil map to be returned unchanged, got %v", result)
	}
}

func TestPrettyPrintJsonLines_ScrubsAccessTokenInBody(t *testing.T) {
	body := `{"accessToken":"` + testAccessTokenSecret + `","tokenType":"Bearer"}`

	out, err := prettyPrintJsonLines([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected accessToken to be scrubbed from output, got:\n%s", out)
	}
	if !strings.Contains(out, "********") {
		t.Fatalf("expected scrubbed placeholder in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Bearer") {
		t.Fatalf("expected non-sensitive fields to be preserved, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_PreservesNonJsonLines(t *testing.T) {
	// Mimics the shape of an httputil.DumpRequestOut/DumpResponse dump:
	// a request/status line and headers (not valid JSON on their own),
	// followed by a blank line and a JSON body.
	dump := strings.Join([]string{
		"POST /admin/directory/v1/users HTTP/1.1",
		"Host: admin.googleapis.com",
		"Authorization: Bearer " + testAccessTokenSecret,
		"",
		`{"accessToken":"` + testAccessTokenSecret + `"}`,
	}, "\n")

	out, err := prettyPrintJsonLines([]byte(dump))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-JSON, non-sensitive-header lines are passed through unmodified.
	if !strings.Contains(out, "Host: admin.googleapis.com") {
		t.Fatalf("expected non-JSON lines to be preserved, got:\n%s", out)
	}

	// The Authorization header line is redacted (see
	// TestPrettyPrintJsonLines_RedactsAuthorizationHeader for the dedicated
	// coverage of that behavior).
	if strings.Contains(out, testAccessTokenSecret) {
		// The JSON body's accessToken is also scrubbed, so if the raw
		// secret still appears anywhere in the output at all, something
		// regressed - either the header or the body scrubbing.
		t.Fatalf("expected no occurrence of the raw secret anywhere in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Authorization: ********") {
		t.Fatalf("expected Authorization header value to be redacted, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_RedactsAuthorizationHeader(t *testing.T) {
	dump := "GET / HTTP/1.1\r\nAuthorization: Bearer " + testAccessTokenSecret + "\r\nHost: example.com\r\n"

	out, err := prettyPrintJsonLines([]byte(dump))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected Authorization header value to be redacted, got:\n%s", out)
	}
	if !strings.Contains(out, "Authorization: ********") {
		t.Fatalf("expected redacted Authorization header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Host: example.com") {
		t.Fatalf("expected unrelated headers to be preserved, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_RedactsProxyAuthorizationHeaderCaseInsensitively(t *testing.T) {
	// Header names are case-insensitive per RFC 7230; a lowercase variant
	// (as some HTTP/2 or proxy implementations emit) must still be caught.
	dump := "proxy-authorization: Basic " + testAccessTokenSecret

	out, err := prettyPrintJsonLines([]byte(dump))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected Proxy-Authorization header value to be redacted regardless of case, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_ScrubsMultipleJsonLines(t *testing.T) {
	dump := strings.Join([]string{
		`{"accessToken":"` + testAccessTokenSecret + `","seq":1}`,
		"some non-json line in between",
		`{"accessToken":"another-secret-token","seq":2}`,
	}, "\n")

	out, err := prettyPrintJsonLines([]byte(dump))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) || strings.Contains(out, "another-secret-token") {
		t.Fatalf("expected accessToken to be scrubbed on every JSON line, got:\n%s", out)
	}
	if !strings.Contains(out, "some non-json line in between") {
		t.Fatalf("expected non-JSON line to be preserved, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_ScrubsNestedAccessToken(t *testing.T) {
	// obfuscateValues recurses into nested objects, so an accessToken buried
	// under a sub-object is scrubbed too, not just top-level fields.
	body := `{"credentials":{"accessToken":"` + testAccessTokenSecret + `"}}`

	out, err := prettyPrintJsonLines([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected nested accessToken to be scrubbed, got:\n%s", out)
	}
	if !strings.Contains(out, "********") {
		t.Fatalf("expected scrubbed placeholder in output, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_ScrubsAccessTokenInArray(t *testing.T) {
	// obfuscateValues also recurses into arrays of objects, e.g. a list
	// response like {"items": [{"accessToken": "..."}]}.
	body := `{"items":[{"accessToken":"` + testAccessTokenSecret + `","id":1},{"id":2}]}`

	out, err := prettyPrintJsonLines([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected accessToken nested in an array element to be scrubbed, got:\n%s", out)
	}
	if !strings.Contains(out, "********") {
		t.Fatalf("expected scrubbed placeholder in output, got:\n%s", out)
	}
}

func TestPrettyPrintJsonLines_NoJsonContent(t *testing.T) {
	dump := "HTTP/1.1 200 OK\nContent-Type: text/plain\n\nnot json at all"

	out, err := prettyPrintJsonLines([]byte(dump))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != dump {
		t.Fatalf("expected content without any JSON lines to be returned unchanged, got:\n%s", out)
	}
}
