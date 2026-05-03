-- +migrate Up
CREATE TABLE IF NOT EXISTS subscription_plans (
    plan           ENUM('pro', 'max')                  NOT NULL,
    billing_period ENUM('monthly', 'yearly')           NOT NULL,
    price          DECIMAL(10, 2)                      NOT NULL,
    currency       CHAR(3)                             NOT NULL DEFAULT 'VND',
    created_at     TIMESTAMP                           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP                           NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (plan, billing_period)
);

INSERT INTO subscription_plans (plan, billing_period, price, currency) VALUES
    ('pro',  'monthly',  99000.00, 'VND'),
    ('pro',  'yearly',  990000.00, 'VND'),
    ('max',  'monthly', 199000.00, 'VND'),
    ('max',  'yearly', 1990000.00, 'VND');

-- +migrate Down
DROP TABLE IF EXISTS subscription_plans;
