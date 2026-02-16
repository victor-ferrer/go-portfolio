CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    type TEXT NOT NULL,
    broker TEXT NOT NULL,
    imported_at TIMESTAMP NOT NULL,
    payload TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    uniqueness_key TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_aggregate_id ON events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_broker ON events(broker);
CREATE INDEX IF NOT EXISTS idx_created_at ON events(created_at);
