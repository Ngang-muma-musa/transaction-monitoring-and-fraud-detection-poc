CREATE TYPE transaction_status AS ENUM ('PENDING', 'APPROVED', 'DECLINED', 'FLAGGED');

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status transaction_status DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for the "Behavioral/Velocity" layer we built earlier
CREATE INDEX idx_transactions_user_velocity ON transactions (user_id, created_at) WHERE status != 'DECLINED';