# Patches nextendo-nex so pair-relay works on LAN (private IPs).
# Run once from the dying-light-nex folder, then: go run .

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

$needle = 'if err != nil || host == "" || isPrivateIP(host) {'
$replacement = @'
if err != nil || host == "" {
		return ""
	}
	if isPrivateIP(host) {
		if os.Getenv("NEXTENDO_RELAY_ALLOW_PRIVATE") == "1" {
			return host
		}
'@

$anyPatched = $false
foreach ($d in $dirs) {
  $file = Join-Path $d.FullName "relaypair.go"
  if (-not (Test-Path $file)) { continue }

  $text = [System.IO.File]::ReadAllText($file)
  if ($text -match "NEXTENDO_RELAY_ALLOW_PRIVATE") {
    Write-Host ("Already patched: " + $d.Name)
    $anyPatched = $true
    continue
  }

  if ($text -notmatch [regex]::Escape($needle)) {
    Write-Host ("WARNING: pattern not found in " + $d.Name)
    continue
  }

  # Clear read-only attribute Go sets on module cache files
  $item = Get-Item $file
  if ($item.IsReadOnly) { $item.IsReadOnly = $false }

  $newText = $text.Replace($needle, $replacement.TrimEnd())
  [System.IO.File]::WriteAllText($file, $newText)
  Write-Host ("Patched: " + $d.Name + "\relaypair.go")
  $anyPatched = $true
}

Write-Host ""
if ($anyPatched) {
  Write-Host "Done. Start the server with go run ."
} else {
  Write-Host "Nothing patched. Paste any WARNING lines above."
}
