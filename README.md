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

## What works today (v10)

- Auth (TicketGranting / LoginEx) with the correct access key
- Secure connection + common matchmaking / NAT protocols
- Guest logins when `NEXTENDO_REQUIRE_ACCOUNT=0`
- Create + browse public lobbies (Switch and Citron)
- **Built-in NNCS NAT-check responder** (UDP 10025/10125) — required for Join
- **Unsigned PID fix** for real Switch account IDs in station URLs
- Pair relay armed for failed hole-punch fallback

## Why Join was broken before v10

Dying Light uses NEX’s own P2P stack. Before joining, the client must learn its public UDP endpoint from the **NAT check** (`nncs1` / `nncs2` on ports **10025** and **10125**). Without a reply it never hole-punches and shows “game no longer available” / gray Join. The v10 server answers those probes itself.

---

## Quick start

### 1. Get the code

```powershell
cd path\to\dying-light-nex
git pull
# First time after v10: wire main.go
.\apply-v10.ps1
```

Or copy `main.go`, `nncs.go`, `pid.go`, `trace.go` from the fixed zip.

### 2. Firewall (Administrator PowerShell)

```powershell
New-NetFirewallRule -DisplayName "DyingLight NEX Auth" -Direction Inbound -Protocol TCP -LocalPort 443 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Secure" -Direction Inbound -Protocol TCP -LocalPort 60010 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Relay" -Direction Inbound -Protocol UDP -LocalPort 40000-40099 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NNCS" -Direction Inbound -Protocol UDP -LocalPort 10025,10125 -Action Allow
```

### 3. Run

```powershell
$env:NEXTENDO_HOST = "192.168.1.106"   # your PC LAN IP
$env:AUTH_PORT = "443"
$env:SECURE_PORT = "60010"
$env:CERT_FILE = "certs/server.crt"
$env:KEY_FILE = "certs/server.key"
$env:NEXTENDO_REQUIRE_ACCOUNT = "0"
go mod tidy
go run .
```

You should see:

```
[DL] build: NAT-check responder built in v10
[NNCS] NAT-check responder on UDP 0.0.0.0:10025
[NNCS] NAT-check responder on UDP 0.0.0.0:10125
[DL] pair relay armed on ...
[DL Auth] listening :443 ...
[DL Secure] listening :60010 ...
```

### 4. Point the Switch at your server

Atmosphere hosts (`emummc.txt` / `sysmmc.txt`):

```
192.168.1.106  g27332e00-lp1.s.n.srv.nintendo.net
```

Fully power off and on after editing hosts.

For emulators (Citron), set the server IP and NAT IP so both nncs1 and nncs2 reach this machine.

---

## Important notes

- **Do not** share your game dump, keys, or copyrighted files.
- Self-signed certs are fine for private testing.
- This server ships **no** Nintendo code. You must own a legitimate copy of the game.

## License

Adapted from the public Nextendo game-server pattern.  
For personal / private use while developing Dying Light support.
