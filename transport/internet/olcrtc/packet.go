package olcrtc

import (
	"context"
	"time"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/stat"
)

type packetConn struct {
	ctx     context.Context
	cancel  context.CancelFunc
	manager *olclib.Manager
	local   net.Addr
	remote  net.Addr
}

func newPacketConnection(ctx context.Context, manager *olclib.Manager, dest net.Destination) stat.Connection {
	runCtx, cancel := context.WithCancel(ctx)
	addr := packetAddr{dest: dest}
	return &internet.PacketConnWrapper{
		PacketConn: &packetConn{
			ctx:     runCtx,
			cancel:  cancel,
			manager: manager,
			local:   olcrtcAddr("packet-local"),
			remote:  addr,
		},
		Dest: addr,
	}
}

func (c *packetConn) ReadFrom(p []byte) (int, net.Addr, error) {
	payload, err := c.manager.ReceiveDatagram(c.ctx)
	if err != nil {
		return 0, nil, err
	}
	n := copy(p, payload)
	return n, c.remote, nil
}

func (c *packetConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	if err := c.manager.SendDatagram(c.ctx, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *packetConn) Close() error {
	c.cancel()
	return nil
}

func (c *packetConn) LocalAddr() net.Addr {
	return c.local
}

func (c *packetConn) SetDeadline(time.Time) error {
	return nil
}

func (c *packetConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *packetConn) SetWriteDeadline(time.Time) error {
	return nil
}

type packetAddr struct {
	dest net.Destination
}

func (a packetAddr) Network() string {
	return "udp"
}

func (a packetAddr) String() string {
	return a.dest.NetAddr()
}
