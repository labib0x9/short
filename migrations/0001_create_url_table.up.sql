CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE urls (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url        TEXT         NOT NULL,
    short      VARCHAR(20)  NOT NULL UNIQUE,
    created_at    TIMESTAMP     DEFAULT NOW(),
    expire_at  TIMESTAMP
);