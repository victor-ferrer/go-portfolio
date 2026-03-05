CREATE TABLE IF NOT EXISTS stock_quotations (
    id VARCHAR(255) PRIMARY KEY,
    ticker VARCHAR(255) NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    volume DOUBLE PRECISION NULL
);

CREATE INDEX IF NOT EXISTS idx_stock_quotations_ticker ON stock_quotations(ticker);
