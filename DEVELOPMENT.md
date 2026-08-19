# Zervyra Vault — razvojna pravila

## Kanonski repozitorij

Jedino kanonsko mjesto za razvoj Zervyra Vaulta je:

`bren-wp/Zervyra-Vault`

Sve nove verzije, sigurnosni popravci, UI promjene, testovi, dokumentacija i release pipeline moraju završiti u ovom repozitoriju. Stari Brendigo/Velunox repozitoriji ili grane nisu izvor novih Zervyra izdanja.

## Main

`main` mora ostati buildable i predstavljati trenutno prihvaćenu verziju. Promjene se prvo provjeravaju testovima i Windows build provjerama, zatim spajaju u `main`.

## Release artefakti

Windows Setup, Portable, standalone EXE, Uninstaller, source ZIP, Windows ZIP, FULL ZIP i SHA-256 manifest objavljuju se kroz GitHub Releases ovog istog repozitorija. Binarni release artefakti ne trebaju se trajno spremati u Git source history jer su reproducibilni iz sourcea i workflowa.

## Sigurnost podataka

Nikada ne commitati stvarne korisničke `.bvault`, `.recovery`, `.prev*`, `.bak*`, `.lock`, lokalne logove ili preference. `.gitignore` ih mora blokirati.

## Verzije

Korijenska `VERSION` datoteka je jedini izvor broja verzije za build i release nazive.
