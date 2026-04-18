package tui

import (
	"reflect"
	"testing"

	"netbird-manage/internal/models"
)

// ─── cloneService (I5 regression) ───────────────────────────────────

func TestCloneService_DeepCopiesNestedFields(t *testing.T) {
	original := models.ReverseProxyService{
		ID:      "svc-1",
		Name:    "api",
		Domain:  "api.example.com",
		Mode:    models.ReverseProxyModeHTTP,
		Enabled: true,
		Targets: []models.ReverseProxyTarget{
			{
				TargetType: "peer",
				TargetID:   "peer-1",
				Protocol:   "http",
				Port:       8080,
				Options: &models.ReverseProxyTargetOpts{
					SkipTLSVerify: true,
					CustomHeaders: map[string]string{"X-Foo": "bar"},
				},
			},
		},
		Auth: &models.ReverseProxyAuth{
			PasswordAuth: &models.ReverseProxyPasswordAuth{Enabled: true, Password: "secret"},
			BearerAuth: &models.ReverseProxyBearerAuth{
				Enabled:            true,
				DistributionGroups: []string{"g1", "g2"},
			},
			HeaderAuths: []models.ReverseProxyHeaderAuth{
				{Enabled: true, Header: "X-API", Value: "tok"},
			},
		},
		AccessRestrictions: &models.ReverseProxyAccessRestrictions{
			AllowedCIDRs:     []string{"10.0.0.0/24"},
			BlockedCountries: []string{"KP"},
		},
	}

	clone := cloneService(original)

	// Mutate every nested slice/map/pointer and confirm the original is untouched.
	clone.Targets[0].TargetID = "peer-9"
	clone.Targets[0].Options.CustomHeaders["X-Foo"] = "mutated"
	clone.Targets[0].Options.SkipTLSVerify = false
	clone.Auth.PasswordAuth.Password = "changed"
	clone.Auth.BearerAuth.DistributionGroups[0] = "other"
	clone.Auth.HeaderAuths[0].Value = "different"
	clone.AccessRestrictions.AllowedCIDRs[0] = "1.1.1.1/32"
	clone.AccessRestrictions.BlockedCountries = append(clone.AccessRestrictions.BlockedCountries, "CN")

	if original.Targets[0].TargetID != "peer-1" {
		t.Fatalf("target id mutated: %q", original.Targets[0].TargetID)
	}
	if original.Targets[0].Options.CustomHeaders["X-Foo"] != "bar" {
		t.Fatalf("custom headers map shared: %q", original.Targets[0].Options.CustomHeaders["X-Foo"])
	}
	if original.Targets[0].Options.SkipTLSVerify != true {
		t.Fatalf("target options pointer shared")
	}
	if original.Auth.PasswordAuth.Password != "secret" {
		t.Fatalf("password auth pointer shared: %q", original.Auth.PasswordAuth.Password)
	}
	if original.Auth.BearerAuth.DistributionGroups[0] != "g1" {
		t.Fatalf("distribution groups slice shared: %q", original.Auth.BearerAuth.DistributionGroups[0])
	}
	if original.Auth.HeaderAuths[0].Value != "tok" {
		t.Fatalf("header auths slice shared: %q", original.Auth.HeaderAuths[0].Value)
	}
	if original.AccessRestrictions.AllowedCIDRs[0] != "10.0.0.0/24" {
		t.Fatalf("allowed cidrs slice shared: %q", original.AccessRestrictions.AllowedCIDRs[0])
	}
	if len(original.AccessRestrictions.BlockedCountries) != 1 {
		t.Fatalf("blocked countries slice shared: len=%d", len(original.AccessRestrictions.BlockedCountries))
	}
}

func TestCloneService_HandlesNilOptionalFields(t *testing.T) {
	original := models.ReverseProxyService{
		ID:     "svc-0",
		Name:   "minimal",
		Domain: "a.example.com",
		Mode:   models.ReverseProxyModeHTTP,
	}
	clone := cloneService(original)
	if clone.Auth != nil || clone.AccessRestrictions != nil || clone.Meta != nil {
		t.Fatalf("nil optional fields got materialized: %+v", clone)
	}
	if len(clone.Targets) != 0 {
		t.Fatalf("empty targets got materialized with entries: %d", len(clone.Targets))
	}
}

// ─── classifyCIDREntry ──────────────────────────────────────────────

func TestClassifyCIDREntry(t *testing.T) {
	tests := []struct {
		name                          string
		action, value                 string
		wantAction, wantKind, wantVal string
	}{
		{"single ip as /32", "allow", "192.168.1.1/32", "allow", "ip", "192.168.1.1"},
		{"cidr block", "block", "10.0.0.0/24", "block", "cidr", "10.0.0.0/24"},
		{"no suffix falls through to cidr bucket", "allow", "10.0.0.1", "allow", "cidr", "10.0.0.1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotA, gotK, gotV := classifyCIDREntry(tc.action, tc.value)
			if gotA != tc.wantAction || gotK != tc.wantKind || gotV != tc.wantVal {
				t.Errorf("got (%q,%q,%q), want (%q,%q,%q)", gotA, gotK, gotV, tc.wantAction, tc.wantKind, tc.wantVal)
			}
		})
	}
}

// ─── removeAccessRuleAt ─────────────────────────────────────────────

func TestRemoveAccessRuleAt(t *testing.T) {
	// Flattened ordering: AllowedCIDRs, BlockedCIDRs, AllowedCountries, BlockedCountries.
	baseline := func() *models.ReverseProxyAccessRestrictions {
		return &models.ReverseProxyAccessRestrictions{
			AllowedCIDRs:     []string{"10.0.0.0/24", "10.0.1.0/24"},
			BlockedCIDRs:     []string{"1.2.3.4/32"},
			AllowedCountries: []string{"US", "CA"},
			BlockedCountries: []string{"KP"},
		}
	}

	tests := []struct {
		name string
		idx  int
		want models.ReverseProxyAccessRestrictions
	}{
		{
			"first allowed cidr",
			0,
			models.ReverseProxyAccessRestrictions{
				AllowedCIDRs:     []string{"10.0.1.0/24"},
				BlockedCIDRs:     []string{"1.2.3.4/32"},
				AllowedCountries: []string{"US", "CA"},
				BlockedCountries: []string{"KP"},
			},
		},
		{
			"blocked cidr (idx 2)",
			2,
			models.ReverseProxyAccessRestrictions{
				AllowedCIDRs:     []string{"10.0.0.0/24", "10.0.1.0/24"},
				BlockedCIDRs:     []string{},
				AllowedCountries: []string{"US", "CA"},
				BlockedCountries: []string{"KP"},
			},
		},
		{
			"first allowed country (idx 3)",
			3,
			models.ReverseProxyAccessRestrictions{
				AllowedCIDRs:     []string{"10.0.0.0/24", "10.0.1.0/24"},
				BlockedCIDRs:     []string{"1.2.3.4/32"},
				AllowedCountries: []string{"CA"},
				BlockedCountries: []string{"KP"},
			},
		},
		{
			"last blocked country (idx 5)",
			5,
			models.ReverseProxyAccessRestrictions{
				AllowedCIDRs:     []string{"10.0.0.0/24", "10.0.1.0/24"},
				BlockedCIDRs:     []string{"1.2.3.4/32"},
				AllowedCountries: []string{"US", "CA"},
				BlockedCountries: []string{},
			},
		},
		{
			"out-of-range no-op",
			99,
			*baseline(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := removeAccessRuleAt(baseline(), tc.idx)
			if !reflect.DeepEqual(*got, tc.want) {
				t.Errorf("got %+v, want %+v", *got, tc.want)
			}
		})
	}
}

func TestRemoveAccessRuleAt_NilInput(t *testing.T) {
	if removeAccessRuleAt(nil, 0) != nil {
		t.Fatal("nil input should return nil")
	}
}

// ─── applyAccessRuleForm ────────────────────────────────────────────

func TestApplyAccessRuleForm_RoutesToCorrectBucket(t *testing.T) {
	tests := []struct {
		name        string
		data        rpAccessRuleFormData
		wantAllowCD []string
		wantBlockCD []string
		wantAllowCo []string
		wantBlockCo []string
	}{
		{
			"allow country",
			rpAccessRuleFormData{action: "allow", kind: "country", value: "US"},
			nil, nil, []string{"US"}, nil,
		},
		{
			"block country",
			rpAccessRuleFormData{action: "block", kind: "country", value: "KP"},
			nil, nil, nil, []string{"KP"},
		},
		{
			"allow cidr",
			rpAccessRuleFormData{action: "allow", kind: "cidr", value: "10.0.0.0/24"},
			[]string{"10.0.0.0/24"}, nil, nil, nil,
		},
		{
			"block cidr",
			rpAccessRuleFormData{action: "block", kind: "cidr", value: "10.0.0.0/24"},
			nil, []string{"10.0.0.0/24"}, nil, nil,
		},
		{
			"ip becomes /32 cidr on allow",
			rpAccessRuleFormData{action: "allow", kind: "ip", value: "192.168.1.1"},
			[]string{"192.168.1.1/32"}, nil, nil, nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := models.ReverseProxyService{}
			applyAccessRuleForm(&svc, -1, tc.data)
			ar := svc.AccessRestrictions
			if ar == nil {
				t.Fatal("access restrictions not created")
			}
			if !reflect.DeepEqual(ar.AllowedCIDRs, tc.wantAllowCD) {
				t.Errorf("allowed cidrs: got %v, want %v", ar.AllowedCIDRs, tc.wantAllowCD)
			}
			if !reflect.DeepEqual(ar.BlockedCIDRs, tc.wantBlockCD) {
				t.Errorf("blocked cidrs: got %v, want %v", ar.BlockedCIDRs, tc.wantBlockCD)
			}
			if !reflect.DeepEqual(ar.AllowedCountries, tc.wantAllowCo) {
				t.Errorf("allowed countries: got %v, want %v", ar.AllowedCountries, tc.wantAllowCo)
			}
			if !reflect.DeepEqual(ar.BlockedCountries, tc.wantBlockCo) {
				t.Errorf("blocked countries: got %v, want %v", ar.BlockedCountries, tc.wantBlockCo)
			}
		})
	}
}

func TestApplyAccessRuleForm_EditInBucketReplacesInPlace(t *testing.T) {
	svc := models.ReverseProxyService{
		AccessRestrictions: &models.ReverseProxyAccessRestrictions{
			AllowedCIDRs: []string{"10.0.0.0/24", "10.0.1.0/24"},
		},
	}
	// Edit the first rule (idx 0) to a new value in the SAME bucket.
	applyAccessRuleForm(&svc, 0, rpAccessRuleFormData{action: "allow", kind: "cidr", value: "10.0.9.0/24"})

	got := svc.AccessRestrictions.AllowedCIDRs
	want := []string{"10.0.9.0/24", "10.0.1.0/24"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after in-bucket edit: got %v, want %v", got, want)
	}
}

func TestApplyAccessRuleForm_EditAcrossBucketsMovesEntry(t *testing.T) {
	svc := models.ReverseProxyService{
		AccessRestrictions: &models.ReverseProxyAccessRestrictions{
			AllowedCIDRs:     []string{"10.0.0.0/24", "10.0.1.0/24"},
			BlockedCountries: []string{"KP"},
		},
	}
	// Edit idx 1 (second AllowedCIDR) — change action to block. Old slot gone, new entry appended to BlockedCIDRs.
	applyAccessRuleForm(&svc, 1, rpAccessRuleFormData{action: "block", kind: "cidr", value: "10.0.1.0/24"})

	if !reflect.DeepEqual(svc.AccessRestrictions.AllowedCIDRs, []string{"10.0.0.0/24"}) {
		t.Errorf("allowed cidrs after cross-bucket edit: %v", svc.AccessRestrictions.AllowedCIDRs)
	}
	if !reflect.DeepEqual(svc.AccessRestrictions.BlockedCIDRs, []string{"10.0.1.0/24"}) {
		t.Errorf("blocked cidrs after cross-bucket edit: %v", svc.AccessRestrictions.BlockedCIDRs)
	}
	if !reflect.DeepEqual(svc.AccessRestrictions.BlockedCountries, []string{"KP"}) {
		t.Errorf("unrelated bucket disturbed: %v", svc.AccessRestrictions.BlockedCountries)
	}
}

func TestApplyAccessRuleForm_BlankValueIgnored(t *testing.T) {
	svc := models.ReverseProxyService{}
	applyAccessRuleForm(&svc, -1, rpAccessRuleFormData{action: "allow", kind: "cidr", value: "   "})
	if svc.AccessRestrictions == nil {
		t.Fatal("access restrictions should be created even for blank-value no-ops")
	}
	if len(svc.AccessRestrictions.AllowedCIDRs) != 0 {
		t.Errorf("blank value should be ignored: got %v", svc.AccessRestrictions.AllowedCIDRs)
	}
}

// ─── applyTargetForm ────────────────────────────────────────────────

func TestApplyTargetForm_NewTargetAppended(t *testing.T) {
	svc := models.ReverseProxyService{}
	applyTargetForm(&svc, -1, rpTargetFormData{
		targetType: "peer",
		targetID:   "peer-1",
		protocol:   "http",
		port:       "8080",
		path:       "/api",
		enabled:    true,
	})
	if len(svc.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(svc.Targets))
	}
	got := svc.Targets[0]
	if got.TargetType != "peer" || got.TargetID != "peer-1" || got.Port != 8080 || got.Path != "/api" || !got.Enabled {
		t.Errorf("target fields wrong: %+v", got)
	}
}

func TestApplyTargetForm_HostTypeUsesHostField(t *testing.T) {
	svc := models.ReverseProxyService{}
	applyTargetForm(&svc, -1, rpTargetFormData{
		targetType: "host",
		targetID:   "should-be-ignored",
		host:       "10.0.0.5",
		protocol:   "tcp",
		port:       "443",
		enabled:    true,
	})
	got := svc.Targets[0]
	if got.Host != "10.0.0.5" {
		t.Errorf("host not set: %q", got.Host)
	}
	if got.TargetID != "" {
		t.Errorf("target ID should be empty for host targets, got %q", got.TargetID)
	}
}

func TestApplyTargetForm_PreservesExistingCustomHeaders(t *testing.T) {
	svc := models.ReverseProxyService{
		Targets: []models.ReverseProxyTarget{{
			TargetType: "peer",
			TargetID:   "peer-1",
			Protocol:   "http",
			Port:       80,
			Options: &models.ReverseProxyTargetOpts{
				CustomHeaders: map[string]string{"X-Preserved": "yes"},
				PathRewrite:   "preserve",
			},
		}},
	}
	// Edit target at idx 0 with no options-touching fields.
	applyTargetForm(&svc, 0, rpTargetFormData{
		targetType: "peer", targetID: "peer-1", protocol: "http", port: "80", enabled: true,
	})
	got := svc.Targets[0]
	if got.Options == nil {
		t.Fatal("options dropped")
	}
	if got.Options.CustomHeaders["X-Preserved"] != "yes" {
		t.Errorf("custom headers lost: %v", got.Options.CustomHeaders)
	}
	if got.Options.PathRewrite != "preserve" {
		t.Errorf("path rewrite lost: %q", got.Options.PathRewrite)
	}
}

func TestApplyTargetForm_OptionsOmittedWhenEmpty(t *testing.T) {
	svc := models.ReverseProxyService{}
	applyTargetForm(&svc, -1, rpTargetFormData{
		targetType: "peer", targetID: "peer-1", protocol: "http", port: "80", enabled: true,
	})
	if svc.Targets[0].Options != nil {
		t.Errorf("options should be nil when no option fields set, got %+v", svc.Targets[0].Options)
	}
}

// ─── Summary helpers ────────────────────────────────────────────────

func TestSvcHasAnyAuth(t *testing.T) {
	tests := []struct {
		name string
		auth *models.ReverseProxyAuth
		want bool
	}{
		{"nil", nil, false},
		{"empty", &models.ReverseProxyAuth{}, false},
		{"password disabled", &models.ReverseProxyAuth{PasswordAuth: &models.ReverseProxyPasswordAuth{Enabled: false}}, false},
		{"password enabled", &models.ReverseProxyAuth{PasswordAuth: &models.ReverseProxyPasswordAuth{Enabled: true}}, true},
		{"pin enabled", &models.ReverseProxyAuth{PinAuth: &models.ReverseProxyPinAuth{Enabled: true}}, true},
		{"bearer enabled", &models.ReverseProxyAuth{BearerAuth: &models.ReverseProxyBearerAuth{Enabled: true}}, true},
		{"link enabled", &models.ReverseProxyAuth{LinkAuth: &models.ReverseProxyLinkAuth{Enabled: true}}, true},
		{"one disabled header", &models.ReverseProxyAuth{HeaderAuths: []models.ReverseProxyHeaderAuth{{Enabled: false}}}, false},
		{"one enabled header", &models.ReverseProxyAuth{HeaderAuths: []models.ReverseProxyHeaderAuth{{Enabled: false}, {Enabled: true}}}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := svcHasAnyAuth(tc.auth); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSvcHasAnyAccess(t *testing.T) {
	tests := []struct {
		name string
		r    *models.ReverseProxyAccessRestrictions
		want bool
	}{
		{"nil", nil, false},
		{"empty", &models.ReverseProxyAccessRestrictions{}, false},
		{"allowed cidr", &models.ReverseProxyAccessRestrictions{AllowedCIDRs: []string{"10.0.0.0/24"}}, true},
		{"blocked country", &models.ReverseProxyAccessRestrictions{BlockedCountries: []string{"KP"}}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := svcHasAnyAccess(tc.r); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasAnyOpts(t *testing.T) {
	tests := []struct {
		name string
		opts *models.ReverseProxyTargetOpts
		want bool
	}{
		{"nil", nil, false},
		{"empty", &models.ReverseProxyTargetOpts{}, false},
		{"skip tls", &models.ReverseProxyTargetOpts{SkipTLSVerify: true}, true},
		{"timeout", &models.ReverseProxyTargetOpts{RequestTimeout: "30s"}, true},
		{"headers", &models.ReverseProxyTargetOpts{CustomHeaders: map[string]string{"X": "y"}}, true},
		{"proxy protocol", &models.ReverseProxyTargetOpts{ProxyProtocol: true}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasAnyOpts(tc.opts); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultProtocolForMode(t *testing.T) {
	tests := []struct {
		mode string
		want string
	}{
		{models.ReverseProxyModeHTTP, "http"},
		{models.ReverseProxyModeTCP, "tcp"},
		{models.ReverseProxyModeTLS, "tcp"},
		{models.ReverseProxyModeUDP, "udp"},
		{"", "http"}, // fallback
	}
	for _, tc := range tests {
		t.Run(tc.mode, func(t *testing.T) {
			if got := defaultProtocolForMode(tc.mode); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRPAuthSummary(t *testing.T) {
	tests := []struct {
		name string
		auth *models.ReverseProxyAuth
		want string
	}{
		{"nil", nil, "-"},
		{"empty", &models.ReverseProxyAuth{}, "-"},
		{"pwd only", &models.ReverseProxyAuth{PasswordAuth: &models.ReverseProxyPasswordAuth{Enabled: true}}, "pwd"},
		{
			"pwd+pin+sso+link+headers",
			&models.ReverseProxyAuth{
				PasswordAuth: &models.ReverseProxyPasswordAuth{Enabled: true},
				PinAuth:      &models.ReverseProxyPinAuth{Enabled: true},
				BearerAuth:   &models.ReverseProxyBearerAuth{Enabled: true},
				LinkAuth:     &models.ReverseProxyLinkAuth{Enabled: true},
				HeaderAuths:  []models.ReverseProxyHeaderAuth{{Enabled: true}, {Enabled: true}},
			},
			"pwd,pin,sso,link,hdr×2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := rpAuthSummary(tc.auth); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRPAccessSummary(t *testing.T) {
	tests := []struct {
		name string
		r    *models.ReverseProxyAccessRestrictions
		want string
	}{
		{"nil", nil, "-"},
		{"empty", &models.ReverseProxyAccessRestrictions{}, "-"},
		{"one rule", &models.ReverseProxyAccessRestrictions{AllowedCIDRs: []string{"10.0.0.0/24"}}, "1 rule"},
		{
			"multiple rules",
			&models.ReverseProxyAccessRestrictions{
				AllowedCIDRs:     []string{"10.0.0.0/24"},
				BlockedCountries: []string{"KP", "CN"},
			},
			"3 rules",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := rpAccessSummary(tc.r); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// ─── listenPortValidator (I4) ───────────────────────────────────────

func TestListenPortValidator_EmptyAlwaysValid(t *testing.T) {
	data := &rpServiceFormData{proxyCluster: "cluster-1"}
	clusters := map[string]models.ReverseProxyCluster{
		"cluster-1": {ID: "cluster-1", Name: "c1", SupportsCustomPorts: false},
	}
	if err := listenPortValidator(data, clusters)(""); err != nil {
		t.Errorf("empty port should always pass: %v", err)
	}
}

func TestListenPortValidator_RejectsCustomPortOnNonSupportingCluster(t *testing.T) {
	data := &rpServiceFormData{proxyCluster: "cluster-1"}
	clusters := map[string]models.ReverseProxyCluster{
		"cluster-1": {ID: "cluster-1", Name: "restricted", SupportsCustomPorts: false},
	}
	err := listenPortValidator(data, clusters)("8443")
	if err == nil {
		t.Fatal("expected rejection, got nil")
	}
}

func TestListenPortValidator_AcceptsCustomPortOnSupportingCluster(t *testing.T) {
	data := &rpServiceFormData{proxyCluster: "cluster-1"}
	clusters := map[string]models.ReverseProxyCluster{
		"cluster-1": {ID: "cluster-1", Name: "flexible", SupportsCustomPorts: true},
	}
	if err := listenPortValidator(data, clusters)("8443"); err != nil {
		t.Errorf("custom port should be allowed: %v", err)
	}
}

func TestListenPortValidator_UnknownClusterDefers(t *testing.T) {
	data := &rpServiceFormData{proxyCluster: "unknown"}
	clusters := map[string]models.ReverseProxyCluster{}
	if err := listenPortValidator(data, clusters)("8443"); err != nil {
		t.Errorf("unknown cluster should defer to server: %v", err)
	}
}

func TestListenPortValidator_InvalidPortRejectedFirst(t *testing.T) {
	data := &rpServiceFormData{proxyCluster: "cluster-1"}
	clusters := map[string]models.ReverseProxyCluster{
		"cluster-1": {ID: "cluster-1", Name: "flexible", SupportsCustomPorts: true},
	}
	if err := listenPortValidator(data, clusters)("99999"); err == nil {
		t.Fatal("out-of-range port should be rejected before cluster check")
	}
}

// ─── ReverseProxyUpdateFromService ──────────────────────────────────

func TestReverseProxyUpdateFromService_DropsIDAndMeta(t *testing.T) {
	svc := models.ReverseProxyService{
		ID:               "svc-1",
		Name:             "api",
		Domain:           "api.example.com",
		Mode:             models.ReverseProxyModeHTTP,
		ListenPort:       443,
		ProxyCluster:     "cluster-1",
		Enabled:          true,
		PassHostHeader:   true,
		RewriteRedirects: true,
		Targets:          []models.ReverseProxyTarget{{TargetType: "peer", Port: 8080}},
		Meta:             &models.ReverseProxyMeta{Status: "active"},
	}
	req := ReverseProxyUpdateFromService(svc)
	if req.Name != svc.Name || req.ListenPort != svc.ListenPort || req.ProxyCluster != svc.ProxyCluster {
		t.Errorf("core fields not copied: %+v", req)
	}
	if !req.Enabled || !req.PassHostHeader || !req.RewriteRedirects {
		t.Errorf("bool flags not copied: %+v", req)
	}
	if len(req.Targets) != 1 || req.Targets[0].Port != 8080 {
		t.Errorf("targets not copied: %+v", req.Targets)
	}
}
