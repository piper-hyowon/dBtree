CREATE TABLE IF NOT EXISTS port_allocations
(
    instance_id  VARCHAR(36) PRIMARY KEY,
    port         INTEGER NOT NULL UNIQUE CHECK (port >= 30000 AND port <= 31999),
    allocated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_port_allocations_port ON port_allocations (port);