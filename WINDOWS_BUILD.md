# Windows build — Zervyra Vault 1.1.0

## Zahtjevi

- Windows 10/11 x64
- Go 1.23+
- PowerShell 5.1+

Pokreni `build_windows.bat` ili:

```powershell
./native/build_windows.ps1
```

Build redoslijed:

1. `go test ./...`
2. Windows `go vet ./...`
3. standalone x64 PE build
4. portable x64 PE build
5. uninstaller
6. self-contained setup koji embedda aplikaciju, uninstaller i Zervyra ICO
7. SHA-256 manifest

Build se prekida ako testovi ili vet ne prođu.

Verzija se čita iz korijenske `VERSION` datoteke. Build zaustavlja izdavanje ako testovi, race smoke test ili Windows `go vet` ne prođu.
