package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer

	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(reader)
		done <- string(data)
	}()

	callErr := fn()
	_ = writer.Close()
	os.Stdout = original
	output := <-done
	_ = reader.Close()
	return output, callErr
}

func testService(serverURL string) *Service {
	return NewService(client.New("test-token", serverURL))
}

func TestListCountriesAcceptsObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/locations/countries" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"country_code":"US","country_name":"United States"},{"country_code":"DE","country_name":"Germany"}]`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listCountryCodes("json")
	})
	if err != nil {
		t.Fatalf("list countries: %v", err)
	}

	var countries []map[string]string
	if err := json.Unmarshal([]byte(output), &countries); err != nil {
		t.Fatalf("decode output: %v\noutput: %s", err, output)
	}
	if got := countries[0]["country_name"]; got != "United States" {
		t.Fatalf("country name = %q, want United States", got)
	}
}

func TestListCountriesAcceptsLegacyStringResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `["US","DE"]`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listCountryCodes("json")
	})
	if err != nil {
		t.Fatalf("list legacy countries: %v", err)
	}

	var countries []map[string]string
	if err := json.Unmarshal([]byte(output), &countries); err != nil {
		t.Fatalf("decode output: %v\noutput: %s", err, output)
	}
	if got := countries[0]["country_code"]; got != "US" {
		t.Fatalf("country code = %q, want US", got)
	}
}

func TestUserInspectUsesUsersListEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"user-1","email":"user@example.com","name":"Test User","role":"admin","status":"active","is_service_user":true}]`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).HandleUsersCommand([]string{"user", "--inspect", "user-1", "--output", "json"})
	})
	if err != nil {
		t.Fatalf("inspect user: %v", err)
	}

	var user map[string]any
	if err := json.Unmarshal([]byte(output), &user); err != nil {
		t.Fatalf("decode output: %v\noutput: %s", err, output)
	}
	if got := user["id"]; got != "user-1" {
		t.Fatalf("user id = %v, want user-1", got)
	}
}

func TestBuiltInHelpMatchesDNSAndGeoParsers(t *testing.T) {
	dnsHelp, err := captureStdout(t, func() error {
		PrintDNSUsage()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dnsHelp, "--get-settings") || strings.Contains(dnsHelp, "  --settings") {
		t.Fatalf("DNS help does not match parser:\n%s", dnsHelp)
	}

	geoHelp, err := captureStdout(t, func() error {
		PrintGeoLocationUsage()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"--cities --country <country-code>", "--output <table|json>"} {
		if !strings.Contains(geoHelp, expected) {
			t.Fatalf("Geo help missing %q:\n%s", expected, geoHelp)
		}
	}
	if strings.Contains(geoHelp, "--json") {
		t.Fatalf("Geo help advertises unsupported --json flag:\n%s", geoHelp)
	}
}

func TestAccountInspectSelectsAccountFromListEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/accounts" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"account-1","domain":"example.test","settings":{"network_range":"100.64.0.0/16"}}]`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).inspectAccount("account-1", "json")
	})
	if err != nil {
		t.Fatalf("inspect account: %v", err)
	}

	var account map[string]any
	if err := json.Unmarshal([]byte(output), &account); err != nil {
		t.Fatalf("decode output: %v\noutput: %s", err, output)
	}
	if got := account["id"]; got != "account-1" {
		t.Fatalf("account id = %v, want account-1", got)
	}
}

func TestAccountUpdateLoadsCurrentStateFromListEndpoint(t *testing.T) {
	putCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/accounts":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `[{"id":"account-1","domain":"example.test","settings":{"network_range":"100.64.0.0/16","peer_login_expiration":86400}}]`)
		case r.Method == http.MethodPut && r.URL.Path == "/accounts/account-1":
			putCalled = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, err := captureStdout(t, func() error {
		return testService(server.URL).updateAccountFromFlags("account-1", accountUpdateFlags{})
	})
	if err != nil {
		t.Fatalf("update account: %v", err)
	}
	if !putCalled {
		t.Fatal("account update did not send PUT request")
	}
}

func TestPeerWildcardFiltersAreAppliedLocally(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/peers" {
			http.NotFound(w, r)
			return
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("wildcard filter sent to exact-match API: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"peer-1","name":"fedora-ai","ip":"100.64.151.90"},{"id":"peer-2","name":"laptop","ip":"100.90.1.2"}]`)
	}))
	defer server.Close()

	for name, filters := range map[string][2]string{
		"name": {"fed*", ""},
		"ip":   {"", "100.64.*"},
	} {
		t.Run(name, func(t *testing.T) {
			filterName, filterIP := filters[0], filters[1]
			output, err := captureStdout(t, func() error {
				return testService(server.URL).listPeers(filterName, filterIP, "json")
			})
			if err != nil {
				t.Fatalf("list peers: %v", err)
			}
			var peers []map[string]any
			if err := json.Unmarshal([]byte(output), &peers); err != nil {
				t.Fatalf("decode peer output: %v\noutput: %s", err, output)
			}
			if len(peers) != 1 || peers[0]["id"] != "peer-1" {
				t.Fatalf("filtered peers = %#v, want peer-1", peers)
			}
		})
	}
}

func TestAuditFiltersAreAppliedLocally(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/audit" {
			http.NotFound(w, r)
			return
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("unsupported audit filters sent to API: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
			{"id":"1","timestamp":"2026-07-20T12:00:00Z","activity":"Peer created","activity_code":"peer.add","initiator_id":"user-1","initiator_name":"Alice","initiator_email":"alice@example.com","target_id":"peer-1","meta":{"name":"Fedora"}},
			{"id":"2","timestamp":"2026-07-22T12:00:00Z","activity":"Route deleted","activity_code":"route.delete","initiator_id":"user-2","initiator_name":"Bob","initiator_email":"bob@example.com","target_id":"route-1","meta":{"name":"Office"}}
		]`)
	}))
	defer server.Close()

	cases := map[string]models.AuditEventFilters{
		"user":     {UserID: "user-1"},
		"target":   {TargetID: "peer-1"},
		"activity": {ActivityCode: "peer.add"},
		"search":   {Search: "fedora"},
		"dates":    {StartDate: "2026-07-20T00:00:00Z", EndDate: "2026-07-21T00:00:00Z"},
	}
	for name, filters := range cases {
		t.Run(name, func(t *testing.T) {
			output, err := captureStdout(t, func() error {
				return testService(server.URL).listAuditEvents(filters, "json")
			})
			if err != nil {
				t.Fatalf("list audit events: %v", err)
			}
			var events []map[string]any
			if err := json.Unmarshal([]byte(output), &events); err != nil {
				t.Fatalf("decode audit output: %v\noutput: %s", err, output)
			}
			if len(events) != 1 || events[0]["id"] != "1" {
				t.Fatalf("filtered events = %#v, want event 1", events)
			}
		})
	}
}

func TestAuditDateFiltersParseRFC3339Instants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
			{"id":"before","timestamp":"2026-07-20T01:00:00+01:00","activity":"Before"},
			{"id":"inside","timestamp":"2026-07-20T00:45:00.123456789Z","activity":"Inside"}
		]`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listAuditEvents(models.AuditEventFilters{
			StartDate: "2026-07-20T00:30:00Z",
			EndDate:   "2026-07-20T02:00:00Z",
		}, "json")
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	var events []models.AuditEvent
	if err := json.Unmarshal([]byte(output), &events); err != nil {
		t.Fatalf("decode audit output: %v", err)
	}
	if len(events) != 1 || events[0].ID != "inside" {
		t.Fatalf("date-filtered events = %#v, want inside", events)
	}
}

func TestAuditDateFiltersRejectMalformedTimestamps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"1","timestamp":"not-a-date"}]`)
	}))
	defer server.Close()
	svc := testService(server.URL)

	if err := svc.listAuditEvents(models.AuditEventFilters{StartDate: "bad-bound"}, "json"); err == nil {
		t.Fatal("malformed audit bound returned nil error")
	}
	if err := svc.listAuditEvents(models.AuditEventFilters{StartDate: "2026-07-20T00:00:00Z"}, "json"); err == nil {
		t.Fatal("malformed event timestamp returned nil error")
	}
}

func TestProxySourceIPUsesWorkingServerFilter(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("source_ip") {
		case "10.0.0.2":
			fmt.Fprint(w, `{"data":[{"id":"native","source_ip":"10.0.0.2"}],"page":1,"page_size":10,"total_records":1,"total_pages":1}`)
		case "":
			if r.URL.Query().Get("search") != "" {
				t.Fatalf("native filter support incorrectly used fallback search")
			}
			fmt.Fprint(w, `{"data":[{"id":"other","source_ip":"10.0.0.1"}],"page":1,"page_size":10,"total_records":50,"total_pages":5}`)
		default:
			t.Fatalf("unexpected source_ip query")
		}
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listProxyEvents(models.ProxyEventFilters{
			Page: 1, PageSize: 10, SourceIP: "10.0.0.2",
		}, "json")
	})
	if err != nil {
		t.Fatalf("list proxy events: %v", err)
	}
	var response models.ProxyEventResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("decode proxy output: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "native" {
		t.Fatalf("native source-IP response = %#v", response.Data)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want filtered request plus support probe", requestCount)
	}
}

func TestProxySourceIPFallsBackToBoundedSearch(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/proxy" {
			http.NotFound(w, r)
			return
		}
		requestCount++
		if got := r.URL.Query().Get("source_ip"); got != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[],"page":1,"page_size":10,"total_records":0,"total_pages":0}`)
			return
		}
		if got := r.URL.Query().Get("search"); got != "10.0.0.2" {
			t.Fatalf("fallback search = %q, want source IP", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `{"data":[{"id":"log-2","source_ip":"10.0.0.2"}],"page":2,"page_size":100,"total_records":3,"total_pages":2}`)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"wrong","source_ip":"10.0.0.1"},{"id":"log-1","source_ip":"10.0.0.2"}],"page":1,"page_size":100,"total_records":3,"total_pages":2}`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listProxyEvents(models.ProxyEventFilters{
			Page: 2, PageSize: 1, SourceIP: "10.0.0.2",
		}, "json")
	})
	if err != nil {
		t.Fatalf("list proxy events: %v", err)
	}
	var response models.ProxyEventResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("decode proxy output: %v\noutput: %s", err, output)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "log-2" {
		t.Fatalf("proxy response = %#v, want log-2", response.Data)
	}
	if response.TotalRecords != 2 || response.TotalPages != 2 || response.Page != 2 || response.PageSize != 1 {
		t.Fatalf("proxy pagination = page %d size %d records %d pages %d, want 2/1/2/2",
			response.Page, response.PageSize, response.TotalRecords, response.TotalPages)
	}
	if requestCount != 3 {
		t.Fatalf("request count = %d, want native attempt plus two bounded search pages", requestCount)
	}
}

func TestProxySourceIPPreservesServerSearchSemantics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source_ip") != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[],"page":1,"page_size":50,"total_records":0,"total_pages":0}`)
			return
		}
		if got := r.URL.Query().Get("search"); got != "alice@example.com" {
			t.Fatalf("fallback search = %q, want original search", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"alice-wrong-ip","source_ip":"10.0.0.1"},{"id":"alice-right-ip","source_ip":"10.0.0.2"}],"page":1,"page_size":100,"total_records":2,"total_pages":1}`)
	}))
	defer server.Close()

	output, err := captureStdout(t, func() error {
		return testService(server.URL).listProxyEvents(models.ProxyEventFilters{
			SourceIP: "10.0.0.2", Search: "alice@example.com",
		}, "json")
	})
	if err != nil {
		t.Fatalf("list proxy events: %v", err)
	}
	var response models.ProxyEventResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("decode proxy output: %v", err)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "alice-right-ip" {
		t.Fatalf("combined filter response = %#v", response.Data)
	}
}

func TestProxySourceIPFallbackIsCappedAndUsesEmptyArray(t *testing.T) {
	t.Run("cap", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("source_ip") != "" {
				fmt.Fprint(w, `{"data":[],"page":1,"page_size":50,"total_records":0,"total_pages":0}`)
				return
			}
			fmt.Fprint(w, `{"data":[{"id":"1","source_ip":"10.0.0.2"}],"page":1,"page_size":100,"total_records":999999,"total_pages":9999}`)
		}))
		defer server.Close()
		_, err := captureStdout(t, func() error {
			return testService(server.URL).listProxyEvents(models.ProxyEventFilters{SourceIP: "10.0.0.2"}, "json")
		})
		if err == nil || !strings.Contains(err.Error(), "too many") {
			t.Fatalf("cap error = %v, want too many", err)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("source_ip") != "" {
				fmt.Fprint(w, `{"data":[],"page":1,"page_size":50,"total_records":0,"total_pages":0}`)
				return
			}
			fmt.Fprint(w, `{"data":[{"id":"wrong","source_ip":"10.0.0.1"}],"page":1,"page_size":100,"total_records":1,"total_pages":1}`)
		}))
		defer server.Close()
		output, err := captureStdout(t, func() error {
			return testService(server.URL).listProxyEvents(models.ProxyEventFilters{SourceIP: "10.0.0.2"}, "json")
		})
		if err != nil {
			t.Fatalf("list proxy events: %v", err)
		}
		if !strings.Contains(output, `"data": []`) {
			t.Fatalf("empty proxy data was not []: %s", output)
		}
	})
}

func TestProxySourceIPFallbackRejectsDishonestMetadataAndExtremePagination(t *testing.T) {
	t.Run("dishonest metadata", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("source_ip") != "" {
				fmt.Fprint(w, `{"data":[],"page":1,"page_size":50,"total_records":0,"total_pages":0}`)
				return
			}
			requestCount++
			data := make([]models.ProxyEvent, 101)
			for i := range data {
				data[i] = models.ProxyEvent{ID: fmt.Sprintf("%d", i), SourceIP: "10.0.0.2"}
			}
			payload, err := json.Marshal(models.ProxyEventResponse{
				Data: data, Page: requestCount, PageSize: 100,
				TotalRecords: 1, TotalPages: 21,
			})
			if err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write(payload)
		}))
		defer server.Close()
		_, err := captureStdout(t, func() error {
			return testService(server.URL).listProxyEvents(models.ProxyEventFilters{SourceIP: "10.0.0.2"}, "json")
		})
		if err == nil || !strings.Contains(err.Error(), "too many") {
			t.Fatalf("dishonest metadata cap error = %v, want too many", err)
		}
	})

	for name, filters := range map[string]models.ProxyEventFilters{
		"extreme page":      {SourceIP: "10.0.0.2", Page: int(^uint(0) >> 1), PageSize: 2},
		"extreme page size": {SourceIP: "10.0.0.2", Page: 2, PageSize: int(^uint(0) >> 1)},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Query().Get("source_ip") != "" {
					fmt.Fprint(w, `{"data":[],"page":1,"page_size":50,"total_records":0,"total_pages":0}`)
					return
				}
				fmt.Fprint(w, `{"data":[{"id":"one","source_ip":"10.0.0.2"}],"page":1,"page_size":100,"total_records":1,"total_pages":1}`)
			}))
			defer server.Close()
			_, err := captureStdout(t, func() error {
				return testService(server.URL).listProxyEvents(filters, "json")
			})
			if err == nil || !strings.Contains(err.Error(), "pagination") {
				t.Fatalf("pagination error = %v, want pagination", err)
			}
		})
	}
}

func TestEmptyCollectionsRemainValidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	}))
	defer server.Close()
	svc := testService(server.URL)

	cases := map[string]func() error{
		"peers":          func() error { return svc.listPeers("", "", "json") },
		"groups":         func() error { return svc.listGroups("", "json") },
		"networks":       func() error { return svc.listNetworks("", "json") },
		"setup keys":     func() error { return svc.listSetupKeys("", "", false, "json") },
		"users":          func() error { return svc.listUsers("", "json") },
		"user invites":   func() error { return svc.listUserInvites("json") },
		"tokens":         func() error { return svc.listTokens("user-1", "json") },
		"accounts":       func() error { return svc.listAccounts("json") },
		"dns groups":     func() error { return svc.listDNSGroups(&DNSFilters{EnabledOnly: true}, "json") },
		"dns zones":      func() error { return svc.listDNSZones("json") },
		"dns records":    func() error { return svc.listDNSRecords("zone-1", "json") },
		"jobs":           func() error { return svc.listPeerJobs("peer-1", "json") },
		"ingress ports":  func() error { return svc.listIngressPorts("peer-1", "", "json") },
		"ingress peers":  func() error { return svc.listIngressPeers("json") },
		"notifications":  func() error { return svc.listNotificationChannels("json") },
		"routes":         func() error { return svc.listRoutes(&RouteFilters{EnabledOnly: true}, "json") },
		"policies":       func() error { return svc.listPolicies(&policyFilters{EnabledOnly: true}, "json") },
		"posture checks": func() error { return svc.listPostureChecks(&PostureCheckFilters{CheckType: "process"}, "json") },
	}
	for name, invoke := range cases {
		t.Run(name, func(t *testing.T) {
			output, err := captureStdout(t, invoke)
			if err != nil {
				t.Fatalf("list empty collection: %v", err)
			}
			var values []any
			if err := json.Unmarshal([]byte(output), &values); err != nil {
				t.Fatalf("empty output is not JSON: %v\noutput: %q", err, output)
			}
			if strings.TrimSpace(output) != "[]" {
				t.Fatalf("empty collection output = %q, want []", output)
			}
		})
	}
}

func TestCommandHandlersReturnFlagParseErrors(t *testing.T) {
	svc := &Service{}
	handlers := map[string]func([]string) error{
		"accounts":       svc.HandleAccountsCommand,
		"dns":            svc.HandleDNSCommand,
		"dns-zones":      svc.HandleDNSZonesCommand,
		"events":         svc.HandleEventsCommand,
		"geo":            svc.HandleGeoLocationsCommand,
		"groups":         svc.HandleGroupsCommand,
		"ingress-peers":  svc.HandleIngressPeersCommand,
		"ingress-ports":  svc.HandleIngressPortsCommand,
		"jobs":           svc.HandleJobsCommand,
		"networks":       svc.HandleNetworkCommand,
		"notifications":  svc.HandleNotificationsCommand,
		"peers":          svc.HandlePeersCommand,
		"policies":       svc.HandlePoliciesCommand,
		"posture-checks": svc.HandlePostureChecksCommand,
		"routes":         svc.HandleRoutesCommand,
		"setup-keys":     svc.HandleSetupKeysCommand,
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			if err := handler([]string{name, "--not-a-real-flag"}); err == nil {
				t.Fatal("invalid flag returned nil error")
			}
		})
	}

	if err := HandleMigrateCommand([]string{"migrate", "--not-a-real-flag"}, false); err == nil {
		t.Fatal("migrate invalid flag returned nil error")
	}
}
