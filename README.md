# Razpravljalnica

Razpravljalnica je porazdeljen sistem za sporočanje, zasnovan na arhitekturi **verižne replikacije** (chain replication). Sistem zagotavlja visoko razpoložljivost, konsistentnost podatkov in porazdeljevanje obremenitev za naročnine.

## Arhitektura sistema

Sistem sestavljajo tri glavne komponente:

1.  **Control Plane (Nadzorna plošča):** Skrbi za upravljanje verige vozlišč. Spremlja zdravje vozlišč in določa, katero vozlišče je GLAVA (Head) in katero REP (Tail).
2.  **Node (Vozlišče):** Shranjuje podatke in izvaja replikacijo.
    *   **GLAVA (Head):** Sprejema vse zapise (novi uporabniki, teme, sporočila).
    *   **REP (Tail):** Sprejema vse bralne zahteve in služi kot razpošiljatelj za naročnine.
    *   **Replikacija:** Podatki potujejo od glave skozi vmesna vozlišča do repa.
3.  **Client (Odjemalec):** TUI za interakcijo z razpravljalnico.

## Posebne funkcije

*   **Verižna replikacija:** Zagotavlja, da so podatki varno podvojeni na vseh vozliščih, preden je operacija potrjena.
*   **Samodejna sinhronizacija:** Nova vozlišča ob vstopu v verigo samodejno prenesejo manjkajočo zgodovino operacij od svojih predhodnikov.
*   **Porazdeljene naročnine:** Naročnine na teme se porazdelijo med vsa vozlišča
*   **Interaktivni nadzor:** Tako nadzorna plošča kot vozlišča imajo lasten TUI za spremljanje stanja verige in logov

## Navodila za zagon

### 1. Prevajanje projekta
Preverite, ali imate nameščen Go, in zaženite:
```bash
go build ./...
```

### 2. Zagon nadzorne plošče (Control Plane)
```bash
go run cmd/controlplane/main.go -port 12345
```

### 3. Zagon vozlišč (Nodes)
Zaženite več vozlišč v ločenih terminalih. Vsako mora imeti svoja vrata in unikaten ID:
```bash
go run cmd/node/main.go -port 5001 -id node-1 -cp localhost:12345
go run cmd/node/main.go -port 5002 -id node-2 -cp localhost:12345
go run cmd/node/main.go -port 5003 -id node-3 -cp localhost:12345
```

### 4. Zagon odjemalca (Client)
```bash
go run cmd/client/main.go -cp localhost:12345
```

## Uporaba odjemalca
*   **Prijava:** Vnesite uporabniško ime. Če uporabnik že obstaja, se boste prijavili vanj.
*   **Navigacija:** Uporabljajte puščice in tipko `Enter` za vstop v teme.
*   **Sporočanje:** Tipkajte neposredno v vnosno polje.
*   **Všečkanje:** Pritisnite `F2` na izbranem sporočilu.
*   **Osveževanje:** Pritisnite `R` za osvežitev seznama tem.
