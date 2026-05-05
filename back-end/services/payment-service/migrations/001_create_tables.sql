-- Payment Service Database Schema

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    method VARCHAR(20) NOT NULL,       -- VA, E_WALLET, CARD, COD
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',  -- PENDING, PAID, FAILED, REFUNDED
    payment_url TEXT,
    external_id VARCHAR(100),
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    paid_at TIMESTAMP,
    failed_at TIMESTAMP,
    refunded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_order_id ON transactions(order_id);
CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);
CREATE INDEX idx_transactions_status ON transactions(status);
