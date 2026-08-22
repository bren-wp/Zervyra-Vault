# Zervyra

**Tvoje tajne. Tvoja kontrola.**

Zervyra Vault je lokalni Windows password manager napravljen kao jedna kanonska native Go aplikacija: bez obaveznog računa, bez obaveznog clouda i bez Python runtimea. Fokus 1.1.x izdanja je jednostavno korištenje, stabilan tamni desktop UI i višeslojna zaštita od slučajnog gubitka e-mailova, korisničkih imena, lozinki, TOTP tajni i bilješki.

## Što dobivaš

- moderan tamni **Zervyra** UI i vlastiti brand/icon
- lokalni šifrirani `.bvault`
- AES-256-GCM autentificiranu enkripciju
- PBKDF2-HMAC-SHA256 s 600.000 iteracija, nasumičnim saltom po zapisu vaulta i zaštitom od modificiranog ciphertexta
- optimizirani PBKDF2 bez promjene rezultata ili sigurnosnog troška: HMAC stanje i fiksni bufferi ponovno se koriste kako bi autosave imao manje alokacija i kraće UI zastajkivanje
- zasebna polja za naziv, korisničko ime, e-mail, lozinku, URL, 2FA/TOTP, tagove i bilješke
- favorite, pretragu i koš
- generator jakih lozinki
- TOTP (`otpauth://`, SHA-1/SHA-256/SHA-512)
- sigurnosni pregled slabih i ponovljenih lozinki
- password history i **povijest cijelog zapisa** za vraćanje slučajno promijenjenog maila/korisnika/lozinke/URL-a/bilješke
- autosave otprilike 1 sekundu nakon prestanka uređivanja
- automatsko zaključavanje nakon 5 minuta neaktivnosti, pri minimiziranju i prije suspenda
- sigurno čišćenje clipboarda uz retry ako je clipboard privremeno zauzet
- provjerenu šifriranu backup kopiju na drugi disk/USB
- Setup, Portable i standalone Windows izdanja

## Zaštita od gubitka podataka

Zervyra ne ovisi o jednoj jedinoj datoteci:

1. prije glavnog zapisa stvara verificiranu šifriranu `*.recovery` generaciju
2. glavni vault zamjenjuje atomskim Windows `MoveFileExW(REPLACE_EXISTING | WRITE_THROUGH)` postupkom
3. nakon spremanja vault se ponovno dekriptira radi provjere integriteta
4. pri svakom save/autosave ciklusu čuva do **3 neposredne prethodne šifrirane generacije** (`.prev1`–`.prev3`)
5. čuva do **10 rotirajućih šifriranih backup generacija** (`.bak1`–`.bak10`)
6. pri otvaranju provjerava main/recovery/prev/backup generacije i bira najnoviju valjanu
7. svaki uređivani zapis može čuvati do **20 punih šifriranih revizija**
8. „Novi trezor” koristi `O_EXCL` i **nikada ne prepisuje postojeći vault**
9. ručni backup se generira iz trenutačnog in-memory vaulta i ponovno verificira nakon zapisa
10. lock datoteka ima vlasnički token, pa stara instanca ne smije obrisati lock nove instance
11. već verificirani novi vault ostaje sačuvan čak i ako naknadno zakaže izrada redundantne recovery kopije; aplikacija ga može nastaviti koristiti uz jasno upozorenje korisniku

Ako je na računalu ostao stari `Velunox Vault` ili `Brendigo Vault` default vault, Zervyra ga prepoznaje kao legacy lokaciju kako promjena brenda ne bi izgledala kao gubitak podataka.

Nijedan lokalni program ne može garantirati oporavak nakon fizičkog kvara/gubitka cijelog diska. Za važne podatke koristi **Backup kopija** i čuvaj je na drugom fizičkom uređaju.

## Windows artefakti

`release/` lokalnog builda sadrži:

- `Zervyra-Vault-Setup-1.1.1.exe`
- `Zervyra-Vault-Portable-1.1.1.exe`
- `Zervyra-Vault-1.1.1.exe`
- `Zervyra-Vault-Uninstall-1.1.1.exe`
- `SHA256SUMS.txt`

Setup verificira instalirane embedded binarije i brand ikonu prije završetka instalacije. Uninstaller uklanja program i shortcutove, ali **ne briše korisnički šifrirani vault**.

Svaki pull request prema `main` sada mora proći gofmt provjeru, nekeshirane testove, puni core race detector, `go vet` i Windows x64 build. Push/ručni build na `main` nakon toga objavljuje verificirane EXE datoteke, Windows ZIP, source ZIP i FULL ZIP u **GitHub Releases** pod tagom trenutne verzije.

## Build i provjera

Potreban je Go 1.23+.

Na Windowsu pokreni:

```bat
build_windows.bat
```

Release build prvo pokreće testove i Windows `go vet`, provjerava svaki build korak i očekuje točno četiri EXE artefakta prije izrade SHA-256 manifesta. Puni race-detector test izvodi GitHub CI na Linux runneru; lokalno ga možeš uključiti postavljanjem `ZERVYRA_RUN_RACE=1` prije build skripte.

## Kompatibilnost

Interni format ID `BRENDIGO_VAULT_NATIVE_V1` namjerno ostaje isti radi čitanja native Brendigo Vault 0.4/0.5 i Velunox vaultova. To je samo stabilni identifikator formata na disku; proizvod i novi buildovi nose brand **Zervyra Vault**.

## Status sigurnosti

1.1.1 ima najmanje 30 automatiziranih core/hardening testova, uključujući kriptografski round-trip, PBKDF2-HMAC-SHA256 kompatibilnosne vektore, pogrešnu master lozinku, tamper rejection, malformed nonce, recovery, backup rotaciju, backup export, lock ownership, očuvanje verificiranog novog vaulta, long-master obradu, TOTP, URL policy i record revisions. Automatizirani testovi nisu zamjena za neovisni profesionalni security audit.

## Zaštita od gubitka podataka u 1.1.1

Svako spremanje prvo provjerava šifrirani recovery zapis i čuva do tri neposredne prethodne generacije (`.prev1`–`.prev3`). Uz to postoje `.recovery` i 10 rotirajućih `.bak` generacija. `LoadBest` pri otključavanju bira najnoviju valjanu šifriranu generaciju. Plaintext limit i on-disk encrypted/Base64 limit sada su odvojeni kako velik, ali valjan vault ne bi postao nečitljiv nakon spremanja.

Portable izdanje je samostalno: ako naziv EXE-a sadrži `Portable`, koristi vlastitu `data` mapu uz aplikaciju i ne čita installed `%LOCALAPPDATA%` stanje.
