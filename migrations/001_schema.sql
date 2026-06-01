-- BarterSwap — schema initial.
-- Ce fichier est rejoué à chaque démarrage : toutes les instructions doivent
-- être idempotentes (CREATE ... IF NOT EXISTS, contraintes inline dans CREATE TABLE).
-- Pas d'ALTER ADD CONSTRAINT IF NOT EXISTS car Postgres ne le supporte pas.

-- ─── users ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    pseudo          TEXT        NOT NULL UNIQUE,
    bio             TEXT        NOT NULL DEFAULT '',
    ville           TEXT        NOT NULL DEFAULT '',
    credit_balance  INT         NOT NULL DEFAULT 10 CHECK (credit_balance >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── skills ────────────────────────────────────────────────────────────────
-- Composite PK (user_id, nom) : un même utilisateur ne peut pas déclarer la
-- même compétence deux fois. PUT /api/users/{id}/skills écrase complètement
-- les lignes (DELETE puis INSERT dans une transaction côté usecase).
CREATE TABLE IF NOT EXISTS skills (
    user_id  INT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nom      TEXT NOT NULL CHECK (length(nom) > 0),
    niveau   TEXT NOT NULL CHECK (niveau IN ('debutant', 'intermediaire', 'expert')),
    PRIMARY KEY (user_id, nom)
);

-- ─── services ──────────────────────────────────────────────────────────────
-- Liste fermée des catégories validée par contrainte CHECK. Pour ajouter une
-- catégorie il faudra une migration ALTER TABLE DROP CONSTRAINT + ADD.
CREATE TABLE IF NOT EXISTS services (
    id            SERIAL      PRIMARY KEY,
    provider_id   INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titre         TEXT        NOT NULL CHECK (length(titre) > 0),
    description   TEXT        NOT NULL DEFAULT '',
    categorie     TEXT        NOT NULL CHECK (categorie IN (
        'Informatique', 'Jardinage', 'Bricolage', 'Cuisine', 'Musique',
        'Langues', 'Sport', 'Tutorat', 'Demenagement', 'Photographie',
        'Animalier', 'Couture', 'Autre'
    )),
    duree_minutes INT         NOT NULL CHECK (duree_minutes > 0),
    credits       INT         NOT NULL CHECK (credits > 0),
    ville         TEXT        NOT NULL DEFAULT '',
    actif         BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index filtrés sur services actifs (la majorité des SELECT filtrent dessus).
CREATE INDEX IF NOT EXISTS idx_services_categorie ON services(categorie) WHERE actif;
CREATE INDEX IF NOT EXISTS idx_services_ville     ON services(ville)     WHERE actif;
CREATE INDEX IF NOT EXISTS idx_services_provider  ON services(provider_id);

-- ─── exchanges ─────────────────────────────────────────────────────────────
-- credits : snapshot du coût au moment de la demande (le service peut être
-- modifié ou supprimé après, le contrat de l'échange reste figé).
-- CHECK requester_id <> owner_id : impossible de s'échanger un service à soi-même.
CREATE TABLE IF NOT EXISTS exchanges (
    id           SERIAL      PRIMARY KEY,
    service_id   INT         NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    requester_id INT         NOT NULL REFERENCES users(id)    ON DELETE RESTRICT,
    owner_id     INT         NOT NULL REFERENCES users(id)    ON DELETE RESTRICT,
    credits      INT         NOT NULL CHECK (credits > 0),
    status       TEXT        NOT NULL CHECK (status IN (
        'pending', 'accepted', 'rejected', 'cancelled', 'completed'
    )),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (requester_id <> owner_id)
);

-- Verrou métier critique : un service ne peut avoir qu'UN seul échange actif
-- (pending ou accepted) à la fois. Index unique partiel = la DB rejette tout
-- INSERT concurrent en violation, l'app mappe 23505 vers ErrServiceAlreadyBooked.
CREATE UNIQUE INDEX IF NOT EXISTS uq_exchange_active_per_service
    ON exchanges(service_id)
    WHERE status IN ('pending', 'accepted');

CREATE INDEX IF NOT EXISTS idx_exchanges_requester ON exchanges(requester_id);
CREATE INDEX IF NOT EXISTS idx_exchanges_owner     ON exchanges(owner_id);
CREATE INDEX IF NOT EXISTS idx_exchanges_status    ON exchanges(status);

-- ─── credit_transactions ───────────────────────────────────────────────────
-- Journal append-only des mouvements de crédit. Source de vérité auditable :
-- SUM(montant) WHERE user_id=X doit toujours = users.credit_balance.
-- Types :
--   'welcome' : crédits offerts à l'inscription (+10)
--   'spend'   : crédits débités à l'acceptation d'un échange (négatif)
--   'earn'    : crédits crédités à la complétion d'un échange (positif)
--   'refund'  : crédits restitués au demandeur en cas de cancel/reject (positif)
-- exchange_id NULL pour les opérations 'welcome', NOT NULL sinon (vérifié en app).
CREATE TABLE IF NOT EXISTS credit_transactions (
    id          SERIAL      PRIMARY KEY,
    user_id     INT         NOT NULL REFERENCES users(id)     ON DELETE RESTRICT,
    exchange_id INT                  REFERENCES exchanges(id) ON DELETE RESTRICT,
    montant     INT         NOT NULL,
    type        TEXT        NOT NULL CHECK (type IN ('welcome', 'spend', 'earn', 'refund')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_credit_tx_user     ON credit_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_tx_exchange ON credit_transactions(exchange_id);

-- ─── reviews ───────────────────────────────────────────────────────────────
-- UNIQUE (exchange_id, author_id) : un auteur ne peut laisser qu'un seul avis
-- par échange. INSERT en double → 23505 mappé vers ErrAlreadyReviewed.
CREATE TABLE IF NOT EXISTS reviews (
    id          SERIAL      PRIMARY KEY,
    exchange_id INT         NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,
    author_id   INT         NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    target_id   INT         NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    note        INT         NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (exchange_id, author_id),
    CHECK (author_id <> target_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_target   ON reviews(target_id);
CREATE INDEX IF NOT EXISTS idx_reviews_exchange ON reviews(exchange_id);
