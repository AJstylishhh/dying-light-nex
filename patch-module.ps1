# Patches nextendo-nex in the local Go module cache so pair-relay works on LAN.
# Root cause: PublicIPOf() returns "" for private IPs (192.168.x.x), so PairRelayFor never arms.
# Run once from the dying-light-nex folder, then go run . again.

$ErrorActionPreference = "Stop"
Write-Host "Looking for nextendo-nex in module cache..."

$gopath = (go env GOPATH).Trim()
$modRoot = Join-Path $gopath "pkg\mod\github.com\!nextendo!network"
if (-not (Test-Path $modRoot)) {
  Write-Host "Module not downloaded yet. Running go mod download..."
  go mod download github.com/NextendoNetwork/nextendo-nex
}

$dirs = Get-ChildItem -Path $modRoot -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like "nextendo-nex@*" }
if (-not $dirs) {
  Write-Host "ERROR: nextendo-nex not found under $modRoot"
  Write-Host "Run: go mod tidy   then run this script again."
  exit 1
}

foreach ($d in $dirs) {
  $file = Join-Path $d.FullName "relaypair.go"
  if (-not (Test-Path $file)) { continue }
  $text = Get-Content -Raw $file
  if ($text -match "NEXTENDO_RELAY_ALLOW_PRIVATE") {
    Write-Host "Already patched: $($d.Name)"
    continue
  }
  # Make module files writable
  attrib -r $file
  $old = @'
func PublicIPOf(c *Connection) string {
	if c == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(c.RemoteAddr)
	if err != nil || host == "" || isPrivateIP(host) {
		return ""
	}

	return host
}
'@
  $new = @'
func PublicIPOf(c *Connection) string {
	if c == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(c.RemoteAddr)
	if err != nil || host == "" {
		return ""
	}
	if isPrivateIP(host) {
		if os.Getenv("NEXTENDO_RELAY_ALLOW_PRIVATE") == "1" {
			return host
		}
		return ""
	}

	return host
}
'@
  if ($text -notlike "*$($old.Substring(0,40))*") {
    # flexible match: only the critical if line
    $patched = $text -replace 'if err != nil || host == "" || isPrivateIP\(host\) \{\r?\n\t\treturn ""\r?\n\t\}', "if err != nil || host == `"`" {`n`t`treturn `"`"`n`t}`n`tif isPrivateIP(host) {`n`t`tif os.Getenv(`"NEXTENDO_RELAY_ALLOW_PRIVATE`") == `"1`" {`n`t`t`treturn host`n`t`t}`n`t`treturn `"`"`n`t}"
    if ($patched -eq $text) {
      Write-Host "WARNING: pattern not found in $($d.Name) — open relaypair.go and patch PublicIPOf manually"
      continue
    }
    Set-Content -Path $file -Value $patched -NoNewline
  } else {
    Set-Content -Path $file -Value ($text.Replace($old, $new)) -NoNewline
  }
  Write-Host "Patched: $($d.Name)\relaypair.go"
}

Write-Host ""
Write-Host "Done. Now start the server with:"
Write-Host '  $env:NEXTENDO_RELAY_ALLOW_PRIVATE = "1"'
Write-Host "  (main.go v9 sets this automatically)"
Write-Host "  go run ."
