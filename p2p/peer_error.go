package p2p

import (
	"errors"
	"fmt"
)

const (
	errInvalidMsgCode = iota
	errInvalidMsg
)

var errorToString = map[int]string{
	errInvalidMsgCode: "invalid message code",
	errInvalidMsg:     "invalid message",
}

const (
	DiscRequested DiscReason = iota
	DiscNetworkError
	DiscProtocolError
	DiscUselessPeer
	DiscTooManyPeers
	DiscTooManyInboundPeers
	DiscAlreadyConnected
	DiscIncompatibleVersion
	DiscInvalidIdentity
	DiscQuitting
	DiscUnexpectedIdentity
	DiscSelf
	DiscReadTimeout
	DiscSubprotocolError = 0x10
)

var discReasonToString = [...]string{
	DiscRequested:           "disconnect requested",
	DiscNetworkError:        "network error",
	DiscProtocolError:       "breach of protocol",
	DiscUselessPeer:         "useless peer",
	DiscTooManyPeers:        "too many peers",
	DiscTooManyInboundPeers: "too many inbound peers",
	DiscAlreadyConnected:    "already connected",
	DiscIncompatibleVersion: "incompatible p2p protocol version",
	DiscInvalidIdentity:     "invalid node identity",
	DiscQuitting:            "client quitting",
	DiscUnexpectedIdentity:  "unexpected identity",
	DiscSelf:                "connected to self",
	DiscReadTimeout:         "read timeout",
	DiscSubprotocolError:    "subprotocol error",
}

var errProtocolReturned = errors.New("protocol returned")

type DiscReason uint

func (d DiscReason) String() string {
	if len(discReasonToString) < int(d) {
		return fmt.Sprintf("unknown disconnect reason %d", d)
	}
	return discReasonToString[d]
}

func (d DiscReason) Error() string {
	return d.String()
}

func discReasonForError(err error) DiscReason {
	if reason, ok := err.(DiscReason); ok {
		return reason
	}
	if err == errProtocolReturned {
		return DiscQuitting
	}
	peerError, ok := err.(*peerError)
	if ok {
		switch peerError.code {
		case errInvalidMsgCode, errInvalidMsg:
			return DiscProtocolError
		default:
			return DiscSubprotocolError
		}
	}
	return DiscSubprotocolError
}

type peerError struct {
	code    int
	message string
}

func (pe *peerError) Error() string {
	return pe.message
}

func newPeerError(code int, format string, v ...interface{}) *peerError {
	desc, ok := errorToString[code]
	if !ok {
		panic("invalid error code")
	}
	err := &peerError{code, desc}
	if format != "" {
		err.message += ": " + fmt.Sprintf(format, v...)
	}
	return err
}
