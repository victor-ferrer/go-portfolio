-- Add price as a dedicated queryable column.
-- DEFAULT 0 matches the Go float64 zero value, ensuring backward compatibility
-- with existing rows where price was not previously tracked separately.
ALTER TABLE events ADD COLUMN IF NOT EXISTS price NUMERIC(20, 8) NOT NULL DEFAULT 0;
