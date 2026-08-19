$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$Native = $PSScriptRoot
$Root = (Resolve-Path (Join-Path $Native "..")).Path
$Release = Join-Path $Root "release"
$Version = (Get-Content (Join-Path $Root "VERSION") -Raw).Trim()
if ($Version -notmatch '^\d+\.\d+\.\d+$') { throw "Neispravan VERSION: $Version" }

Set-Location $Native
New-Item -ItemType Directory -Force $Release | Out-Null
Get-ChildItem $Release -Filter "Zervyra-Vault-*.exe" -ErrorAction SilentlyContinue | Remove-Item -Force -ErrorAction SilentlyContinue

& go test -count=1 ./...
if ($LASTEXITCODE -ne 0) { throw "Native core testovi nisu prosli." }

& go test -race -count=1 -run 'Test(CloneVaultIsDeepCopy|PasswordHistory|EntryRevisionRestoresIdentityAndSecretFields|SafeURL|SecureCharRejectsEmptyAlphabet)$' ./internal/core
if ($LASTEXITCODE -ne 0) { throw "Race-detector smoke test nije prosao." }

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

& go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet nije prosao." }

& go build -trimpath -ldflags "-H windowsgui -s -w -X main.version=$Version -X main.portableBuild=false" -o (Join-Path $Release "Zervyra-Vault-$Version.exe") ./cmd/vault
& go build -trimpath -ldflags "-H windowsgui -s -w -X main.version=$Version -X main.portableBuild=true" -o (Join-Path $Release "Zervyra-Vault-Portable-$Version.exe") ./cmd/vault
& go build -trimpath -ldflags "-H windowsgui -s -w" -o (Join-Path $Release "Zervyra-Vault-Uninstall-$Version.exe") ./cmd/uninstall

Copy-Item (Join-Path $Release "Zervyra-Vault-$Version.exe") (Join-Path $Native "cmd\setup\Zervyra-Vault.exe") -Force
Copy-Item (Join-Path $Release "Zervyra-Vault-Uninstall-$Version.exe") (Join-Path $Native "cmd\setup\Zervyra-Vault-Uninstall.exe") -Force
$IconParts = Get-ChildItem (Join-Path $Native "cmd\vault\assets") -Filter "zervyra.ico.b64.part*" | Sort-Object Name
$IconB64 = (($IconParts | ForEach-Object { (Get-Content $_.FullName -Raw).Trim() }) -join "")
[IO.File]::WriteAllBytes((Join-Path $Native "cmd\setup\Zervyra.ico"), [Convert]::FromBase64String($IconB64))
& go build -tags installer -trimpath -ldflags "-H windowsgui -s -w -X main.version=$Version" -o (Join-Path $Release "Zervyra-Vault-Setup-$Version.exe") ./cmd/setup
Remove-Item -Force (Join-Path $Native "cmd\setup\Zervyra-Vault.exe"), (Join-Path $Native "cmd\setup\Zervyra-Vault-Uninstall.exe"), (Join-Path $Native "cmd\setup\Zervyra.ico") -ErrorAction SilentlyContinue

$items = Get-ChildItem $Release -Filter "*.exe" | Sort-Object Name
$lines = foreach ($item in $items) {
    $hash = (Get-FileHash $item.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
    "$hash  $($item.Name)"
}
$lines | Set-Content -Encoding ASCII (Join-Path $Release "SHA256SUMS.txt")
$items | Format-Table Name,Length -AutoSize
