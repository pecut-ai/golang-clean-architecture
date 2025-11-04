CREATE TABLE addresses
(
    id          VARCHAR(100) NOT NULL,
    contact_id  VARCHAR(100) NOT NULL,
    street      VARCHAR(255),
    city        VARCHAR(255),
    province    VARCHAR(255),
    postal_code VARCHAR(10),
    country     VARCHAR(100),
    created_at  BIGINT       NOT NULL,
    updated_at  BIGINT       NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_addresses_contact_id FOREIGN KEY (contact_id) REFERENCES contacts (id)
);