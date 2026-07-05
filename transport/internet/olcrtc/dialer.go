package olcrtc

import (
	"context"
	"sync"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/stat"
)

var dialManagers sync.Map

func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (stat.Connection, error) {
	olclib.RegisterDefaults()
	if streamSettings == nil {
		return nil, errors.New("missing olcrtc stream settings").AtError()
	}
	settings, ok := streamSettings.ProtocolSettings.(*Config)
	if !ok {
		return nil, errors.New("invalid olcrtc settings").AtError()
	}
	key := configKey(settings)
	raw, _ := dialManagers.LoadOrStore(key, olclib.NewManager(managerConfig(settings, false)))
	manager := raw.(*olclib.Manager)
	if dest.Network == net.Network_UDP {
		return newPacketConnection(ctx, manager, dest), nil
	}
	conn, err := manager.OpenStream(ctx)
	if err != nil {
		return nil, errors.New("open olcrtc stream").Base(err)
	}
	return stat.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
