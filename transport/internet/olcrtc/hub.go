package olcrtc

import (
	"context"
	"time"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
)

type Listener struct {
	cancel  context.CancelFunc
	done    chan struct{}
	addr    net.Addr
	manager *olclib.Manager
}

func ListenTCP(
	ctx context.Context,
	_ net.Address,
	_ net.Port,
	streamSettings *internet.MemoryStreamConfig,
	handler internet.ConnHandler,
) (internet.Listener, error) {
	olclib.RegisterDefaults()
	if streamSettings == nil {
		return nil, errors.New("missing olcrtc stream settings").AtError()
	}
	settings, ok := streamSettings.ProtocolSettings.(*Config)
	if !ok {
		return nil, errors.New("invalid olcrtc settings").AtError()
	}
	runCtx, cancel := context.WithCancel(ctx)
	manager := olclib.NewManager(managerConfig(settings, true))
	l := &Listener{
		cancel:  cancel,
		done:    make(chan struct{}),
		addr:    olcrtcAddr("listener"),
		manager: manager,
	}
	go l.acceptLoop(runCtx, manager, handler)
	return l, nil
}

func (l *Listener) acceptLoop(ctx context.Context, manager *olclib.Manager, handler internet.ConnHandler) {
	defer close(l.done)
	for ctx.Err() == nil {
		conn, err := manager.AcceptStream(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			errors.LogWarningInner(ctx, err, "failed to accept olcrtc stream")
			time.Sleep(500 * time.Millisecond)
			continue
		}
		go handler(&addrConn{Conn: conn})
	}
}

func (l *Listener) Addr() net.Addr {
	return l.addr
}

func (l *Listener) Close() error {
	l.cancel()
	err := l.manager.Close()
	<-l.done
	return err
}

type olcrtcAddr string

func (a olcrtcAddr) Network() string { return "olcrtc" }
func (a olcrtcAddr) String() string  { return string(a) }

type addrConn struct {
	net.Conn
}

func (c *addrConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 24444}
}

func (c *addrConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, ListenTCP))
}
