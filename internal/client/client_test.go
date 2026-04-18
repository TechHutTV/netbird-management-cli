package client

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRedactSensitiveJSON_MasksTopLevelKeys(t *testing.T) {
	in := []byte(`{"name":"svc","password":"hunter2","pin":"1234","port":443}`)
	out := redactSensitiveJSON(in)

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if m["password"] != "[REDACTED]" {
		t.Errorf("password not redacted: %v", m["password"])
	}
	if m["pin"] != "[REDACTED]" {
		t.Errorf("pin not redacted: %v", m["pin"])
	}
	if m["name"] != "svc" {
		t.Errorf("name altered: %v", m["name"])
	}
	if m["port"].(float64) != 443 {
		t.Errorf("port altered: %v", m["port"])
	}
}

func TestRedactSensitiveJSON_MasksNestedAuth(t *testing.T) {
	in := []byte(`{
		"name":"api",
		"auth": {
			"password_auth": {"enabled": true, "password": "s3cret"},
			"pin_auth": {"enabled": true, "pin": "4242"},
			"header_auths": [
				{"enabled": true, "header": "X-Auth", "value": "shared-token"},
				{"enabled": false, "header": "X-Other", "value": "other-value"}
			]
		}
	}`)
	out := redactSensitiveJSON(in)
	s := string(out)

	if strings.Contains(s, "s3cret") {
		t.Errorf("password leaked: %s", s)
	}
	if strings.Contains(s, "4242") {
		t.Errorf("pin leaked: %s", s)
	}
	if strings.Contains(s, "shared-token") || strings.Contains(s, "other-value") {
		t.Errorf("header_auth value leaked: %s", s)
	}
	if !strings.Contains(s, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker: %s", s)
	}
	// Header NAMES should stay visible (they're not secret).
	if !strings.Contains(s, "X-Auth") || !strings.Contains(s, "X-Other") {
		t.Errorf("header names should not be redacted: %s", s)
	}
}

func TestRedactSensitiveJSON_DoesNotTouchValueOutsideHeaderAuths(t *testing.T) {
	// A top-level "value" key (not inside header_auths) must not be redacted —
	// e.g., access-rule values like "10.0.0.0/24" or country codes must survive.
	in := []byte(`{"rules":[{"action":"allow","type":"cidr","value":"10.0.0.0/24"}]}`)
	out := redactSensitiveJSON(in)
	if !strings.Contains(string(out), "10.0.0.0/24") {
		t.Errorf("access-rule value got over-redacted: %s", string(out))
	}
}

func TestRedactSensitiveJSON_HandlesGenericSecretKeys(t *testing.T) {
	in := []byte(`{"token":"tk","access_token":"at","refresh_token":"rt","secret":"s","api_key":"k","client_secret":"cs"}`)
	out := redactSensitiveJSON(in)
	for _, leak := range []string{"tk", "at", "rt", "\"s\"", "\"k\"", "cs"} {
		if strings.Contains(string(out), leak) {
			t.Errorf("leak %q in output: %s", leak, string(out))
		}
	}
}

func TestRedactSensitiveJSON_InvalidJSONPassesThrough(t *testing.T) {
	in := []byte(`not valid json`)
	out := redactSensitiveJSON(in)
	if !reflect.DeepEqual(in, out) {
		t.Errorf("non-JSON input should pass through unchanged: %q vs %q", in, out)
	}
}

func TestRedactSensitiveJSON_NonStringSensitiveValueSkipped(t *testing.T) {
	// `token: null` or `password: 123` (unusual but possible) should not crash
	// and should not replace non-string values (would mis-type).
	in := []byte(`{"password":null,"token":0}`)
	out := redactSensitiveJSON(in)
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if m["password"] != nil {
		t.Errorf("null password altered: %v", m["password"])
	}
	if m["token"].(float64) != 0 {
		t.Errorf("numeric token altered: %v", m["token"])
	}
}
