# dying-light-nex

Nextendo-compatible **NEX server for Dying Light 1 (Nintendo Switch)**.

| Constant | Value | Source |
|----------|-------|--------|
| Game server ID | `27332e00` | Wireshark capture of Public matchmaking |
| Access key | `0da38dbb` | Extracted from the game's main NSO |
| Hostname | `g27332e00-lp1.s.n.srv.nintendo.net` | NEX convention + capture |
| NEX SDK | 4.6.6 | String in the binary |

Built on [nextendo-nex](https://github.com/NextendoNetwork/nextendo-nex).  
This is a **private / self-hosted** starting point. It is **not** an official Nextendo Network game server yet.

---

## What works today

- Auth (TicketGranting / LoginEx) with the correct access key
- Secure connection + common matchmaking / NAT protocols
- Guest logins when `NEXTENDO_REQUIRE_ACCOUNT=0`
- Create + browse public lobbies (Switch and Citron)
- **Pair relay** armed by default: after 2 failed hole-punches the server can relay traffic so Join is not stuck on `relaye=false`

## What still needs work

- Station URL quality (`port=1` on local stations still appears in some Register paths)
- Presence / `UpdateNotificationData` if the game expects it for Join UI
- Full polish for public matchmaking across the internet

Progressive testing order:

1. Two clients can create a session with each other
2. Private / friend co-op
3. Public matchmaking

---

## Quick start (local test)

### 1. Generate a self-signed certificate

```bash
mkdir -p certs
openssl req -x509 -newkey rsa:2048 -keyout certs/server.key -out certs/server.crt \
  -days 365 -nodes -subj "/CN=g27332e00-lp1.s.n.srv.nintendo.net"
```

### 2. Open firewall ports (Windows, Administrator PowerShell)

```powershell
New-NetFirewallRule -DisplayName "DyingLight NEX Auth" -Direction Inbound -Protocol TCP -LocalPort 443 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Secure" -Direction Inbound -Protocol TCP -LocalPort 60010 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Relay" -Direction Inbound -Protocol UDP -LocalPort 40000-40099 -Action Allow
```

### 3. Run (set env in the same PowerShell window)

```powershell
cd path\to\dying-light-nex
$env:NEXTENDO_HOST = "192.168.1.106"   # your PC LAN IP
$env:AUTH_PORT = "443"
$env:SECURE_PORT = "60010"
$env:CERT_FILE = "certs/server.crt"
$env:KEY_FILE = "certs/server.key"
$env:NEXTENDO_REQUIRE_ACCOUNT = "0"
go run .
```

You should see:

```
[DL] build: open lobbies enabled v8
[DL] build: pair relay armed v8
[DL] pair relay armed on 192.168.1.106 UDP 40000-40099 (open to all who need it)
[DL Auth] listening :443 ...
[DL Secure] listening :60010  (accessKey=0da38dbb gameServer=27332e00)
```

### 4. Point the Switch at your server

Add to Atmosphere hosts (`emummc.txt` / `sysmmc.txt`):

```
192.168.1.106  g27332e00-lp1.s.n.srv.nintendo.net
```

Fully power off and on the Switch after editing hosts.

### 5. Test join

1. Start this server, leave it running.
2. Host creates a lobby on Citron or Switch.
3. Other client opens Find Games.
4. Watch for `[relais-besoin]` then later `relaye=true` (or successful join) instead of only `hole-punch FAILED ... relaye=false`.

---

## Important notes

- **Do not** share your game dump, keys, or copyrighted files.
- Self-signed certs are fine for private testing.
- Port 443 often needs admin on Windows.
- This server ships **no** Nintendo code. You must own a legitimate copy of the game.

## License

Adapted from the public Nextendo game-server pattern.  
For personal / private use while developing Dying Light support.
