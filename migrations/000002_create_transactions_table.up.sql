CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    prompt TEXT NOT NULL DEFAULT '',
    operation_type TEXT NOT NULL CHECK (operation_type IN ('credit', 'debit', 'free_usage')),
    amount BIGINT NOT NULL CHECK (
        (operation_type = 'credit' AND amount > 0) OR
        (operation_type = 'debit' AND amount < 0) OR
        (operation_type = 'free_usage' AND amount = 0)
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id_created_at ON transactions(user_id, created_at DESC);