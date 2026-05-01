-- +migrate Up
CREATE TABLE IF NOT EXISTS account_subscriptions (
    of_account_id BIGINT UNSIGNED,
    plan          ENUM('free', 'pro', 'max') NOT NULL DEFAULT 'free',
    status        ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (of_account_id),
    CONSTRAINT fk_account_subscriptions_of_account_id FOREIGN KEY (of_account_id) REFERENCES accounts(id)
);

-- +migrate Down
DROP TABLE IF EXISTS account_subscriptions;
