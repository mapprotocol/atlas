package p2p

import (
	"github.com/ethereum/go-ethereum/metrics"
	"net"
)

const (
	ingressMeterName = "p2p/ingress"
	egressMeterName  = "p2p/egress"
)

var (
	ingressConnectMeter              = metrics.NewRegisteredMeter("p2p/serves", nil)
	ingressConnectWithHandshakeMeter = metrics.NewRegisteredMeter("p2p/serves/handshakes", nil) // Meter counting the ingress with successful handshake connections
	ingressTrafficMeter              = metrics.NewRegisteredMeter(ingressMeterName, nil)
	egressConnectMeter               = metrics.NewRegisteredMeter("p2p/dials", nil)
	egressConnectWithHandshakeMeter  = metrics.NewRegisteredMeter("p2p/dials/handshakes", nil) // Meter counting the egress with successful handshake connections
	egressTrafficMeter               = metrics.NewRegisteredMeter(egressMeterName, nil)
	activePeerGauge                  = metrics.NewRegisteredGauge("p2p/peers", nil)
	activeValidatorsPeerGauge        = metrics.NewRegisteredGauge("p2p/peers/validators", nil)   // Gauge tracking the current validators peer count
	activeProxiesPeerGauge           = metrics.NewRegisteredGauge("p2p/peers/proxies", nil)      // Gauge tracking the current proxies peer count
	discoveredPeersCounter           = metrics.NewRegisteredCounter("p2p/peers/discovered", nil) // Counter of the total discovered peers
)

// meteredConn is a wrapper around a net.Conn that meters both the
// inbound and outbound network traffic.
type meteredConn struct {
	net.Conn
}

// Read delegates a network read to the underlying connection, bumping the common
// and the peer ingress traffic meters along the way.
func (c *meteredConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	ingressTrafficMeter.Mark(int64(n))
	return n, err
}

// Write delegates a network write to the underlying connection, bumping the common
// and the peer egress traffic meters along the way.
func (c *meteredConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	egressTrafficMeter.Mark(int64(n))
	return n, err
}

// Close delegates a close operation to the underlying connection, unregisters
// the peer from the traffic registries and emits close event.
func (c *meteredConn) Close() error {
	err := c.Conn.Close()
	if err == nil {
		activePeerGauge.Dec(1)
	}
	return err
}

// newMeteredConn creates a new metered connection, bumps the ingress or egress
// connection meter and also increases the metered peer count. If the metrics
// system is disabled, function returns the original connection.
func newMeteredConn(conn net.Conn, ingress bool, addr *net.TCPAddr) net.Conn {
	// Short circuit if metrics are disabled
	if !metrics.Enabled {
		return conn
	}
	// Bump the connection counters and wrap the connection
	if ingress {
		ingressConnectMeter.Mark(1)
	} else {
		egressConnectMeter.Mark(1)
	}
	activePeerGauge.Inc(1)
	return &meteredConn{Conn: conn}
}
