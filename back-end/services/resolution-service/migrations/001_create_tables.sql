-- Resolution Service Database Schema

CREATE TABLE IF NOT EXISTS tickets (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    awb VARCHAR(20),
    user_id VARCHAR(36),
    type VARCHAR(30) NOT NULL,      -- LOST, DAMAGED, DELIVERY_FAILED, COMPLAINT
    priority VARCHAR(20) NOT NULL DEFAULT 'LOW',  -- LOW, MEDIUM, HIGH, CRITICAL
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',   -- OPEN, IN_PROGRESS, RESOLVED, CLOSED
    description TEXT,
    resolution VARCHAR(30),  -- REFUND, RESEND, RETURN, CLOSED
    agent_id VARCHAR(36),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tickets_order_id ON tickets(order_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_type ON tickets(type);

CREATE TABLE IF NOT EXISTS claims (
    id VARCHAR(36) PRIMARY KEY,
    ticket_id VARCHAR(36) NOT NULL REFERENCES tickets(id),
    order_id VARCHAR(36) NOT NULL,
    claim_type VARCHAR(20) NOT NULL,  -- INSURANCE, REFUND
    amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',  -- PENDING, APPROVED, REJECTED, PAID
    approved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claims_ticket_id ON claims(ticket_id);
