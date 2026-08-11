// Copyright (c) HashiCorp, Inc.
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

	// Non-JSON lines (including the Authorization header) are passed through
	// unmodified by this function - obfuscateValues only ever runs on lines
	// that are themselves valid, whole-line JSON. Header scrubbing is not
	// this function's job; it documents current behavior so a change here
	// doesn't silently start leaking (or silently start redacting) header
	// lines without a test noticing.
	if !strings.Contains(out, "Authorization: Bearer "+testAccessTokenSecret) {
		t.Fatalf("expected non-JSON header lines to pass through unchanged, got:\n%s", out)
	}
	if !strings.Contains(out, "Host: admin.googleapis.com") {
		t.Fatalf("expected non-JSON lines to be preserved, got:\n%s", out)
	}

	// The JSON body line, however, must have its accessToken field scrubbed.
	lines := strings.Split(out, "\n")
	bodyLine := lines[len(lines)-1]
	if strings.Contains(bodyLine, testAccessTokenSecret) {
		t.Fatalf("expected accessToken in JSON body line to be scrubbed, got:\n%s", bodyLine)
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

func TestPrettyPrintJsonLines_DoesNotScrubNestedAccessToken(t *testing.T) {
	// obfuscateValues only scrubs top-level keys, so an accessToken nested
	// inside a sub-object currently survives. This test documents that
	// known limitation rather than asserting it's desired behavior.
	body := `{"credentials":{"accessToken":"` + testAccessTokenSecret + `"}}`

	out, err := prettyPrintJsonLines([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, testAccessTokenSecret) {
		t.Fatalf("expected current implementation to leave nested accessToken unscrubbed, got:\n%s", out)
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
