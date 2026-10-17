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

## What still needs work

- Game-specific matchmaking methods Dying Light may call beyond the common ones
- Full integration with Nextendo account / presence (optional)
- Confirmed P2P session between two real Switches
- Public matchmaking polish

Progressive testing order (from the project notes):

1. Two consoles can create a session with each other
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

### 2. Configure

```bash
cp example.env .env
# Edit NEXTENDO_HOST to your PC's LAN IP that the Switch can reach
# (e.g. the ICS/hotspot gateway 192.168.137.1)
```

### 3. Run

```bash
go mod tidy
go run .
```

You should see:

```
[DL Auth] listening :443 ...
[DL Secure] listening :60010  (accessKey=0da38dbb gameServer=27332e00)
```

### 4. Point the Switch at your server

You need the hostname `g27332e00-lp1.s.n.srv.nintendo.net` to resolve to your PC.

**Option A – hosts file on the PC (if Switch uses your PC as DNS / ICS)**  
Add to the Windows hosts file (or the DNS the Switch is using):

```
192.168.137.1  g27332e00-lp1.s.n.srv.nintendo.net
```

**Option B – Prelude / Atmosphere DNS MITM**  
Add an explicit entry for the same hostname in your Nextendo/Prelude hosts list so it overrides Nintendo.

**Option C – local DNS server**  
Run a small DNS that answers only that name with your IP and leave everything else alone.

### 5. Test

1. Close Dying Light completely.
2. Start this server.
3. Launch Dying Light → Online → Public (or Private with a friend).
4. Watch the server console for `[DL Auth]` / `[DL Secure]` lines.
5. Capture with Wireshark again if the error changes — the new packets will tell us which RMC method is missing.

---

## Important notes

- **Do not** share your game dump, keys, or copyrighted files.
- Self-signed certs are fine for private testing; browsers and some tools will complain, the game usually does not care about the CA.
- Port 443 often needs admin/root on Windows/Linux. Use a high port + port-forward, or run elevated.
- This server ships **no** Nintendo code. You must own a legitimate copy of the game.

## License

Adapted from the public Nextendo game-server pattern.  
For personal / private use while developing Dying Light support.
