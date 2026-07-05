# olcRTC transport

This fork includes an experimental native Xray transport named `olcrtc`.
It lets Xray open VLESS streams over a persistent olcRTC carrier session instead
of creating a new videoroom/WebRTC path for each proxied connection.

Current stack:

```text
VLESS
  -> Xray olcrtc transport
    -> olcRTC Manager
      -> smux streams
        -> persistent olcRTC carrier session
          -> WebRTC/SFU room
```

UDP dial path:

```text
Xray UDP dial
  -> olcrtc PacketConnWrapper
    -> olcRTC Manager datagrams
      -> persistent olcRTC carrier session
        -> WebRTC/SFU room
```

## Status

This is a custom-build feature. It is not part of upstream Xray.

Supported now:

- `network: "olcrtc"`
- `security: "none"`
- one persistent carrier session reused for many Xray connections
- `profiles[]` failover for new streams
- `maxConcurrentStreams`
- client-side `Dial()` and server-side `ListenTCP()` adapter support
- low-level Xray UDP dial support through an olcRTC `PacketConnWrapper`

Not supported yet:

- `security: "tls"` or `security: "reality"` over `olcrtc`
- stock Xray or stock mobile clients
- seamless migration of already-open TCP streams after a room/provider failure
- live-room guarantee without testing the selected SFU/provider

## Requirements

Both sides must run a build from this fork, not stock Xray.

The build currently pins `github.com/openlibrecommunity/olcrtc` to the matching
`github.com/cyber-debug/olcrtc` fork commit in `go.mod`. Update that pseudo
version whenever the olcRTC transport library changes.

Both peers must know the same room/profile details. Start the server-side Xray
first, then start the client-side Xray.

## Config fields

`streamSettings`:

```json
{
  "network": "olcrtc",
  "security": "none",
  "olcrtcSettings": {
    "auth": "jitsi",
    "roomId": "https://meet.example.org/very-random-room",
    "name": "xray-client",
    "dnsServer": "8.8.8.8:53",
    "maxConcurrentStreams": 128,
    "profiles": []
  }
}
```

Fields:

- `auth`: built-in olcRTC auth provider, for example `jitsi`, `telemost`, or `wbstream`.
- `roomId`: provider room URL or room identifier.
- `engine`: direct engine mode, used when `auth` is empty.
- `url`: direct engine URL, used when `auth` is empty.
- `token`: direct engine token, used when `auth` is empty.
- `name`: display/client name used when joining the room.
- `dnsServer`: DNS resolver used by olcRTC provider/network operations.
- `proxyAddr` / `proxyPort`: optional outbound SOCKS proxy for olcRTC provider access.
- `datagramBuffer`: olcRTC lossy datagram queue size. It does not enable Xray UDP yet.
- `maxConcurrentStreams`: max active Xray streams over one carrier. Default is `128`.
- `profiles`: optional ordered fallback profiles. New streams try the next profile if carrier setup fails.

## Server example

This example accepts VLESS over olcRTC and forwards traffic directly.

Replace:

- `SERVER_UUID`
- `https://meet.example.org/honej-random-room`

```json
{
  "log": {
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "tag": "vless-olcrtc-in",
      "listen": "127.0.0.1",
      "port": 443,
      "protocol": "vless",
      "settings": {
        "clients": [
          {
            "id": "SERVER_UUID",
            "flow": ""
          }
        ],
        "decryption": "none"
      },
      "streamSettings": {
        "network": "olcrtc",
        "security": "none",
        "olcrtcSettings": {
          "auth": "jitsi",
          "roomId": "https://meet.example.org/honej-random-room",
          "name": "honej-olcrtc-server",
          "dnsServer": "8.8.8.8:53",
          "maxConcurrentStreams": 128
        }
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "tag": "direct"
    }
  ]
}
```

Note: the `listen`/`port` values are still required by Xray config shape, but
the olcRTC listener is room-backed rather than a normal public TCP listener.

## Client example

This example exposes a local SOCKS inbound and sends traffic to the server over
the same olcRTC room.

Replace:

- `SERVER_UUID`
- `https://meet.example.org/honej-random-room`

```json
{
  "log": {
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "tag": "local-socks",
      "listen": "127.0.0.1",
      "port": 10808,
      "protocol": "socks",
      "settings": {
        "udp": true
      }
    }
  ],
  "outbounds": [
    {
      "tag": "vless-olcrtc-out",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "olcrtc.internal",
            "port": 443,
            "users": [
              {
                "id": "SERVER_UUID",
                "encryption": "none"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "olcrtc",
        "security": "none",
        "olcrtcSettings": {
          "auth": "jitsi",
          "roomId": "https://meet.example.org/honej-random-room",
          "name": "honej-olcrtc-client",
          "dnsServer": "8.8.8.8:53",
          "maxConcurrentStreams": 128
        }
      }
    }
  ]
}
```

The outbound `address` and `port` are still required by the VLESS config model.
For olcRTC, rendezvous happens through `olcrtcSettings.roomId`.

## Fallback profiles

Use `profiles` when you want new streams to try another provider/room if the
first carrier setup fails.

```json
{
  "network": "olcrtc",
  "security": "none",
  "olcrtcSettings": {
    "maxConcurrentStreams": 128,
    "profiles": [
      {
        "auth": "jitsi",
        "roomId": "https://meet-a.example.org/honej-primary",
        "name": "honej-primary",
        "dnsServer": "8.8.8.8:53"
      },
      {
        "auth": "jitsi",
        "roomId": "https://meet-b.example.org/honej-fallback",
        "name": "honej-fallback",
        "dnsServer": "1.1.1.1:53"
      }
    ]
  }
}
```

When `profiles` is set, profiles are the source of truth. Top-level `auth` and
`roomId` are used only when `profiles` is empty.

Existing TCP streams may break if the active room/provider dies. New streams can
rebuild the carrier and use the next working profile.

## Operational notes

- Use long, random room IDs.
- Keep server and client profile lists in the same order.
- Start with one `jitsi` profile and `datachannel` behavior before adding fallbacks.
- Keep `maxConcurrentStreams` conservative at first, for example `32` or `64`.
- Do not enable `tls` or `reality` yet; config validation intentionally rejects them for `olcrtc`.
- Live-test every provider/room combination before giving it to users.

## Quick validation

Build:

```bash
go build -o ./xray ./main
```

Check config parsing:

```bash
./xray run -test -config ./server.json
./xray run -test -config ./client.json
```

Run server first:

```bash
./xray run -config ./server.json
```

Run client second:

```bash
./xray run -config ./client.json
```

Then point an application at:

```text
socks5://127.0.0.1:10808
```
