// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func getValuesToScrub() []string {
	return []string{
		"accessToken",
	}
}

// sensitiveHeaders lists HTTP header names (matched case-insensitively)
// whose values must never be written to logs verbatim, even at
// TF_LOG=DEBUG.
var sensitiveHeaders = []string{"authorization", "proxy-authorization"}

type loggingTransport struct {
	name      string
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if logging.IsDebugOrHigher() {
		reqData, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(reqData)
			if err != nil {
				return nil, err
			}
			log.Printf("[DEBUG] "+logReqMsg, t.name, prettyPrint)
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if logging.IsDebugOrHigher() {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(respData)
			if err != nil {
				return nil, err
			}
			log.Printf("[DEBUG] "+logRespMsg, t.name, prettyPrint)
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

func NewTransportWithScrubbedLogs(name string, t http.RoundTripper) *loggingTransport {
	return &loggingTransport{name, t}
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
// this was copied from the SDK's logging package
func prettyPrintJsonLines(b []byte) (string, error) {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if redacted, ok := redactSensitiveHeaderLine(p); ok {
			parts[i] = redacted
			continue
		}
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer

			var jsonMap map[string]interface{}
			err := json.Unmarshal(b, &jsonMap)
			if err != nil {
				return "", err
			}

			jsonMap = obfuscateValues(jsonMap)

			b, err = json.Marshal(jsonMap)
			if err != nil {
				return "", err
			}

			json.Indent(&out, b, "", " ")
			parts[i] = out.String()
		}
	}
	return strings.Join(parts, "\n"), nil
}

// redactSensitiveHeaderLine checks whether a single line from an HTTP
// request/response dump (as produced by httputil.DumpRequestOut/
// DumpResponse) is a header line for one of sensitiveHeaders, and if so
// returns it with the value replaced.
//
// In this provider's current transport chain, the OAuth2 transport that
// attaches the Authorization header wraps *inside* this logging transport
// (see SetupClient in provider_config.go), so the header isn't actually
// present yet at the point requests get dumped for logging. This check
// exists anyway as defense-in-depth: that safety depends on transport
// wiring in a different file, which is exactly the kind of invariant a
// future refactor - or a different call path reusing this function - could
// break silently. Redacting here doesn't depend on any of that.
func redactSensitiveHeaderLine(line string) (string, bool) {
	hasCR := strings.HasSuffix(line, "\r")
	trimmed := strings.TrimSuffix(line, "\r")

	name, _, found := strings.Cut(trimmed, ":")
	if !found {
		return "", false
	}
	if !containsString(sensitiveHeaders, strings.ToLower(strings.TrimSpace(name))) {
		return "", false
	}

	redacted := name + ": ********"
	if hasCR {
		redacted += "\r"
	}
	return redacted, true
}

// obfuscateValues scrubs sensitive fields out of a decoded JSON object,
// recursing into nested objects and arrays so a sensitive field isn't only
// caught at the top level (e.g. a token nested under a "credentials" object,
// or inside a list of items).
func obfuscateValues(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return m
	}

	scrub := getValuesToScrub()
	for k, v := range m {
		if containsString(scrub, k) {
			m[k] = "********"
			continue
		}
		m[k] = obfuscateValue(v)
	}

	return m
}

// obfuscateValue recurses into a single decoded JSON value, scrubbing any
// sensitive fields found in nested objects/arrays. Scalars are returned as-is.
func obfuscateValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return obfuscateValues(val)
	case []interface{}:
		for i, item := range val {
			val[i] = obfuscateValue(item)
		}
		return val
	default:
		return v
	}
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
