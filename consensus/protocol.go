package consensus

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// Broadcaster defines the interface to enqueue blocks to fetcher, find peer
type Broadcaster interface {
	// FindPeers retrieves peers by addresses
	FindPeers(targets map[enode.ID]bool, purpose PurposeFlag) map[enode.ID]Peer
}

// P2PServer defines the interface for a p2p.server to get the local node's enode and to add/remove for static/trusted peers
type P2PServer interface {
	// Gets this node's enode
	Self() *enode.Node
	// AddPeer will add a peer to the p2p server instance
	AddPeer(node *enode.Node, purpose PurposeFlag)
	// RemovePeer will remove a peer from the p2p server instance
	RemovePeer(node *enode.Node, purpose PurposeFlag)
	// AddTrustedPeer will add a trusted peer to the p2p server instance
	AddTrustedPeer(node *enode.Node, purpose PurposeFlag)
	// RemoveTrustedPeer will remove a trusted peer from the p2p server instance
	RemoveTrustedPeer(node *enode.Node, purpose PurposeFlag)
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
	PurposeIsSet(purpose PurposeFlag) bool
}

const (
	NoPurpose              PurposeFlag = 0
	ExplicitStaticPurpose              = 1 << 0
	ExplicitTrustedPurpose             = 1 << 1
	ValidatorPurpose                   = 1 << 2
	ProxyPurpose                       = 1 << 3
	AnyPurpose                         = ExplicitStaticPurpose | ExplicitTrustedPurpose | ValidatorPurpose | ProxyPurpose // This value should be the bitwise OR of all possible PurposeFlag values
)

// Note that this type is NOT threadsafe.  The reason that it is not is that it's read and written
// only by the p2p server's single threaded event loop.
type PurposeFlag uint32

func (pf PurposeFlag) Add(f PurposeFlag) PurposeFlag {
	return pf | f
}

func (pf PurposeFlag) Remove(f PurposeFlag) PurposeFlag {
	return pf & ^f
}

func (pf PurposeFlag) IsSet(f PurposeFlag) bool {
	return (pf & f) != 0
}

func (pf PurposeFlag) HasNoPurpose() bool {
	return pf == NoPurpose
}

func (pf PurposeFlag) HasPurpose() bool {
	return pf != NoPurpose
}

func (pf PurposeFlag) String() string {
	s := ""
	if pf.IsSet(ExplicitStaticPurpose) {
		s += "-ExplicitStaticPurpose"
	}
	if pf.IsSet(ExplicitTrustedPurpose) {
		s += "-ExplicitTrustedPurpose"
	}
	if pf.IsSet(ValidatorPurpose) {
		s += "-ValidatorPurpose"
	}
	if pf.IsSet(ProxyPurpose) {
		s += "-ProxyPurpose"
	}
	if s != "" {
		s = s[1:]
	}
	return s
}
