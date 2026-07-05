package olcrtc

import (
	"context"
	"sync"
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

	mu              sync.Mutex
	readDeadline    time.Time
	writeDeadline   time.Time
	deadlineChanged chan struct{}
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

			deadlineChanged: make(chan struct{}),
		},
		Dest: addr,
	}
}

func (c *packetConn) ReadFrom(p []byte) (int, net.Addr, error) {
	ctx, cancel := c.operationContext(true)
	defer cancel()
	payload, err := c.manager.ReceiveDatagram(ctx)
	if err != nil {
		return 0, nil, err
	}
	n := copy(p, payload)
	return n, c.remote, nil
}

func (c *packetConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	ctx, cancel := c.operationContext(false)
	defer cancel()
	if err := c.manager.SendDatagram(ctx, p); err != nil {
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

func (c *packetConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.writeDeadline = t
	c.broadcastDeadlineChangeLocked()
	c.mu.Unlock()
	return nil
}

func (c *packetConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	c.readDeadline = t
	c.broadcastDeadlineChangeLocked()
	c.mu.Unlock()
	return nil
}

func (c *packetConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	c.writeDeadline = t
	c.broadcastDeadlineChangeLocked()
	c.mu.Unlock()
	return nil
}

func (c *packetConn) operationContext(read bool) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.ctx)
	go c.watchDeadline(ctx, cancel, read)
	return ctx, cancel
}

func (c *packetConn) watchDeadline(ctx context.Context, cancel context.CancelFunc, read bool) {
	for {
		deadline, changed := c.deadline(read)
		if deadline.IsZero() {
			select {
			case <-ctx.Done():
				return
			case <-changed:
				continue
			}
		}
		wait := time.Until(deadline)
		if wait <= 0 {
			cancel()
			return
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-changed:
			timer.Stop()
			continue
		case <-timer.C:
			cancel()
			return
		}
	}
}

func (c *packetConn) deadline(read bool) (time.Time, <-chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if read {
		return c.readDeadline, c.deadlineChanged
	}
	return c.writeDeadline, c.deadlineChanged
}

func (c *packetConn) broadcastDeadlineChangeLocked() {
	close(c.deadlineChanged)
	c.deadlineChanged = make(chan struct{})
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
