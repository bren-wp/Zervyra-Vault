# Zervyra Vault 1.0.0 — dubinski pregled

## Riješeni P0/P1 problemi

1. **P0 data-loss: Novi trezor je mogao prepisati postojeći vault.**  
   Novi `core.CreateNew` koristi `O_CREATE|O_EXCL`; postojeća datoteka ostaje netaknuta čak i kod TOCTOU racea.

2. **P1 recovery: external backup je kopirao main file i mogao propustiti noviju in-memory recovery verziju.**  
   Backup se sada ponovno šifrira iz aktualnog `Vault` modela i odmah dekriptira radi provjere.

3. **P1 lock correctness: samo timestamp locka dopuštao je teoretski drugi writer ako aktivna instanca duže ne osvježi heartbeat.**  
   Lock sada sadrži PID i provjerava je li proces još živ; dead-PID lock se oporavlja.

4. **P1 recovery granularity: password history nije štitio slučajno prepisani e-mail/username/URL/note/TOTP.**  
   Dodane su pune record revisions i UI opcija `Vrati izmjenu`.

5. **P1 clipboard: ako je clipboard bio zauzet točno u trenutku čišćenja, timer se gasio i tajna je mogla ostati.**  
   Sada se retrya svakih ~500 ms dok je sadržaj još Zervyrin.

6. **P1 lifecycle: suspend/shutdown mogao je doći između editiranja i sljedećeg timer autosavea.**  
   Dodani su `WM_QUERYENDSESSION`, `WM_ENDSESSION` i `WM_POWERBROADCAST` hookovi.

7. **P1 compatibility: stroga nova master-length UI provjera može u budućnosti zaključati legitimni legacy vault s kraćom starom lozinkom.**  
   Minimum se primjenjuje samo pri stvaranju novog trezora; postojeći vault se validira kriptografski.

## Dodatni nalazi i popravci

- `STATIC` kontrole su bile spremljene pod internim map ID-ovima, ali su u Win32 stvarno kreirane kao ID=0; theme routing je zato bio netočan. Svi statics sada dobivaju stvarne ID-eve.
- vrlo duga master lozinka prije je mogla biti pre-hashirana unutar svakog HMAC stvaranja kroz PBKDF2 iteracije; key material se sada jednom pre-hashira kada prelazi SHA-256 block size.
- Setup sada radi post-write integritet provjeru embedded komponenti.
- Uninstaller child cleanup proces mijenja working directory u `%TEMP%` prije `rmdir` operacije.
- default path detektira prethodne Velunox/Brendigo lokacije kako rebrand ne bi izgledao kao nestanak vaulta.

## Automatizirane provjere

- 21 core/hardening test
- Linux native test izvršenje
- Windows x64 `go vet ./...`
- Windows cross-build svih četiriju EXE artefakata
- PE32+ / AMD64 header provjera
- SHA-256 manifest release artefakata

## Preostalo prije enterprise security tvrdnji

- Authenticode code signing
- neovisni penetration/security audit
- memory-hard Argon2id V2 format kroz provjerenu biblioteku i kontroliranu migraciju V1→V2
- Windows Hello/TPM re-unlock
