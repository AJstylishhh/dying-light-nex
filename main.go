// Dying Light 1 (Switch) NEX server on the nextendo-nex core.
//
// Game server ID : 27332e00  → hostname g27332e00-lp1.s.n.srv.nintendo.net
// Access key     : 0da38dbb  (extracted from the game's main NSO)
// NEX SDK        : 4.6.6
//
// Two endpoints in one process:
//   - auth   (AUTH_PORT, default 443)  TicketGranting / LoginEx
//   - secure (SECURE_PORT, default 60010) SecureConnection + matchmaking + NAT
//
// For private testing set NEXTENDO_REQUIRE_ACCOUNT=0 so guest logins work.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"time"

	nex "github.com/NextendoNetwork/nextendo-nex"
)

const (
	accessKey     = "0da38dbb" // Dying Light Switch
	nexVersion    = 40000
	securePID     = 2
	sessionKeyLen = 32
)

var (
	nextendoHost   = envOr("NEXTENDO_HOST", "127.0.0.1")
	authPort       = envOrInt("AUTH_PORT", 443)
	securePort     = envOrInt("SECURE_PORT", 60010)
	certFile       = envOr("CERT_FILE", "certs/server.crt")
	keyFile        = envOr("KEY_FILE", "certs/server.key")
	securePassword = envOr("NEXTENDO_SECURE_PASSWORD", "securepasswordplz1")
	nextendoSecret = loadNextendoSecret()
	requireAccount = os.Getenv("NEXTENDO_REQUIRE_ACCOUNT") == "1"
)

func main() {
	// Dying Light public lobbies must be marked active/open so the game enables Join.
	os.Setenv("NEXTENDO_OPEN_PARTICIPATION", "1")
	fmt.Println("[DL] build: open lobbies enabled v7")
	fmt.Println("[DL] build: host station URLs restored v6")
	settings := nex.NewSwitchSettings(accessKey, nexVersion)

	// --- Auth server ---
	secureURL := nex.NewStationURL("prudp")
	secureURL.Set("address", nextendoHost)
	secureURL.SetInt("port", securePort)
	secureURL.SetInt("CID", 1)
	secureURL.SetInt("PID", securePID)
	secureURL.SetInt("sid", 1)
	secureURL.SetInt("stream", 10)
	secureURL.SetInt("type", 2) // public

	authEndpoint := nex.NewEndpoint(settings)
	authCfg := &nex.AuthConfig{
		Settings:         settings,
		SecurePID:        securePID,
		SecurePassword:   securePassword,
		SecureStationURL: secureURL,
		ServerName:       "Nextendo-DL",
		SessionKeyLength: sessionKeyLen,
		ResolveUser:      resolveUser,
	}
	authEndpoint.Register(nex.ProtocolTicketGranting, authCfg.Handler())
	authEndpoint.OnRMC = logRMC("Auth")
	authServer := nex.NewServer(authEndpoint)

	// --- Secure server ---
	secureEndpoint := nex.NewEndpoint(settings)
	secureEndpoint.SetSecureAccount(securePassword, securePID)

	mm := nex.NewMatchmaking()
	secureEndpoint.Register(nex.ProtocolSecureConnection, nex.SecureConnectionHandler())
	secureEndpoint.Register(nex.ProtocolMatchmakeExtension, extensionWithBrowseURLs(mm.ExtensionHandler()))
	secureEndpoint.Register(nex.ProtocolMatchMaking, mm.MatchMakingHandler())
	secureEndpoint.Register(nex.ProtocolMatchMakingExt, mm.MatchMakingExtHandler())
	secureEndpoint.Register(nex.ProtocolNATTraversal, nex.NATTraversalHandler())
	secureEndpoint.Register(nex.ProtocolRanking, nex.RankingHandler())
	secureEndpoint.Register(nex.ProtocolUtility, nex.UtilityHandler())
	secureEndpoint.OnRMC = logRMC("Secure")
	secureEndpoint.OnConnect = func(c *nex.Connection) {
		fmt.Printf("[DL Secure] connected pid=%d id=%d addr=%s\n", c.PID, c.ID, c.RemoteAddr)
	}
	secureEndpoint.OnDisconnect = func(c *nex.Connection) {
		mm.RemovePlayer(c.PID)
	}
	secureServer := nex.NewServer(secureEndpoint)
	secureEndpoint.StartReaper()

	proxyProto := os.Getenv("NEXTENDO_PROXY_PROTOCOL") == "1"
	go func() {
		fmt.Printf("[DL Auth] listening :%d (proxyProto=%v, secure → %s)\n", authPort, proxyProto, secureURL.String())
		var err error
		if proxyProto {
			err = authServer.ListenSecureProxy(authPort, certFile, keyFile)
		} else {
			err = authServer.ListenSecure(authPort, certFile, keyFile)
		}
		if err != nil {
			fmt.Printf("[DL Auth] stopped: %v\n", err)
		}
	}()

	fmt.Printf("[DL Secure] listening :%d  (accessKey=%s gameServer=27332e00)\n", securePort, accessKey)
	if err := secureServer.ListenSecure(securePort, certFile, keyFile); err != nil {
		fmt.Printf("[DL Secure] stopped: %v\n", err)
	}
}

func resolveUser(username string, extraData []byte) (uint64, []byte, bool) {
	sk := sha256.Sum256([]byte("nextendo-src:" + username))
	sourceKey := sk[:]

	// 1. Signed nx2 token from Nextendo account service
	if pid, ok := nextendoPIDFromToken(username); ok {
		fmt.Printf("[Auth] nx2 token → pid=%d\n", pid)
		return pid, sourceKey, true
	}

	// 2. Numeric username (emulator PID or Switch NSA id range)
	if n, err := strconv.ParseUint(username, 10, 64); err == nil && n >= 1800000000 {
		fmt.Printf("[Auth] numeric identity pid=%d\n", n)
		if requireAccount {
			fmt.Printf("[Auth] REFUSED: NEXTENDO_REQUIRE_ACCOUNT=1 and no valid nx2 token\n")
			return 0, nil, false
		}
		return n, sourceKey, true
	}

	// 3. Anonymous / guest
	if requireAccount {
		fmt.Printf("[Auth] anonymous REFUSED (account required): %q\n", username)
		return 0, nil, false
	}
	pid := anonymousPID(username)
	fmt.Printf("[Auth] guest %q → pid=%d\n", username, pid)
	return pid, sourceKey, true
}

func nextendoPIDFromToken(s string) (uint64, bool) {
	if len(nextendoSecret) == 0 || !strings.HasPrefix(s, "nx2.") {
		return 0, false
	}
	parts := strings.Split(s[len("nx2."):], ".")
	if len(parts) != 2 {
		return 0, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, false
	}
	mac := hmac.New(sha256.New, nextendoSecret)
	mac.Write([]byte("nex:" + string(raw)))
	want := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(parts[1])) {
		return 0, false
	}
	f := strings.SplitN(string(raw), ".", 3)
	if len(f) != 3 {
		return 0, false
	}
	pid, err := strconv.ParseUint(f[0], 10, 64)
	if err != nil {
		return 0, false
	}
	if exp, err := strconv.ParseInt(f[2], 10, 64); err != nil || time.Now().Unix() > exp {
		return 0, false
	}
	return pid, true
}

func loadNextendoSecret() []byte {
	if v := os.Getenv("NEXTENDO_SECRET"); v != "" {
		return []byte(v)
	}
	path := envOr("NEXTENDO_SECRET_FILE", "nextendo_secret.key")
	if b, err := os.ReadFile(path); err == nil {
		if dec, derr := hex.DecodeString(strings.TrimSpace(string(b))); derr == nil && len(dec) >= 16 {
			return dec
		}
	}
	return nil
}

func anonymousPID(username string) uint64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(username))
	return 1800000000 + uint64(h.Sum32()%100000000)
}

// Method 53 = BrowseMatchmakeSessionWithHostUrlsNoHolderNoResultRange.
// Same request as method 52 (0x34, handled by the library). Response = List<MatchmakeSession>
// followed by List<GatheringURLs> {u32 gid; List<StationURL>}.
const methodBrowseWithHostURLs uint32 = 0x35

// gatheringURLs is the NEX structure GatheringURLs {u32 m_gid; List<StationURL> m_lstStationURLs}.
// It must be framed like any other structure (out.Add does that).
type gatheringURLs struct {
	gid  uint32
	urls []*nex.StationURL
}

func (g *gatheringURLs) Levels() []nex.Level {
	return []nex.Level{{
		Save: func(o *nex.StreamOut) {
			o.U32(g.gid)
			nex.WriteList(o, g.urls, func(o *nex.StreamOut, u *nex.StationURL) { o.StationURL(u) })
		},
	}}
}

func extensionWithBrowseURLs(base nex.RMCHandler) nex.RMCHandler {
	return func(conn *nex.Connection, req *nex.RMCMessage) *nex.RMCMessage {
		if req.Method != methodBrowseWithHostURLs {
			return base(conn, req)
		}

		as52 := *req
		as52.Method = 0x34
		resp := base(conn, &as52)
		if resp == nil || resp.IsError {
			return resp
		}

		type entry struct {
			ss   nex.MatchmakeSession
			urls []*nex.StationURL
		}
		var kept []*entry

		in := nex.NewStreamIn(resp.Body, conn.Settings)
		count := in.U32()
		for i := uint32(0); i < count && in.Err() == nil; i++ {
			e := &entry{}
			in.Extract(&e.ss)
			if in.Err() != nil {
				break
			}
			fmt.Printf("[DL] session candidate gid=%d owner=%d host=%d state=%d open=%v policy=%d flags=%#x min=%d max=%d players=%d mode=%d system=%d appdata=%d key=%d userpw=%v syspw=%v desc=%q\\n", e.ss.ID, e.ss.OwnerPID, e.ss.HostPID, e.ss.State, e.ss.OpenParticipation, e.ss.ParticipationPolicy, e.ss.Flags, e.ss.MinParticipants, e.ss.MaxParticipants, e.ss.NumParticipants, e.ss.GameMode, e.ss.MatchmakeSystem, len(e.ss.ApplicationData), len(e.ss.SessionKey), e.ss.UserPasswordEnabled, e.ss.SystemPasswordEnabled, e.ss.Description)
			if e.ss.HostPID == conn.PID || e.ss.OwnerPID == conn.PID {
				fmt.Printf("[DL] browse(53) skip own lobby gid=%d\n", e.ss.ID)
				continue
			}
			host := conn.Endpoint.FindConnectionByPID(e.ss.HostPID)
			if host == nil {
				fmt.Printf("[DL] browse(53) skip gid=%d (host pid=%d not connected)\n", e.ss.ID, e.ss.HostPID)
				continue
			}
			e.urls = host.Stations()
			kept = append(kept, e)
		}
		if in.Err() != nil {
			fmt.Printf("[DL] browse(53) could not re-read sessions (%v) -> returning no lobbies\n", in.Err())
			kept = nil
		}

		out := nex.NewStreamOut(conn.Settings)
		out.U32(uint32(len(kept)))
		for _, e := range kept {
			out.Add(&e.ss)
		}
		out.U32(uint32(len(kept)))
		for _, e := range kept {
			out.Add(&gatheringURLs{gid: e.ss.ID, urls: e.urls})
			fmt.Printf("[DL] browse(53) gid=%d host=%d urls=%d\n", e.ss.ID, e.ss.HostPID, len(e.urls))
			for _, u := range e.urls {
				fmt.Printf("[DL]    %s\n", u.String())
			}
		}
		fmt.Printf("[DL] browse(53) pid=%d -> %d lobby(ies)\n", conn.PID, len(kept))

		return nex.NewRMCSuccess(conn.Settings, nex.ProtocolMatchmakeExtension, methodBrowseWithHostURLs, req.CallID, out.Bytes())
	}
}

func logRMC(tag string) func(*nex.Connection, *nex.RMCMessage) {
	return func(c *nex.Connection, req *nex.RMCMessage) {
		fmt.Printf("[DL %s] pid=%d proto=%#x method=%d call=%d\n", tag, c.PID, req.Protocol, req.Method, req.CallID)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
