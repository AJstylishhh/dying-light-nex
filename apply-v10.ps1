# apply-v10.ps1 — patches local main.go to wire NNCS + unsigned PID + trace
# Run from dying-light-nex AFTER: git pull (gets nncs.go, pid.go, trace.go)
$ErrorActionPreference = "Stop"
$main = Join-Path $PWD "main.go"
if (-not (Test-Path $main)) { Write-Host "main.go not found"; exit 1 }
$t = [System.IO.File]::ReadAllText($main)
if ($t -match "startNNCS\(\)") { Write-Host "main.go already has v10 wiring"; exit 0 }

$t = $t.Replace(
  "fmt.Println(`"[DL] build: pair relay armed v9 (LAN private IPs allowed)`")`r`n`tsettings := nex.NewSwitchSettings(accessKey, nexVersion)",
  "fmt.Println(`"[DL] build: pair relay armed v9 (LAN private IPs allowed)`")`r`n`tfmt.Println(`"[DL] build: NAT-check responder built in v10`")`r`n`tsettings := nex.NewSwitchSettings(accessKey, nexVersion)`r`n`r`n`t// Without a NAT-check answer the joiner never hole-punches (see nncs.go).`r`n`tstartNNCS()"
)
$t = $t.Replace(
  "fmt.Println(`"[DL] build: pair relay armed v9 (LAN private IPs allowed)`")`n`tsettings := nex.NewSwitchSettings(accessKey, nexVersion)",
  "fmt.Println(`"[DL] build: pair relay armed v9 (LAN private IPs allowed)`")`n`tfmt.Println(`"[DL] build: NAT-check responder built in v10`")`n`tsettings := nex.NewSwitchSettings(accessKey, nexVersion)`n`n`t// Without a NAT-check answer the joiner never hole-punches (see nncs.go).`n`tstartNNCS()"
)

$t = $t.Replace(
  "authEndpoint.Register(nex.ProtocolTicketGranting, authCfg.Handler())",
  "authEndpoint.Register(nex.ProtocolTicketGranting, traced(authCfg.Handler()))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolSecureConnection, nex.SecureConnectionHandler())",
  "secureEndpoint.Register(nex.ProtocolSecureConnection, traced(withUnsignedPIDs(nex.SecureConnectionHandler())))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolMatchmakeExtension, extensionWithBrowseURLs(mm.ExtensionHandler()))",
  "secureEndpoint.Register(nex.ProtocolMatchmakeExtension, traced(extensionWithBrowseURLs(mm.ExtensionHandler())))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolMatchMaking, mm.MatchMakingHandler())",
  "secureEndpoint.Register(nex.ProtocolMatchMaking, traced(mm.MatchMakingHandler()))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolMatchMakingExt, mm.MatchMakingExtHandler())",
  "secureEndpoint.Register(nex.ProtocolMatchMakingExt, traced(mm.MatchMakingExtHandler()))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolNATTraversal, nex.NATTraversalHandler())",
  "secureEndpoint.Register(nex.ProtocolNATTraversal, traced(nex.NATTraversalHandler()))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolRanking, nex.RankingHandler())",
  "secureEndpoint.Register(nex.ProtocolRanking, traced(nex.RankingHandler()))"
)
$t = $t.Replace(
  "secureEndpoint.Register(nex.ProtocolUtility, nex.UtilityHandler())",
  "secureEndpoint.Register(nex.ProtocolUtility, traced(nex.UtilityHandler()))"
)

$oldLog = "fmt.Printf(`"[DL %s] pid=%d proto=%#x method=%d call=%d\n`", tag, c.PID, req.Protocol, req.Method, req.CallID)"
$newLog = "fmt.Printf(`"[DL %s] pid=%d proto=%#x method=%d call=%d\n`", tag, c.PID, req.Protocol, req.Method, req.CallID)`r`n`t`ttraceRequest(tag, c, req)"
if ($t -notmatch "traceRequest") {
  $t = $t.Replace(
    "fmt.Printf(\"[DL %s] pid=%d proto=%#x method=%d call=%d\\n\", tag, c.PID, req.Protocol, req.Method, req.CallID)",
    "fmt.Printf(\"[DL %s] pid=%d proto=%#x method=%d call=%d\\n\", tag, c.PID, req.Protocol, req.Method, req.CallID)`n`t`ttraceRequest(tag, c, req)"
  )
}

if ($t -notmatch "startNNCS") {
  Write-Host "ERROR: patch did not apply (line endings?). Copy main.go from the fixed zip instead."
  exit 1
}
[System.IO.File]::WriteAllText($main, $t)
Write-Host "OK: main.go patched for v10 (NNCS + PID + trace)"
Write-Host "Open firewall: UDP 10025, 10125, 40000-40099"
Write-Host "Then: go mod tidy ; go run ."
