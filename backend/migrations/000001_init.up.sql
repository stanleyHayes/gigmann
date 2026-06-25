-- 000001_init.up.sql — initial schema (GEC-11). Render-managed Postgres + pgvector.
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE facilities (
    id              text PRIMARY KEY,
    name            text NOT NULL,
    region          text NOT NULL,
    town            text NOT NULL DEFAULT '',
    type            text NOT NULL DEFAULT '',
    beds            integer NOT NULL DEFAULT 0 CHECK (beds >= 0),
    lifecycle       text NOT NULL CHECK (lifecycle IN ('active','ramping','flagship')),
    health          text NOT NULL DEFAULT 'good' CHECK (health IN ('good','watch','critical')),
    manager_name    text NOT NULL DEFAULT '',
    payer_nhis      smallint NOT NULL DEFAULT 0 CHECK (payer_nhis BETWEEN 0 AND 100),
    payer_cash_momo smallint NOT NULL DEFAULT 0 CHECK (payer_cash_momo BETWEEN 0 AND 100),
    payer_private   smallint NOT NULL DEFAULT 0 CHECK (payer_private BETWEEN 0 AND 100),
    latitude        double precision NOT NULL DEFAULT 0,
    longitude       double precision NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT payer_mix_sums_100 CHECK (payer_nhis + payer_cash_momo + payer_private = 100)
);

CREATE TABLE facility_metrics (
    facility_id           text NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    metric_date           date NOT NULL,
    revenue               bigint NOT NULL DEFAULT 0,
    cash_revenue          bigint NOT NULL DEFAULT 0,
    momo_revenue          bigint NOT NULL DEFAULT 0,
    patients_seen         integer NOT NULL DEFAULT 0,
    admissions            integer NOT NULL DEFAULT 0,
    occupancy_rate        double precision NOT NULL DEFAULT 0,
    avg_wait_minutes      integer NOT NULL DEFAULT 0,
    nhis_claims_submitted integer NOT NULL DEFAULT 0,
    nhis_claims_paid      integer NOT NULL DEFAULT 0,
    nhis_claims_denied    integer NOT NULL DEFAULT 0,
    nhis_outstanding      bigint NOT NULL DEFAULT 0,
    unbilled_amount       bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (facility_id, metric_date)
);
CREATE INDEX idx_facility_metrics_facility_date ON facility_metrics (facility_id, metric_date DESC);

CREATE TABLE inventory_items (
    id             text PRIMARY KEY,
    facility_id    text NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    name           text NOT NULL,
    category       text NOT NULL DEFAULT '',
    stock_level    integer NOT NULL DEFAULT 0 CHECK (stock_level >= 0),
    daily_burn     double precision NOT NULL DEFAULT 0 CHECK (daily_burn >= 0),
    reorder_point  integer NOT NULL DEFAULT 0,
    lead_time_days integer NOT NULL DEFAULT 0 CHECK (lead_time_days >= 0),
    unit_cost      bigint NOT NULL DEFAULT 0
);
CREATE INDEX idx_inventory_facility ON inventory_items (facility_id);

CREATE TABLE staff (
    id             text PRIMARY KEY,
    facility_id    text NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    name           text NOT NULL,
    role           text NOT NULL,
    licence_number text NOT NULL DEFAULT '',
    licence_expiry date,
    status         text NOT NULL DEFAULT 'active',
    attrition_risk double precision NOT NULL DEFAULT 0 CHECK (attrition_risk BETWEEN 0 AND 1),
    joined_date    date
);
CREATE INDEX idx_staff_facility ON staff (facility_id);

CREATE TABLE alerts (
    id                 text PRIMARY KEY,
    facility_id        text NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    type               text NOT NULL,
    severity           text NOT NULL CHECK (severity IN ('good','watch','critical')),
    title              text NOT NULL,
    detail             text NOT NULL DEFAULT '',
    supporting_figures jsonb NOT NULL DEFAULT '{}'::jsonb,
    status             text NOT NULL DEFAULT 'open' CHECK (status IN ('open','dismissed','resolved')),
    created_at         timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_alerts_facility_status ON alerts (facility_id, status);

CREATE TABLE tasks (
    id          text PRIMARY KEY,
    title       text NOT NULL,
    detail      text NOT NULL DEFAULT '',
    facility_id text REFERENCES facilities(id) ON DELETE SET NULL,
    priority    text NOT NULL CHECK (priority IN ('low','medium','high')),
    status      text NOT NULL CHECK (status IN ('todo','in_progress','done')),
    due_date    timestamptz,
    assigned_to text NOT NULL DEFAULT '',
    created_by  text NOT NULL DEFAULT '',
    source      text NOT NULL CHECK (source IN ('manual','brief','alert')),
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_tasks_status ON tasks (status);

CREATE TABLE approvals (
    id            text PRIMARY KEY,
    type          text NOT NULL CHECK (type IN ('capital','hire','reorder')),
    facility_id   text REFERENCES facilities(id) ON DELETE SET NULL,
    amount        bigint NOT NULL DEFAULT 0,
    title         text NOT NULL,
    context       text NOT NULL DEFAULT '',
    requested_by  text NOT NULL DEFAULT '',
    status        text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','declined')),
    decided_at    timestamptz,
    decision_note text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_approvals_status ON approvals (status);

CREATE TABLE briefs (
    id                text PRIMARY KEY,
    brief_date        date NOT NULL,
    prose             text NOT NULL DEFAULT '',
    items             jsonb NOT NULL DEFAULT '[]'::jsonb,
    generated_at      timestamptz NOT NULL DEFAULT now(),
    model             text NOT NULL DEFAULT '',
    source_signal_ids text[] NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_briefs_date ON briefs (brief_date DESC);

CREATE TABLE insights (
    id                 text PRIMARY KEY,
    type               text NOT NULL,
    facility_id        text REFERENCES facilities(id) ON DELETE CASCADE,
    content            text NOT NULL,
    supporting_figures jsonb NOT NULL DEFAULT '{}'::jsonb,
    generated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id          text PRIMARY KEY,
    name        text NOT NULL,
    role        text NOT NULL CHECK (role IN ('executive','facility_manager')),
    facility_id text REFERENCES facilities(id) ON DELETE SET NULL,
    preferences jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT manager_has_facility CHECK (
        (role = 'facility_manager' AND facility_id IS NOT NULL) OR
        (role = 'executive' AND facility_id IS NULL)
    )
);
