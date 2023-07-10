package tls

import (
	"context"
	"crypto/tls"
	"net"
)

// Conn is the recommended interface for the connection
// returned by the DailTLS function (Client.SetDialTLS,
// Transport.DialTLSContext), so that the TLS handshake negotiation
// can automatically decide whether to use HTTP2 or HTTP1 (ALPN).
// If this interface is not implemented, HTTP1 will be used by default.
type Conn interface {
	net.Conn
	// ConnectionState returns basic TLS details about the connection.
	ConnectionState() tls.ConnectionState
	// Handshake runs the client or server handshake
	// protocol if it has not yet been run.
	//
	// Most uses of this package need not call Handshake explicitly: the
	// first Read or Write will call it automatically.
	//
	// For control over canceling or setting a timeout on a handshake, use
	// HandshakeContext or the Dialer's DialContext method instead.
	Handshake() error

	// HandshakeContext runs the client or server handshake
	// protocol if it has not yet been run.
	//
	// The provided Context must be non-nil. If the context is canceled before
	// the handshake is complete, the handshake is interrupted and an error is returned.
	// Once the handshake has completed, cancellation of the context will not affect the
	// connection.
	//
	// Most uses of this package need not call HandshakeContext explicitly: the
	// first Read or Write will call it automatically.
	HandshakeContext(ctx context.Context) error
}
