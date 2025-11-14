# **NetBird Management CLI (netbird-manage)**

netbird-manage is an unofficial command-line tool written in Go for interacting with the [NetBird](https://netbird.io/) API. It allows you to quickly manage peers, groups, policies, and other network resources directly from your terminal.  
This tool is built based on the official NetBird API documentation.

## **Setup & Installation**

### **Prerequisites**

You must have the [Go toolchain](https://go.dev/doc/install) (version 1.18 or later) installed on your system.

### **1\. Initial Setup**

Place all .go files from this project into a new directory (e.g., netbird-manage).  
From inside that directory, initialize the Go module:  
go mod init netbird-manage

### **2\. Build the Executable**

Run the go build command. This will find all package main files in the directory, compile them, and create a single executable.  
go build

You will now have an executable file named netbird-manage (or netbird-manage.exe on Windows) in your directory.

### **3\. Connect Your Account**

Before you can use the tool, you must authenticate. This tool stores your API token in a configuration file at $HOME/.netbird-manage.conf.  
Generate a Personal Access Token (PAT) or a Service User token from your NetBird dashboard. Then, run the connect command:  
./netbird-manage connect \--token \<your-netbird-token-here\>

If successful, you will see a "Connection successful" message. The tool is now ready to use.

## **Current Commands & Functionality**

### **connect**

Saves your API token to the local config file after testing it.  
netbird-manage connect \--token \<api\_token\>

### **peer**

Manage network peers. Running netbird-manage peer by itself will display the help menu.

| Command | Description |
| :---- | :---- |
| netbird-manage peer \--list | List all peers in your network. |
| netbird-manage peer \--inspect \<peer-id\> | View detailed information for a single peer. |
| netbird-manage peer \--remove \<peer-id\> | Remove a peer from your network. |
| netbird-manage peer \--edit \<peer-id\> \--add-group \<group-name\> | Add a peer to a specified group. |
| netbird-manage peer \--edit \<peer-id\> \--remove-group \<group-name\> | Remove a peer from a specified group. |

### **group**

Manage peer groups.  
netbird-manage group

* Lists all available groups in your account.

### **networks**

Manage networks.  
netbird-manage networks

* Lists all configured networks.

### **policy**

Manage access control policies.  
netbird-manage policy

* Lists all access control policies and their rules.

## **🚀 Future Plans**

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
