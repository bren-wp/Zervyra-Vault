# Changelog

## 1.1.0 — 2026-08-19

### Stabilnost i jednostavniji prvi start
- Portable EXE sada se prepoznaje i po nazivu datoteke (`Portable`) te opcionalnom `portable.flag`, ne samo build-time varijabli
- prvi start više ne prikazuje zbunjujuće “trezor zaključan” kada trezor još ne postoji
- zaključani ekran razlikuje postojeći vault od prvog pokretanja i skriva nepotrebnu potvrdu master lozinke pri običnom unlocku
- `Novi trezor` uz postojeći vault otvara normalni Save-As dijalog umjesto da traži ručno uređivanje putanje
- pamti se samo putanja zadnjeg vaulta (bez tajni) radi jednostavnijeg ponovnog otvaranja
- log se rotira na 1 MiB umjesto neograničenog rasta

### Anti-data-loss
- tri neposredne šifrirane `.prev1`–`.prev3` generacije čuvaju se pri svakom uspješnom save/autosave ciklusu
- `LoadBest` sada oporavlja i iz neposrednih prethodnih generacija, uz `.recovery` i 10 rotirajućih backupa
- nakon svakog uspješnog autosavea sljedeća izmjena dobiva novu full-record reviziju; više se ne čuva samo jedna revizija po cijeloj sesiji uređivanja
- trenutačna putanja vaulta ostaje poznata nakon zaključavanja, ali se master lozinka i dekriptirani sadržaj iz memorijskog modela i dalje odbacuju

### Sigurnost
- generator lozinki sada radi fail-closed: ako CSPRNG zakaže u bilo kojoj fazi, ne vraća djelomično determinističku lozinku
- Windows lock provjera konzervativno tretira `ACCESS_DENIED` kao živ proces kako niži privilege proces ne bi uklonio lock povišene instance
- build skripta pokreće nekeshirane testove, race detector za core i Windows `go vet` prije izrade releasea
- verzija builda dolazi iz jedne `VERSION` datoteke kako se installer/portable/release nazivi ne bi razišli

## 1.0.0 — 2026-08-18

### Novi brand i dizajn
- novi proizvodni naziv **Zervyra Vault**
- brand sustav, SVG/PNG logo i multi-size ICO
- tamniji Obsidian/Iris UI umjesto klasičnog sivog Win32 izgleda
- odvojeni jednostavni locked/unlocked prikaz; login kontrole više ne zatrpavaju glavni editor
- tamna DWM title bar, zaobljene kartice i custom flat buttons
- brand ikona se embedda u aplikaciju i učitava za titlebar/taskbar
- jasnija polja i zaseban copy e-mail workflow

### Stabilnost
- ispravljeni stvarni ID-evi `STATIC` kontrola; prethodno su vizualni stilovi dobivali ID=0
- UI i svi user32 clipboard pozivi ostaju na istom OS threadu
- process-aware vault lock provjerava živi PID, a ne samo starost `.lock` datoteke
- cleanup install foldera više ne koristi install directory kao child-process working directory
- Setup verificira instalirani EXE/uninstaller/icon SHA-256 sadržajno prije uspješne potvrde
- shutdown/suspend handling dodan u Win32 lifecycle
- release build zahtijeva testove + Windows `go vet`

### Anti-data-loss
- `CreateNew` nikada ne overwrita postojeći vault ni u filesystem raceu
- write-ahead recovery + atomic replace + post-write decrypt verify
- 10 rotirajućih šifriranih backup generacija
- autosave nakon približno 1 s mirovanja
- puni record revisions: vraćanje prethodnog maila, korisničkog imena, lozinke, URL-a, 2FA, tagova i bilješki
- external backup sada se gradi iz najnovijeg in-memory vaulta, ne nužno iz main filea
- automatski legacy path fallback za Velunox/Brendigo default vault
- clipboard clear retry ako ga drugi proces privremeno drži otvorenim

### Sigurnost
- AES-256-GCM + PBKDF2-HMAC-SHA256 600k zadržani radi native V1 kompatibilnosti
- long-master HMAC key pre-hash optimizacija uklanja potencijalni CPU DoS uz očuvanje istog PBKDF2 rezultata
- stroge envelope/field/revision granice
- HTTP(S)-only URL policy
- master minimum primjenjuje se pri kreiranju novog vaulta; postojeći legitimni legacy vault nije blokiran novijom UI politikom

### Održavanje
- jedna kanonska native Go implementacija; nema paralelnog Python release runtimea
- brand, build, setup, portable i CI nazivi objedinjeni na Zervyra
