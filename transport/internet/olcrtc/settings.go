package olcrtc

import (
	"strconv"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
)

func managerConfig(cfg *Config, server bool) olclib.ManagerConfig {
	return olclib.ManagerConfig{
		Server: server,
		Session: olclib.Config{
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
		},
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
