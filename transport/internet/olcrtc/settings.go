package olcrtc

import (
	"strconv"

	olclib "github.com/openlibrecommunity/olcrtc/pkg/olcrtc"
)

func managerConfig(cfg *Config, server bool) olclib.ManagerConfig {
	return olclib.ManagerConfig{
		Server:               server,
		Session:              sessionConfig(cfg),
		Profiles:             profileConfigs(cfg.GetProfiles()),
		MaxConcurrentStreams: int(cfg.GetMaxConcurrentStreams()),
	}
}

func configKey(cfg *Config) string {
	key := profileKey(sessionConfig(cfg)) + "\x00" +
		strconv.FormatUint(uint64(cfg.GetMaxConcurrentStreams()), 10)
	for _, profile := range cfg.GetProfiles() {
		key += "\x00" + profileKey(profileConfig(profile))
	}
	return key
}

func sessionConfig(cfg *Config) olclib.Config {
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

func profileConfigs(profiles []*Profile) []olclib.ProfileConfig {
	if len(profiles) == 0 {
		return nil
	}
	out := make([]olclib.ProfileConfig, 0, len(profiles))
	for _, profile := range profiles {
		out = append(out, olclib.ProfileConfig(profileConfig(profile)))
	}
	return out
}

func profileConfig(profile *Profile) olclib.Config {
	return olclib.Config{
		Auth:           profile.GetAuth(),
		RoomID:         profile.GetRoomId(),
		Engine:         profile.GetEngine(),
		URL:            profile.GetUrl(),
		Token:          profile.GetToken(),
		Name:           profile.GetName(),
		DNSServer:      profile.GetDnsServer(),
		ProxyAddr:      profile.GetProxyAddr(),
		ProxyPort:      int(profile.GetProxyPort()),
		DatagramBuffer: int(profile.GetDatagramBuffer()),
	}
}

func profileKey(cfg olclib.Config) string {
	return cfg.Auth + "\x00" +
		cfg.RoomID + "\x00" +
		cfg.Engine + "\x00" +
		cfg.URL + "\x00" +
		cfg.Token + "\x00" +
		cfg.Name + "\x00" +
		cfg.DNSServer + "\x00" +
		cfg.ProxyAddr + "\x00" +
		strconv.Itoa(cfg.ProxyPort) + "\x00" +
		strconv.Itoa(cfg.DatagramBuffer)
}
