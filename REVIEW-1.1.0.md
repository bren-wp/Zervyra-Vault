# Zervyra Vault 1.1.0 — dubinski pregled

Datum: 2026-08-19

## Rezultat

- 25/25 core i hardening testova prolazi (`go test -count=1 ./...`)
- Windows x64 `go vet ./...` prolazi
- race-detector smoke test prolazi za mutation/revision/URL/generator helper sloj
- sva četiri lokalno cross-compileana artefakta imaju PE32+ x86-64 GUI format

## Ispravljeni problemi

1. **Portable detekcija** — standalone Portable EXE više ne ovisi samo o build-time flagu; prepoznaje i `Portable` u imenu te `portable.flag`.
2. **Zbunjujući prvi start** — nepostojeći vault više se ne prikazuje kao zaključani postojeći vault.
3. **Prekratka recovery povijest** — uz `.recovery` i `.bak1`–`.bak10` sada postoje `.prev1`–`.prev3` generacije pri svakom save/autosave ciklusu.
4. **Premalo record revizija tijekom duge sesije uređivanja** — nakon svakog uspješnog persistence pointa sljedeća izmjena dobiva novu full-record reviziju.
5. **Generator fail-open rubni slučaj** — CSPRNG greška više ne može vratiti djelomično deterministički rezultat.
6. **Elevated-process lock rubni slučaj** — `OpenProcess(ERROR_ACCESS_DENIED)` konzervativno se tretira kao živ proces, pa druga instanca ne uklanja tuđi lock.
7. **Nepotreban ručni path workflow** — `Novi trezor` uz postojeću putanju otvara Save-As, dok `CreateNew` i dalje koristi `O_EXCL`.
8. **Neograničen app.log** — log se rotira nakon 1 MiB.
9. **Razilaženje verzija builda** — jedna `VERSION` datoteka upravlja nazivima build artefakata.
10. **Zaboravljena custom vault putanja** — aplikacija pamti samo zadnju putanju (bez tajni), pa je ponovno otvaranje jednostavnije.

## Sigurnosna ograničenja koja ostaju

- native V1 format i dalje koristi PBKDF2-HMAC-SHA256 600k radi kompatibilnosti; budući V2 migration treba koristiti provjerenu Argon2id biblioteku, ne vlastitu implementaciju.
- Go stringovi nisu secure-memory container.
- EXE datoteke nisu Authenticode potpisane bez code-signing certifikata.
- profesionalni neovisni security audit i dalje je potreban prije enterprise/high-risk tvrdnji.
