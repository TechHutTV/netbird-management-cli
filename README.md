# NetBird Management CLI

netbird-manage is an unofficial command-line tool written in Go for interacting with the [NetBird](https://netbird.io/) API. It allows you to quickly manage peers, groups, policies, and other network resources directly from your terminal. This tool is built based on the official [NetBird REST API documentation](https://docs.netbird.io/api).

![](https://github.com/TechHutTV/netbird-management-cli/blob/main/demo.png)

## Quick Start

### Installation

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system. Clone this repository and build the binary:

```bash
git clone https://github.com/TechHutTV/netbird-management-cli.git
cd netbird-management-cli
go build ./cmd/netbird-manage
```

This creates a `netbird-manage` binary in the current directory. Optionally, move it to a directory in your PATH:

```bash
sudo mv netbird-manage /usr/local/bin/
```

### Connect

Before you can use the tool, you must authenticate. Generate a Personal Access Token (PAT) or a Service User token from your NetBird dashboard. Then, run the connect command:

```bash
netbird-manage connect --token <token>
```

If successful, you will see a "Connection successful" message. To check status or change api url:

```
netbird-manage connect          Check current connection status
  connect [flags]               Connect and save your API token
    --token <token>             (Required) Your NetBird API token
    --management-url <url>      (Optional) Your self-hosted management URL
```

### Help

Run any command without flags to see available options:

```bash
netbird-manage peer           # Shows peer command help
netbird-manage group          # Shows group command help
netbird-manage --help         # Shows all available commands
```

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, safety features, debug mode, batch operations |
| [Peers](docs/peers.md) | Manage network peers |
| [Setup Keys](docs/setup-keys.md) | Device registration and onboarding keys |
| [Users](docs/users.md) | User management and invitations |
| [Tokens](docs/tokens.md) | Personal access token management |
| [Groups](docs/groups.md) | Peer group management |
| [Networks](docs/networks.md) | Networks, resources, and routers |
| [Policies](docs/policies.md) | Access control policies and firewall rules |
| [Routes](docs/routes.md) | Network routing configuration |
| [DNS](docs/dns.md) | DNS nameserver groups and settings |
| [Posture Checks](docs/posture-checks.md) | Device compliance validation |
| [Events](docs/events.md) | Audit logs and traffic monitoring |
| [Geo-Locations](docs/geo-locations.md) | Geographic location data |
| [Accounts](docs/accounts.md) | Account settings and configuration |
| [Ingress Ports](docs/ingress-ports.md) | Port forwarding (Cloud-only) |
| [Export & Import](docs/export-import.md) | YAML/JSON configuration management |
| [Migrate](docs/migrate.md) | Migration between NetBird accounts |

## API Coverage

**14/14 NetBird API resource types fully implemented (100%)**

| Resource | Status |
|----------|--------|
| Peers | Full CRUD |
| Groups | Full CRUD |
| Networks | Full CRUD |
| Policies | Full CRUD |
| Setup Keys | Full CRUD |
| Users | Full CRUD |
| Tokens | Full CRUD |
| Routes | Full CRUD |
| DNS | Full CRUD |
| Posture Checks | Full CRUD |
| Events | Read |
| Geo-Locations | Read |
| Accounts | Full CRUD |
| Ingress Ports | Full CRUD (Cloud-only) |

## Roadmap

### Planned Features

- **Shell Completion** - Tab completion for bash/zsh/fish

For detailed implementation notes and architecture guidance, see [CLAUDE.md](CLAUDE.md).

## License

MIT/Apache dual license - see [LICENSE](LICENSE) file.
