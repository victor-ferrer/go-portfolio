CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    broker VARCHAR(255) NOT NULL,
    imported_at TIMESTAMP NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    uniqueness_key VARCHAR(64) NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_aggregate_id ON events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_broker ON events(broker);
CREATE INDEX IF NOT EXISTS idx_created_at ON events(created_at);
