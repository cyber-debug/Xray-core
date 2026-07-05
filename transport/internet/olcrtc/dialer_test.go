package olcrtc

import (
	"context"
	"testing"

	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
)

func TestDialUDPReturnsPacketConnWrapper(t *testing.T) {
	stream := &internet.MemoryStreamConfig{
		ProtocolName:     protocolName,
		ProtocolSettings: &Config{},
	}
	dest := net.UDPDestination(net.DomainAddress("server.example"), 443)

	conn, err := Dial(context.Background(), dest, stream)
	if err != nil {
		t.Fatalf("Dial(udp) error = %v", err)
	}
	defer conn.Close()

	packetConn, ok := conn.(*internet.PacketConnWrapper)
	if !ok {
		t.Fatalf("Dial(udp) returned %T, want *internet.PacketConnWrapper", conn)
	}
	if got := packetConn.RemoteAddr().String(); got != "server.example:443" {
		t.Fatalf("RemoteAddr() = %q, want server.example:443", got)
	}
}

func TestInternetDialRoutesOLCRTCUDP(t *testing.T) {
	stream := &internet.MemoryStreamConfig{
		ProtocolName:     protocolName,
		ProtocolSettings: &Config{},
	}
	dest := net.UDPDestination(net.DomainAddress("server.example"), 443)

	conn, err := internet.Dial(context.Background(), dest, stream)
	if err != nil {
		t.Fatalf("internet.Dial(olcrtc udp) error = %v", err)
	}
	defer conn.Close()

	if _, ok := conn.(*internet.PacketConnWrapper); !ok {
		t.Fatalf("internet.Dial(olcrtc udp) returned %T, want *internet.PacketConnWrapper", conn)
	}
}
