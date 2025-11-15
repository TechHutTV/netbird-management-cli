# NetBird Management CLI

netbird-manage is an unofficial command-line tool written in Go for interacting with the [NetBird](https://netbird.io/) API. It allows you to quickly manage peers, groups, policies, and other network resources directly from your terminal. This tool is built based on the official [NetBird REST API documentation](https://docs.netbird.io/api).

![](https://github.com/TechHutTV/netbird-management-cli/blob/main/demo.png)

## Setup & Installation

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system. Place all .go files from this project into a new directory (e.g., netbird-manage). From inside that directory, initialize the Go module and run the go build command:  
```
go mod init netbird-manage
go build
```

## Current Commands & Functionality

### Connect
Before you can use the tool, you must authenticate. This tool stores your API token in a configuration file at `$HOME/.netbird-manage.conf`. Generate a Personal Access Token (PAT) or a Service User token from your NetBird dashboard. Then, run the connect command: 
```
netbird-manage connect --token <token>
```
If successful, you will see a "Connection successful" message. To check status or change api url see the flags below.

```
netbird-manage connect          Check current connection status
  connect [flags]               Connect and save your API token
    --token <token>             (Required) Your NetBird API token
    --management-url <url>      (Optional) Your self-hosted management URL
```

### Peer

Manage network peers. Running netbird-manage peer by itself will display the help menu.
```
netbird-manage peer              View the help page
  peer [flags]                   Managment of peers
    --list                       List all peers in your network
    --inspect <peer-id>          View detailed information for a single peer
    --remove <peer-id>           Remove a peer from your network
    --edit <peer-id>             Edit peers
      --add-group <group-id>     Add a peer to a specified group
      --remove-group <group-id>  Remove a peer from a specified group
```

### Group

Manage peer groups. Running netbird-manage group by itself will display the help menu.
```
netbird-manage group             View the help page
  group [flags]                  Management of groups
    --list                       List all groups in your network
```

### Network

Manage networks. Running netbird-manage network by itself will display the help menu.
```
netbird-manage network           View the help page
  network [flags]                Management of networks
    --list                       List all networks in your network
```

### Policy

Manage access control policies. Running netbird-manage policy by itself will display the help menu.
```
netbird-manage policy            View the help page
  policy [flags]                 Management of policies
    --list                       List all access control policies and their rules
```

## 🚀 Future Plans

This tool is in active development. The goal is to build a comprehensive and easy-to-use CLI for all NetBird management tasks.

* **Full API Coverage:** Implement the entire NetBird API, including:  
  * Full CRUD (Create, Read, Update, Delete) for **Groups**.  
  * Full CRUD for **Networks** and **Network Resources**.  
  * Full CRUD for **Policies** and **Rules**.  
  * Full **User Management** (invite, remove, update roles).  
  * Management for **Setup Keys**, **Routes**, and **DNS**.  
* **YAML-based Policy Management:**  
  * Add netbird-manage policy export \> my-policies.yml to save all policies to a file.  
  * Add netbird-manage policy apply \-f my-policies.yml to apply policy changes from a YAML file, enabling GitOps workflows.  
* **Interactive CLI Features:**  
  * Implement interactive prompts for complex operations (e.g., netbird-manage peer \--remove \<id\> asking for confirmation).  
  * Use interactive selectors (like [bubbletea](https://github.com/charmbracelet/bubbletea)) for picking peers or groups from a list.
