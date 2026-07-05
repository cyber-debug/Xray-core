package olcrtc

import (
	"context"
	"net"
	"strconv"
	"sync"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
	"github.com/xtaci/smux"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/transport/internet/stat"
)

type streamOpener interface {
	OpenStream() (*smux.Stream, error)
}

type streamAccepter interface {
	AcceptStream() (*smux.Stream, error)
}

type sessionManager struct {
	mu      sync.Mutex
	cfg     *Config
	session *olclib.Session
	raw     net.Conn
	mux     streamOpener
}

func (m *sessionManager) openStream(ctx context.Context, server bool) (stat.Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mux == nil {
		if err := m.connectLocked(ctx, server); err != nil {
			return nil, err
		}
	}
	stream, err := m.mux.OpenStream()
	if err != nil {
		m.closeLocked()
		return nil, errors.New("open olcrtc smux stream").Base(err)
	}
	return stat.Connection(stream), nil
}

func (m *sessionManager) acceptStream(ctx context.Context) (stat.Connection, error) {
	m.mu.Lock()
	if m.mux == nil {
		if err := m.connectLocked(ctx, true); err != nil {
			m.mu.Unlock()
			return nil, err
		}
	}
	accepter, ok := m.mux.(streamAccepter)
	m.mu.Unlock()
	if !ok {
		return nil, errors.New("olcrtc smux session cannot accept streams")
	}
	stream, err := accepter.AcceptStream()
	if err != nil {
		m.mu.Lock()
		m.closeLocked()
		m.mu.Unlock()
		return nil, errors.New("accept olcrtc smux stream").Base(err)
	}
	return stat.Connection(stream), nil
}

func (m *sessionManager) connectLocked(ctx context.Context, server bool) error {
	sess, err := olclib.New(ctx, toOlcRTCConfig(m.cfg))
	if err != nil {
		return errors.New("create olcrtc session").Base(err)
	}
	var raw net.Conn
	if server {
		raw, err = sess.AcceptStream(ctx)
	} else {
		raw, err = sess.OpenStream(ctx)
	}
	if err != nil {
		_ = sess.Close()
		return errors.New("open olcrtc carrier stream").Base(err)
	}
	var mux *smux.Session
	if server {
		mux, err = smux.Server(raw, smux.DefaultConfig())
	} else {
		mux, err = smux.Client(raw, smux.DefaultConfig())
	}
	if err != nil {
		_ = raw.Close()
		_ = sess.Close()
		return errors.New("open olcrtc smux session").Base(err)
	}
	m.session = sess
	m.raw = raw
	m.mux = mux
	return nil
}

func (m *sessionManager) close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeLocked()
}

func (m *sessionManager) closeLocked() error {
	var err error
	if closer, ok := m.mux.(interface{ Close() error }); ok {
		err = closer.Close()
	}
	m.mux = nil
	if m.raw != nil {
		if closeErr := m.raw.Close(); err == nil {
			err = closeErr
		}
		m.raw = nil
	}
	if m.session != nil {
		if closeErr := m.session.Close(); err == nil {
			err = closeErr
		}
		m.session = nil
	}
	return err
}

func toOlcRTCConfig(cfg *Config) olclib.Config {
	return olclib.Config{
		Auth:           cfg.GetAuth(),
		RoomID:         cfg.GetRoomId(),
		Engine:         cfg.GetEngine(),
		URL:            cfg.GetUrl(),
		Token:          cfg.GetToken(),
		Name:           cfg.GetName(),
		DNSServer:      cfg.GetDnsServer(),
		ProxyAddr:      cfg.GetProxyAddr(),
		ProxyPort:      int(cfg.GetProxyPort()),
		DatagramBuffer: int(cfg.GetDatagramBuffer()),
	}
}

func configKey(cfg *Config) string {
	return cfg.GetAuth() + "\x00" +
		cfg.GetRoomId() + "\x00" +
		cfg.GetEngine() + "\x00" +
		cfg.GetUrl() + "\x00" +
		cfg.GetToken() + "\x00" +
		cfg.GetName() + "\x00" +
		cfg.GetDnsServer() + "\x00" +
		cfg.GetProxyAddr() + "\x00" +
		strconv.FormatUint(uint64(cfg.GetProxyPort()), 10) + "\x00" +
		strconv.FormatUint(uint64(cfg.GetDatagramBuffer()), 10)
}
