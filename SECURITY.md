# Zervyra Vault 1.1.1 — Security

## Kriptografski model

- AES-256-GCM za povjerljivost **i integritet** vault payload-a
- 16-byte nasumični salt
- 12-byte nasumični GCM nonce
- PBKDF2-HMAC-SHA256, 600.000 iteracija u native V1 formatu
- stroge granice KDF parametara prije derivacije ključa
- master lozinke dulje od HMAC-SHA256 block-sizea se pre-hashiraju jednom, što je ekvivalent HMAC key obradi i sprječava ekstremno usporavanje na vrlo dugim unosima
- PBKDF2 u 1.1.1 ponovno koristi HMAC stanje i fiksne U/T buffere kako bi uklonio velik broj privremenih alokacija; iteracije, rezultat i kompatibilnost nisu smanjeni ni promijenjeni
- PBKDF2 kompatibilnost je zaključana poznatim HMAC-SHA256 test-vektorima
- salt/nonce/payload duljine se provjeravaju prije AES-GCM dekripcije
- plaintext i encrypted/Base64 envelope imaju odvojene veličinske limite kako Base64 ekspanzija ne bi valjan vault pretvorila u nečitljiv on-disk zapis
- pogrešna master lozinka i izmijenjeni ciphertext završavaju autentikacijskom greškom, ne djelomično dekriptiranim podacima

OWASP za nove sustave preferira memory-hard Argon2id. Zervyra native V1 namjerno ne uvodi vlastitu/neprovjerenu Argon2 implementaciju samo radi oznake algoritma; PBKDF2-HMAC-SHA256 konfiguracija ostaje kompatibilna i na 600k razini. Budući V2 treba prijeći na Argon2id samo uz provjerenu biblioteku i sigurnu V1→V2 migraciju.

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
- ako početna redundantna recovery kopija zakaže nakon što je main vault već verificiran, verificirani main vault se **ne briše**; UI ga otvara i korisnika upozorava da napravi vanjski backup
- privremene backup/snapshot kopije koriste jedinstvene `CreateTemp` nazive, fsync i atomic replace
- ručni backup se stvara iz aktualnog in-memory modela, ne iz potencijalno zastarjelog main filea
- backup se odmah ponovno dekriptira radi provjere

## Runtime hardening

- Win32 message loop trajno je na jednom OS threadu (`runtime.LockOSThread`)
- nema background goroutine koja zove `user32.dll` za clipboard
- clipboard se briše samo ako korisnik u međuvremenu nije kopirao drugi sadržaj
- ako je clipboard zauzet u trenutku isteka, aplikacija retrya umjesto da odustane
- process-aware lock sprečava drugu aktivnu instancu da piše isti vault
- lock od 1.1.1 ima jedinstveni ownership token; instanca može osvježiti ili obrisati samo lock koji još uvijek pripada njoj
- dead-PID lock se može oporaviti, dok se nov/malformed lock ne uklanja agresivno
- autosave prije sleep/shutdown scenarija gdje Windows daje priliku aplikaciji za spremanje
- automatski lock pri suspendu/minimiziranju/neaktivnosti
- URL launcher dopušta samo HTTP/HTTPS
- lokalni log ne zapisuje lozinke, TOTP tajne ni sadržaj zapisa

## Build i supply-chain hardening

- pull request prema `main` pokreće gofmt provjeru, nekeshirane testove, core race detector, `go vet` i Windows x64 build
- PR workflow koristi read-only repository permission i ne može objaviti release
- `contents: write` dodjeljuje se samo zasebnom publish jobu na `main`
- finalni EXE artefakti provjeravaju se kao PE/MZ datoteke i uspoređuju sa SHA-256 manifestom prije objave
- Setup verificira embedded aplikaciju, uninstaller i ikonu nakon instalacije

## Poznata ograničenja

- Go `string` nije secure-memory container; OS/runtime može kopirati podatke u memoriji
- PBKDF2 se radi sinkrono pri save/autosave operaciji; 1.1.1 znatno smanjuje CPU/alokacijski overhead, ali budući session-key dizajn može dodatno smanjiti UI latenciju bez slabljenja KDF-a
- binarije nisu Authenticode potpisane dok se ne doda code-signing certifikat
- fizički gubitak/kvar cijelog diska zahtijeva backup na drugom uređaju
- potreban je neovisni security audit prije tvrdnje da je proizvod prikladan za visoko-rizične enterprise scenarije

Prijava ranjivosti: `info@brendigo.com`.

## Immediate encrypted generations

Od 1.1.0 svako spremanje zadržava do tri neposredne prethodne valjane šifrirane generacije (`.prev1`–`.prev3`) uz write-ahead `.recovery` i dugoročniji `.bak1`–`.bak10` lanac. Te datoteke su jednako osjetljive kao glavni vault i moraju ostati privatne.

Generator lozinki je fail-closed: CSPRNG greška prekida generiranje umjesto vraćanja djelomično predvidljivog rezultata.
