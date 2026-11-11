CREATE TABLE todoapp.task_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    task_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    schema_version SMALLINT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    lease_until TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    last_error TEXT
);

CREATE INDEX task_outbox_pending_idx
    ON todoapp.task_outbox (available_at, id)
    WHERE published_at IS NULL;

CREATE TABLE todoapp.task_events (
    event_id UUID PRIMARY KEY,
    task_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    schema_version SMALLINT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX task_events_task_id_idx
    ON todoapp.task_events (task_id, occurred_at);