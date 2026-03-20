-- +migrate Up
CREATE TABLE IF NOT EXISTS account_role (
    id INT UNSIGNED,
    name VARCHAR(20) NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL,
    fullname VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    of_role_id INT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    FOREIGN KEY (of_role_id) REFERENCES account_role(id),
    UNIQUE (username),
    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS account_passwords (
    of_account_id BIGINT UNSIGNED,
    hashed_string VARCHAR(128) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (of_account_id),
    FOREIGN KEY (of_account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS chatbot_webhooks (
    id BIGINT UNSIGNED AUTO_INCREMENT,
    of_account_id BIGINT UNSIGNED,
    name VARCHAR(26) NOT NULL,
    url VARCHAR(500) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    FOREIGN KEY (of_account_id) REFERENCES accounts(id),
    UNIQUE (of_account_id, name)
);

INSERT INTO account_role (id, name) VALUES (0, 'none');
INSERT INTO account_role (id, name) VALUES (1, 'admin');
INSERT INTO account_role (id, name) VALUES (2, 'member');

-- +migrate Down
DROP TABLE IF EXISTS chatbot_webhooks;

DROP TABLE IF EXISTS account_passwords;

DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS account_role;
