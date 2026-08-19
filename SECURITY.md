# Zervyra Vault 1.1.0 — Security

## Kriptografski model

- AES-256-GCM za povjerljivost **i integritet** vault payload-a
- 16-byte nasumični salt
- 12-byte nasumični GCM nonce
- PBKDF2-HMAC-SHA256, 600.000 iteracija u native V1 formatu
- stroge granice KDF parametara prije derivacije ključa
- master lozinke dulje od HMAC-SHA256 block-sizea se pre-hashiraju jednom, što je ekvivalent HMAC key obradi i sprječava ekstremno usporavanje na vrlo dugim unosima
- salt/nonce/payload duljine se provjeravaju prije AES-GCM dekripcije
- pogrešna master lozinka i izmijenjeni ciphertext završavaju autentikacijskom greškom, ne djelomično dekriptiranim podacima

OWASP za nove sustave preferira memory-hard Argon2id. Zervyra native V1 namjerno ne uvodi vlastitu/neprovjerenu Argon2 implementaciju samo radi oznake algoritma; PBKDF2-HMAC-SHA256 konfiguracija ostaje kompatibilna i na preporučenoj 600k razini. Budući V2 treba prijeći na Argon2id samo uz provjerenu biblioteku i sigurnu V1→V2 migraciju.

## Zaštita podataka i recovery

- write-ahead šifrirani `.recovery`
- post-write dekripcijska verifikacija
- Windows atomic replace + `WRITE_THROUGH`
- `.prev1`–`.prev3` neposredne šifrirane generacije pri svakom spremanju
- `.bak1`–`.bak10` rotirajuće šifrirane generacije
- `LoadBest` bira najnoviju valjanu generaciju
- do 20 šifriranih punih revizija po zapisu
- destruktivne operacije rade copy → save → commit; aktivni in-memory model se ne objavljuje prije uspješnog spremanja
- novi trezor se stvara s `O_CREATE|O_EXCL` i ne može pregaziti postojeću datoteku
- ručni backup se stvara iz aktualnog in-memory modela, ne iz potencijalno zastarjelog main filea
- backup se odmah ponovno dekriptira radi provjere

## Runtime hardening

- Win32 message loop trajno je na jednom OS threadu (`runtime.LockOSThread`)
- nema background goroutine koja zove `user32.dll` za clipboard
- clipboard se briše samo ako korisnik u međuvremenu nije kopirao drugi sadržaj
- ako je clipboard zauzet u trenutku isteka, aplikacija retrya umjesto da odustane
- process-aware lock sprečava drugu aktivnu instancu da piše isti vault; dead-PID lock se može oporaviti
- autosave prije sleep/shutdown scenarija gdje Windows daje priliku aplikaciji za spremanje
- automatski lock pri suspendu/minimiziranju/neaktivnosti
- URL launcher dopušta samo HTTP/HTTPS
- lokalni log ne zapisuje lozinke, TOTP tajne ni sadržaj zapisa

## Poznata ograničenja

- Go `string` nije secure-memory container; OS/runtime može kopirati podatke u memoriji
- binarije nisu Authenticode potpisane dok se ne doda code-signing certifikat
- fizički gubitak/kvar cijelog diska zahtijeva backup na drugom uređaju
- potreban je neovisni security audit prije tvrdnje da je proizvod prikladan za visoko-rizične enterprise scenarije

Prijava ranjivosti: `info@brendigo.com`.

## Immediate encrypted generations

Od 1.1.0 svako spremanje zadržava do tri neposredne prethodne valjane šifrirane generacije (`.prev1`–`.prev3`) uz write-ahead `.recovery` i dugoročniji `.bak1`–`.bak10` lanac. Te datoteke su jednako osjetljive kao glavni vault i moraju ostati privatne.

Generator lozinki je fail-closed: CSPRNG greška prekida generiranje umjesto vraćanja djelomično predvidljivog rezultata.
