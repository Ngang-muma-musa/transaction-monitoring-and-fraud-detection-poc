-- 1. Remove Indexes first
DROP INDEX IF EXISTS idx_transactions_user_velocity;

-- 2. Drop Tables
DROP TABLE IF EXISTS transactions;

-- 3. Drop Custom Types (Enums)
-- Note: You cannot drop a type if it is still being used by a column, 
-- which is why we drop the table first.
DROP TYPE IF EXISTS transaction_status;