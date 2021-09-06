package consensus

import (
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/mapprotocol/atlas/p2p"
)

// Broadcaster defines the interface to enqueue blocks to fetcher, find peer
type Broadcaster interface {
	// FindPeers retrieves peers by addresses
	FindPeers(targets map[enode.ID]bool, purpose p2p.PurposeFlag) map[enode.ID]Peer
}

// P2PServer defines the interface for a p2p.server to get the local node's enode and to add/remove for static/trusted peers
type P2PServer interface {
	// Gets this node's enode
	Self() *enode.Node
	// AddPeer will add a peer to the p2p server instance
	AddPeer(node *enode.Node, purpose p2p.PurposeFlag)
	// RemovePeer will remove a peer from the p2p server instance
	RemovePeer(node *enode.Node, purpose p2p.PurposeFlag)
	// AddTrustedPeer will add a trusted peer to the p2p server instance
	AddTrustedPeer(node *enode.Node, purpose p2p.PurposeFlag)
	// RemoveTrustedPeer will remove a trusted peer from the p2p server instance
	RemoveTrustedPeer(node *enode.Node, purpose p2p.PurposeFlag)
}

// Peer defines the interface for a p2p.peer
type Peer interface {
	// Send sends the message to this peer
	Send(msgcode uint64, data interface{}) error
	// Node returns the peer's enode
	Node() *enode.Node
	// Version returns the peer's version
	Version() int
	// Blocks until a message is read directly from the peer.
	// This should only be used during a handshake.
	ReadMsg() (p2p.Msg, error)
	// Inbound returns if the peer connection is inbound
	Inbound() bool
	// PurposeIsSet returns if the peer has a purpose set
	PurposeIsSet(purpose p2p.PurposeFlag) bool
}
