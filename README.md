# BarterSwap — API d'échange de compétences

API REST écrite en Go (stdlib pure) qui implémente une **banque de temps** entre particuliers : chaque heure de service rendue donne droit à une heure de service reçue. Pas d'argent, pas de likes, pas de fil d'actualité.

## Architecture

Le projet utilise une **Clean Architecture** matérialisée dans un seul package Go (`main`), avec séparation par préfixes de fichiers :

| Couche | Fichiers | Imports autorisés |
|---|---|---|
| Entities (cœur métier pur) | `entity_*.go` | stdlib uniquement, zéro infra |
| Ports (interfaces) | `ports.go` | `context`, entities |
| Use cases (orchestration) | `usecase_*.go` | `context`, entities, ports |
| HTTP adapters (in) | `http_*.go` | `net/http`, `encoding/json`, use cases |
| SQL adapters (out) | `repo_*.go` | `database/sql`, ports, entities |

**Dépendances** : Go stdlib + `github.com/jackc/pgx/v5/stdlib` (driver PostgreSQL, seule lib externe autorisée par le sujet).

## Installation

Tout tourne dans Docker — pas besoin d'avoir Go installé localement.

```bash
git clone https://github.com/nassimlnd/barter-swap
cd barter-swap
make up
```

L'API répond sur `http://localhost:8080`. Vérifie :

```bash
curl -s http://localhost:8080/health
# {"status":"ok","db":"up"}
```

Le mode dev utilise **Air** pour le live reload : enregistre un fichier `.go`, le binaire est recompilé en ~2s.

## Commandes utiles

```bash
make up         # Build + démarre db + api (dev avec live reload)
make down       # Stop + supprime volumes
make logs       # Tail api logs
make psql       # Shell psql sur la db
make sh         # Shell dans le container api
make test       # Tests dans un container isolé (DB tmpfs)
make rebuild    # Down + up --build complet
make help       # Liste toutes les cibles
```

## Endpoints

(Tableau récapitulatif détaillé une fois les jalons terminés.)

| Méthode | Path | Description |
|---|---|---|
| GET | `/health` | Healthcheck + ping DB |
| POST | `/api/users` | Créer un compte (+10 crédits offerts) |
| GET | `/api/users/{id}` | Profil utilisateur |
| ... | | |

## Exemples curl

(À venir.)

## Tests

```bash
make test
# ou en local si Go est installé :
make test-local
```

Cible de couverture : **75%** (le sujet exige ≥ 60%).

## Variables d'environnement

| Variable | Défaut | Description |
|---|---|---|
| `DATABASE_URL` | — | DSN Postgres (`postgres://user:pass@host:port/db?sslmode=disable`) |
| `PORT` | `8080` | Port d'écoute HTTP |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |

## Structure du dépôt

```
barterswap/
├── *.go                     # un seul package main (Clean par préfixes de fichiers)
├── migrations/              # DDL SQL joué au démarrage si DB vide
├── api/                     # openapi.yaml + Swagger UI embed (à venir)
├── scripts/                 # helpers shell (optionnels)
├── testdata/                # fixtures de tests
├── docs/                    # subject.md + docs internes
├── Dockerfile               # image multi-stage runtime
├── docker-compose.yml       # db + api (prod-like)
├── docker-compose.override.yml  # overlay dev (Air live reload)
├── docker-compose.test.yml  # overlay tests (DB tmpfs)
└── Makefile
```
