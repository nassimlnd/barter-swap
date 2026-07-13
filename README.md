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

| Méthode | Path | Description |
|---|---|---|
| GET | `/health` | Healthcheck + ping DB |
| POST | `/api/users` | Créer un compte (+10 crédits offerts) |
| GET | `/api/users/{id}` | Profil public |
| PUT | `/api/users/{id}` | Modifier son profil (`X-UserID`) |
| GET | `/api/users/{id}/skills` | Compétences utilisateur |
| PUT | `/api/users/{id}/skills` | Remplacer ses compétences (`X-UserID`) |
| GET | `/api/services` | Liste paginée + filtres `categorie`, `ville`, `search` |
| POST | `/api/services` | Créer une annonce (`X-UserID`) |
| GET | `/api/services/{id}` | Détail service |
| PUT | `/api/services/{id}` | Modifier son annonce (`X-UserID`) |
| DELETE | `/api/services/{id}` | Supprimer son annonce (`X-UserID`) |
| POST | `/api/exchanges` | Demander un échange (`X-UserID`) |
| GET | `/api/exchanges` | Mes échanges, option `status` (`X-UserID`) |
| GET | `/api/exchanges/{id}` | Détail échange participant (`X-UserID`) |
| PUT | `/api/exchanges/{id}/accept` | Accepter et bloquer crédits (`X-UserID`) |
| PUT | `/api/exchanges/{id}/reject` | Refuser (`X-UserID`) |
| PUT | `/api/exchanges/{id}/complete` | Terminer et transférer crédits (`X-UserID`) |
| PUT | `/api/exchanges/{id}/cancel` | Annuler et rembourser si besoin (`X-UserID`) |
| POST | `/api/exchanges/{id}/review` | Laisser un avis (`X-UserID`) |
| GET | `/api/users/{id}/reviews` | Avis reçus par utilisateur |
| GET | `/api/services/{id}/reviews` | Avis liés à un service |
| GET | `/api/users/{id}/stats` | Stats utilisateur (`X-UserID`) |

## Exemples curl

Créer deux utilisateurs :

```bash
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":"alice","ville":"Paris"}'

curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":"bob","ville":"Paris"}'
```

Déclarer les compétences d'Alice, puis publier un service :

```bash
curl -X PUT http://localhost:8080/api/users/1/skills \
  -H 'Content-Type: application/json' \
  -H 'X-UserID: 1' \
  -d '{"skills":[{"nom":"Jardinage","niveau":"expert"}]}'

curl -X POST http://localhost:8080/api/services \
  -H 'Content-Type: application/json' \
  -H 'X-UserID: 1' \
  -d '{"titre":"Tondre une pelouse","categorie":"Jardinage","duree_minutes":60,"credits":1,"ville":"Paris"}'
```

Bob demande l'échange, Alice accepte puis termine :

```bash
curl -X POST http://localhost:8080/api/exchanges \
  -H 'Content-Type: application/json' \
  -H 'X-UserID: 2' \
  -d '{"service_id":1}'

curl -X PUT http://localhost:8080/api/exchanges/1/accept -H 'X-UserID: 1'
curl -X PUT http://localhost:8080/api/exchanges/1/complete -H 'X-UserID: 1'
```

Bob laisse un avis et Alice consulte ses stats :

```bash
curl -X POST http://localhost:8080/api/exchanges/1/review \
  -H 'Content-Type: application/json' \
  -H 'X-UserID: 2' \
  -d '{"note":5,"commentaire":"Très efficace"}'

curl http://localhost:8080/api/users/1/stats -H 'X-UserID: 1'
```

OpenAPI minimal : `api/openapi.yaml`.

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
