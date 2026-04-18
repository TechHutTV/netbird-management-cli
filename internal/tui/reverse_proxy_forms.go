package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/huh/v2"

	"netbird-manage/internal/models"
)

// ─── Wizard form data structs ───────────────────────────────────────
//
// Each leaf wizard screen uses a dedicated form-data struct so huh.Form can bind
// directly to string/bool fields. Helpers serialize the struct into/out of the
// models.ReverseProxyService being assembled.

type rpServiceFormData struct {
	name         string
	domain       string
	listenPort   string
	mode         string
	proxyCluster string
	enabled      bool
}

type rpTargetFormData struct {
	targetType string // "peer" | "host" | "domain" | "subnet"
	targetID   string // peer id when targetType == "peer"
	host       string // free-text host when targetType != "peer"
	protocol   string // "http" | "https" | "tcp" | "udp"
	port       string
	path       string
	enabled    bool

	skipTLSVerify      bool
	requestTimeout     string
	sessionIdleTimeout string
	proxyProtocol      bool
	accessLocal        bool
}

type rpAccessRuleFormData struct {
	action string // "allow" | "block"
	kind   string // "country" | "cidr" | "ip"
	value  string
}

type rpAdvancedFormData struct {
	// HTTP
	passHostHeader   bool
	rewriteRedirects bool
	// L4 (TCP/TLS)
	proxyProtocol  bool
	requestTimeout string
	// UDP
	sessionIdleTimeout string
}

type rpPasswordFormData struct {
	enabled  bool
	password string
}

type rpPinFormData struct {
	enabled bool
	pin     string
}

type rpBearerFormData struct {
	enabled            bool
	distributionGroups []string
}

type rpLinkFormData struct {
	enabled bool
}

type rpHeaderAuthFormData struct {
	enabled bool
	header  string
	value   string
}

type rpCustomHeaderFormData struct {
	header string
	value  string
}

type rpDomainFormData struct {
	domain        string
	targetCluster string
}

// ─── Form constructors ──────────────────────────────────────────────

func newRPServiceForm(data *rpServiceFormData, clusters []models.ReverseProxyCluster, lockMode bool) *huh.Form {
	modeOptions := []huh.Option[string]{
		huh.NewOption("HTTP", models.ReverseProxyModeHTTP),
		huh.NewOption("TCP", models.ReverseProxyModeTCP),
		huh.NewOption("UDP", models.ReverseProxyModeUDP),
		huh.NewOption("TLS", models.ReverseProxyModeTLS),
	}

	// Build a cluster lookup so the listen-port validator can see the currently
	// selected cluster's supports_custom_ports flag at submit time.
	clusterByID := make(map[string]models.ReverseProxyCluster, len(clusters))
	for _, cl := range clusters {
		clusterByID[cl.ID] = cl
	}

	clusterOptions := []huh.Option[string]{huh.NewOption("(default)", "")}
	for _, cl := range clusters {
		clusterOptions = append(clusterOptions, huh.NewOption(cl.Name, cl.ID))
	}

	fields := []huh.Field{
		huh.NewInput().
			Title("Service Name").
			Placeholder("e.g. internal-api").
			Validate(nonBlank("name")).
			Value(&data.name),
		huh.NewInput().
			Title("Domain").
			Description("Full domain (with subdomain if required), e.g. api.example.com").
			Validate(nonBlank("domain")).
			Value(&data.domain),
	}

	if !lockMode {
		fields = append(fields, huh.NewSelect[string]().
			Title("Mode").
			Description("HTTP for web traffic, TCP/UDP/TLS for raw L4 forwarding. Cannot be changed later.").
			Options(modeOptions...).
			Value(&data.mode))
	}

	fields = append(fields,
		huh.NewSelect[string]().
			Title("Proxy Cluster").
			Description("Leave default unless you run multiple clusters.").
			Options(clusterOptions...).
			Value(&data.proxyCluster),
		huh.NewInput().
			Title("Listen Port").
			Description("Leave empty for auto-assigned. For L4 modes, some clusters require the default port — setting a custom port will be rejected on those clusters.").
			Placeholder("443").
			Validate(listenPortValidator(data, clusterByID)).
			Value(&data.listenPort),
		huh.NewConfirm().Title("Enabled").Value(&data.enabled),
	)

	return huh.NewForm(huh.NewGroup(fields...))
}

// listenPortValidator combines the port-range check with a cluster-capability gate:
// if the selected cluster reports supports_custom_ports=false, reject any
// non-empty listen port. Validator closes over the form data so it sees the
// latest cluster selection at submit time.
func listenPortValidator(data *rpServiceFormData, clusterByID map[string]models.ReverseProxyCluster) func(string) error {
	portCheck := optionalPort("listen port")
	return func(s string) error {
		if err := portCheck(s); err != nil {
			return err
		}
		if strings.TrimSpace(s) == "" {
			return nil
		}
		cl, ok := clusterByID[data.proxyCluster]
		if !ok {
			return nil // unknown cluster (default) — let the server decide
		}
		if !cl.SupportsCustomPorts {
			return fmt.Errorf("cluster %q does not support custom listen ports — leave empty", cl.Name)
		}
		return nil
	}
}

func newRPTargetForm(data *rpTargetFormData, peers map[string]string, mode string) *huh.Form {
	typeOptions := []huh.Option[string]{
		huh.NewOption("Peer", "peer"),
		huh.NewOption("Host (raw address)", "host"),
		huh.NewOption("Domain", "domain"),
		huh.NewOption("Subnet", "subnet"),
	}

	protocolOptions := []huh.Option[string]{}
	switch mode {
	case models.ReverseProxyModeHTTP:
		protocolOptions = []huh.Option[string]{
			huh.NewOption("HTTP", "http"),
			huh.NewOption("HTTPS", "https"),
		}
	case models.ReverseProxyModeTCP, models.ReverseProxyModeTLS:
		protocolOptions = []huh.Option[string]{huh.NewOption("TCP", "tcp")}
	case models.ReverseProxyModeUDP:
		protocolOptions = []huh.Option[string]{huh.NewOption("UDP", "udp")}
	default:
		protocolOptions = []huh.Option[string]{
			huh.NewOption("HTTP", "http"),
			huh.NewOption("HTTPS", "https"),
			huh.NewOption("TCP", "tcp"),
			huh.NewOption("UDP", "udp"),
		}
	}

	peerOptions := []huh.Option[string]{huh.NewOption("(select a peer)", "")}
	for id, name := range peers {
		peerOptions = append(peerOptions, huh.NewOption(name, id))
	}

	fields := []huh.Field{
		huh.NewSelect[string]().
			Title("Target Type").
			Options(typeOptions...).
			Value(&data.targetType),
		huh.NewSelect[string]().
			Title("Peer").
			Description("Required when target type is 'peer'; ignored otherwise.").
			Options(peerOptions...).
			Validate(func(id string) error {
				if data.targetType == "peer" && strings.TrimSpace(id) == "" {
					return fmt.Errorf("pick a peer or switch target type")
				}
				return nil
			}).
			Value(&data.targetID),
		huh.NewInput().
			Title("Host").
			Description("IP, hostname or subnet. Required when target type is host/domain/subnet.").
			Placeholder("192.168.1.10 or api.internal.example.com").
			Validate(func(s string) error {
				if data.targetType != "peer" && strings.TrimSpace(s) == "" {
					return fmt.Errorf("host is required for target type %s", data.targetType)
				}
				return nil
			}).
			Value(&data.host),
		huh.NewSelect[string]().
			Title("Protocol").
			Options(protocolOptions...).
			Value(&data.protocol),
		huh.NewInput().
			Title("Port").
			Description("Backend port on the target (1-65535). 0 = default for protocol.").
			Validate(optionalPort("port")).
			Value(&data.port),
	}

	if mode == models.ReverseProxyModeHTTP {
		fields = append(fields,
			huh.NewInput().
				Title("Path").
				Description("Path prefix to route to this target (HTTP only). Must start with /.").
				Placeholder("/").
				Validate(httpPath).
				Value(&data.path),
			huh.NewConfirm().
				Title("Skip TLS Verification").
				Description("Don't verify the upstream certificate (HTTPS targets only).").
				Value(&data.skipTLSVerify),
			huh.NewInput().
				Title("Request Timeout").
				Description("Go duration, e.g. 30s. Empty = no timeout.").
				Placeholder("30s").
				Validate(optionalDuration("request timeout")).
				Value(&data.requestTimeout),
		)
	}

	if mode == models.ReverseProxyModeUDP {
		fields = append(fields, huh.NewInput().
			Title("Session Idle Timeout").
			Description("Go duration, e.g. 5m. Empty = no timeout.").
			Placeholder("5m").
			Validate(optionalDuration("session idle timeout")).
			Value(&data.sessionIdleTimeout))
	}

	if mode == models.ReverseProxyModeTCP || mode == models.ReverseProxyModeTLS {
		fields = append(fields,
			huh.NewInput().
				Title("Request Timeout").
				Description("Go duration, e.g. 30s. Empty = no timeout.").
				Placeholder("30s").
				Validate(optionalDuration("request timeout")).
				Value(&data.requestTimeout),
			huh.NewConfirm().
				Title("Proxy Protocol").
				Description("Preserve source IP using PROXY protocol v2.").
				Value(&data.proxyProtocol),
		)
	}

	fields = append(fields,
		huh.NewConfirm().Title("Enabled").Value(&data.enabled),
		huh.NewConfirm().
			Title("Access Local").
			Description("Allow requests originating from the same NetBird network.").
			Value(&data.accessLocal),
	)

	return huh.NewForm(huh.NewGroup(fields...))
}

func newRPAccessRuleForm(data *rpAccessRuleFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Action").
			Options(
				huh.NewOption("Allow Only", "allow"),
				huh.NewOption("Block Only", "block"),
			).
			Value(&data.action),
		huh.NewSelect[string]().
			Title("Type").
			Options(
				huh.NewOption("Country", "country"),
				huh.NewOption("IP Address", "ip"),
				huh.NewOption("CIDR Block", "cidr"),
			).
			Value(&data.kind),
		huh.NewInput().
			Title("Value").
			Description("Country: 2-letter ISO code (e.g. KP). IP: 192.168.1.1. CIDR: 10.0.0.0/24.").
			Validate(nonBlank("value")).
			Value(&data.value),
	))
}

func newRPAdvancedForm(data *rpAdvancedFormData, mode string) *huh.Form {
	var fields []huh.Field
	switch mode {
	case models.ReverseProxyModeHTTP:
		fields = []huh.Field{
			huh.NewConfirm().
				Title("Pass Host Header").
				Description("Forward the original Host header to targets.").
				Value(&data.passHostHeader),
			huh.NewConfirm().
				Title("Rewrite Redirects").
				Description("Rewrite Location headers in redirects to stay on the proxy domain.").
				Value(&data.rewriteRedirects),
		}
	case models.ReverseProxyModeTCP, models.ReverseProxyModeTLS:
		fields = []huh.Field{
			huh.NewConfirm().
				Title("Proxy Protocol").
				Description("Preserve client source IP using PROXY protocol v2.").
				Value(&data.proxyProtocol),
			huh.NewInput().
				Title("Connection Timeout").
				Placeholder("30s").
				Validate(optionalDuration("connection timeout")).
				Value(&data.requestTimeout),
		}
	case models.ReverseProxyModeUDP:
		fields = []huh.Field{
			huh.NewInput().
				Title("Session Idle Timeout").
				Placeholder("5m").
				Validate(optionalDuration("session idle timeout")).
				Value(&data.sessionIdleTimeout),
		}
	}
	if len(fields) == 0 {
		fields = []huh.Field{huh.NewNote().Title("No advanced settings for this mode.")}
	}
	return huh.NewForm(huh.NewGroup(fields...))
}

func newRPPasswordForm(data *rpPasswordFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Enable Password Auth").Value(&data.enabled),
		huh.NewInput().
			Title("Password").
			Description("Stored securely on the server.").
			EchoMode(huh.EchoModePassword).
			Value(&data.password),
	))
}

func newRPPinForm(data *rpPinFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Enable PIN Auth").Value(&data.enabled),
		huh.NewInput().
			Title("PIN").
			Description("Numeric PIN, typically 4-8 digits.").
			Validate(optionalNumeric("PIN")).
			Value(&data.pin),
	))
}

func newRPBearerForm(data *rpBearerFormData, groups map[string]string) *huh.Form {
	opts := make([]huh.Option[string], 0, len(groups))
	for id, name := range groups {
		opts = append(opts, huh.NewOption(name, id))
	}
	return huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Enable SSO (Bearer) Auth").Value(&data.enabled),
		huh.NewMultiSelect[string]().
			Title("Distribution Groups").
			Description("User groups allowed to access this service.").
			Options(opts...).
			Value(&data.distributionGroups),
	))
}

func newRPLinkForm(data *rpLinkFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Enable Magic Link Auth").
			Description("Users receive a one-time link by email to access the service.").
			Value(&data.enabled),
	))
}

func newRPHeaderAuthForm(data *rpHeaderAuthFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Enabled").Value(&data.enabled),
		huh.NewInput().
			Title("Header Name").
			Placeholder("X-Auth-Token").
			Validate(nonBlank("header")).
			Value(&data.header),
		huh.NewInput().
			Title("Expected Value").
			Validate(nonBlank("value")).
			Value(&data.value),
	))
}

func newRPCustomHeaderForm(data *rpCustomHeaderFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Header Name").
			Validate(nonBlank("header")).
			Value(&data.header),
		huh.NewInput().
			Title("Value").
			Value(&data.value),
	))
}

func newRPDomainForm(data *rpDomainFormData) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Custom Domain").
			Description("Fully-qualified domain you own, e.g. api.example.com.").
			Validate(nonBlank("domain")).
			Value(&data.domain),
		huh.NewInput().
			Title("Target Cluster").
			Description("Cluster ID to bind this domain to. Leave empty for default.").
			Value(&data.targetCluster),
	))
}

// ─── Validators ─────────────────────────────────────────────────────

func nonBlank(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func optionalPort(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 || n > 65535 {
			return fmt.Errorf("%s must be 0-65535", field)
		}
		return nil
	}
}

func optionalDuration(field string) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if _, err := time.ParseDuration(s); err != nil {
			return fmt.Errorf("%s must be a Go duration (e.g. 30s, 5m, 1h)", field)
		}
		return nil
	}
}

func optionalNumeric(field string) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				return fmt.Errorf("%s must be numeric", field)
			}
		}
		return nil
	}
}

func httpPath(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if !strings.HasPrefix(s, "/") {
		return fmt.Errorf("path must start with /")
	}
	return nil
}
