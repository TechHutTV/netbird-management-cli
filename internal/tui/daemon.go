package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/netbirdio/netbird/client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// DefaultDaemonSocket is the default path to the NetBird daemon socket
	DefaultDaemonSocket = "unix:///var/run/netbird.sock"

	// daemonTimeout is the timeout for daemon gRPC calls
	daemonTimeout = 3 * time.Second
)

// DaemonClient wraps the NetBird daemon gRPC connection
type DaemonClient struct {
	conn   *grpc.ClientConn
	daemon proto.DaemonServiceClient
}

// NewDaemonClient connects to the NetBird daemon via gRPC.
// Returns nil (not an error) if the connection cannot be established,
// allowing the TUI to run in management-only mode.
func NewDaemonClient(socketPath string) *DaemonClient {
	if socketPath == "" {
		socketPath = DefaultDaemonSocket
	}

	conn, err := grpc.NewClient(socketPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	return &DaemonClient{
		conn:   conn,
		daemon: proto.NewDaemonServiceClient(conn),
	}
}

// Status fetches the full status from the local daemon
func (d *DaemonClient) Status() (*proto.StatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	return d.daemon.Status(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
}

// GetConfig fetches the daemon configuration
func (d *DaemonClient) GetConfig() (*proto.GetConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	return d.daemon.GetConfig(ctx, &proto.GetConfigRequest{})
}

// PeerConnections fetches per-peer connection details from the local daemon
func (d *DaemonClient) PeerConnections() (map[string]PeerConnectionInfo, error) {
	if d == nil {
		return nil, fmt.Errorf("daemon not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	status, err := d.daemon.Status(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
	if err != nil {
		return nil, fmt.Errorf("daemon status: %w", err)
	}

	conns := make(map[string]PeerConnectionInfo)
	if status.GetFullStatus() != nil {
		for _, peer := range status.GetFullStatus().GetPeers() {
			latency := ""
			if peer.GetLatency() != nil {
				latency = peer.GetLatency().AsDuration().String()
			}
			handshake := ""
			if peer.GetLastWireguardHandshake() != nil {
				handshake = peer.GetLastWireguardHandshake().AsTime().Format(time.RFC3339)
			}
			info := PeerConnectionInfo{
				RemoteEndpoint: peer.GetRemoteIceCandidateEndpoint(),
				LocalICEType:   peer.GetLocalIceCandidateType(),
				RemoteICEType:  peer.GetRemoteIceCandidateType(),
				Latency:        latency,
				BytesSent:      peer.GetBytesTx(),
				BytesReceived:  peer.GetBytesRx(),
				LastHandshake:  handshake,
				RelayAddress:   peer.GetRelayAddress(),
			}
			if peer.GetConnStatus() == "Connected" {
				if peer.GetRelayed() {
					info.ConnType = "Relayed"
				} else {
					info.ConnType = "P2P"
				}
			} else {
				info.ConnType = "Disconnected"
			}
			conns[peer.GetIP()] = info
		}
	}
	return conns, nil
}

// Close closes the gRPC connection
func (d *DaemonClient) Close() {
	if d != nil && d.conn != nil {
		d.conn.Close()
	}
}
