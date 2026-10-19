# dying-light-nex

Nextendo-compatible **NEX server for Dying Light 1 (Nintendo Switch)**.

| Constant | Value | Source |
|----------|-------|--------|
| Game server ID | `27332e00` | Wireshark capture of Public matchmaking |
| Access key | `0da38dbb` | Extracted from the game's main NSO |
| Hostname | `g27332e00-lp1.s.n.srv.nintendo.net` | NEX convention + capture |
| NEX SDK | 4.6.6 | String in the binary |

Built on [nextendo-nex](https://github.com/NextendoNetwork/nextendo-nex).
This is a **private / self-hosted** server. It is **not** an official Nextendo Network game server yet.

---

## Status

**Join works.** Tested with two citron-nextendo instances on one PC (game 1.0.5, Public game
type): Find Games lists the host with a ping, Join loads the second player into the host's world,
and both players see each other.

## Why Join used to fail ("The game you have tried to join is no longer available")

Dying Light uses NEX's own P2P stack (not Pia). Before joining, the client must learn its public
UDP endpoint from the **NAT check**: 16-byte probes to `nncs1` / `nncs2` on UDP **10025** and
**10125**. Without a reply, the client never hole-punches and the game shows the error above.

This server includes a built-in NNCS responder (`nncs.go`) that answers those probes.

## What this server does

- Auth (TicketGranting / LoginEx) with access key `0da38dbb`
- Secure connection + matchmaking + NAT protocols
- Guest logins when `NEXTENDO_REQUIRE_ACCOUNT=0`
- Built-in NNCS NAT-check on UDP 10025/10125
- Unsigned PID fix for real Switch account IDs in station URLs
- Pair relay for failed hole-punch fallback

---

## Quick start

### 1. Get the code

```powershell
git clone https://github.com/AJstylishhh/dying-light-nex.git
cd dying-light-nex
# If main.go is missing startNNCS, run:
.\apply-v10.ps1
```

Or use the fixed zip (includes a ready `main.go`).

### 2. Certificates

```powershell
mkdir certs -Force
# Use OpenSSL or the existing self-signed certs for g27332e00-lp1.s.n.srv.nintendo.net
```

### 3. Firewall (Administrator PowerShell)

```powershell
New-NetFirewallRule -DisplayName "DyingLight NEX Auth" -Direction Inbound -Protocol TCP -LocalPort 443 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Secure" -Direction Inbound -Protocol TCP -LocalPort 60010 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NEX Relay" -Direction Inbound -Protocol UDP -LocalPort 40000-40099 -Action Allow
New-NetFirewallRule -DisplayName "DyingLight NNCS" -Direction Inbound -Protocol UDP -LocalPort 10025,10125 -Action Allow
```

### 4. Run

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
[DL Auth] listening :443 ...
[DL Secure] listening :60010 ...
```

### 5. Point clients at this server

**Switch (Atmosphere hosts):**

```
192.168.1.106  g27332e00-lp1.s.n.srv.nintendo.net
```

**Citron / emulator:** set server IP and NAT IP so both nncs1 and nncs2 reach this machine.

---

## Environment variables

| Variable | Default | Meaning |
|----------|---------|--------|
| `NEXTENDO_HOST` | `127.0.0.1` | IP clients use for secure server |
| `AUTH_PORT` | `443` | Auth TLS port |
| `SECURE_PORT` | `60010` | Secure NEX port |
| `NNCS_ENABLE` | on | Set `0` to disable built-in NAT check |
| `NNCS_BIND` | `0.0.0.0` | Bind address for NNCS |
| `DL_TRACE` | `0` | `1` = log every RMC body |

See `example.env` for the full list.

---

## Important notes

- **Do not** share your game dump, keys, or copyrighted files.
- Self-signed certs are fine; emulators accept any certificate.
- This server ships **no** Nintendo code. You must own a legitimate copy of the game.

## License

Adapted from the public Nextendo game-server pattern.
For personal / private use while developing Dying Light support.
