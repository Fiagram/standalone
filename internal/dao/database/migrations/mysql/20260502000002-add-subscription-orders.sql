-- +migrate Up
CREATE TABLE IF NOT EXISTS subscription_orders (
    id                   BIGINT UNSIGNED AUTO_INCREMENT,
    of_account_id        BIGINT UNSIGNED                                     NOT NULL,
    plan                 ENUM('pro', 'max')                                  NOT NULL,
    billing_period       ENUM('monthly', 'yearly')                           NOT NULL,
    amount               DECIMAL(10, 2)                                      NOT NULL,
    currency             CHAR(3)                                             NOT NULL DEFAULT 'VND',
    status               ENUM('pending', 'paid', 'expired', 'cancelled')    NOT NULL DEFAULT 'pending',
    reference_code       VARCHAR(50)                                         NOT NULL,
    sepay_transaction_id VARCHAR(100)                                        NULL,
    payment_expires_at   TIMESTAMP                                           NOT NULL,
    sub_start_at         TIMESTAMP                                           NULL,
    sub_expires_at       TIMESTAMP                                           NULL,
    created_at           TIMESTAMP                                           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP                                           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT udx_subscription_orders_reference_code UNIQUE (reference_code),
    CONSTRAINT fk_subscription_orders_of_account_id FOREIGN KEY (of_account_id) REFERENCES accounts(id),
    INDEX idx_subscription_orders_of_account_id (of_account_id)
);

-- +migrate Down
DROP TABLE IF EXISTS subscription_orders;
