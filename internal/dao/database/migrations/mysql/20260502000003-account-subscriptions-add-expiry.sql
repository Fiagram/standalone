-- +migrate Up
ALTER TABLE account_subscriptions
    ADD COLUMN billing_period ENUM('free', 'monthly', 'yearly') NOT NULL DEFAULT 'free' AFTER plan,
    ADD COLUMN expires_at TIMESTAMP NULL DEFAULT NULL AFTER billing_period;

-- +migrate Down
ALTER TABLE account_subscriptions
    DROP COLUMN expires_at,
    DROP COLUMN billing_period;
