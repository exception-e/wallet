
CREATE TABLE IF NOT EXISTS wallets (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    balance bigint NOT NULL DEFAULT 0 CHECK ( balance >=0 ),
    currency varchar DEFAULT 'RUB',
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now());


CREATE TABLE IF NOT EXISTS operations (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    idempotency_key varchar UNIQUE,
    wallet_id uuid NOT NULL REFERENCES wallets(id),
    type            text        NOT NULL CHECK (type IN ('DEPOSIT','WITHDRAW')),
    amount bigint NOT NULL CHECK ( amount >= 0 ),
    currency varchar DEFAULT 'RUB',
    status varchar NOT NULL CHECK ( status IN ('APPLIED', 'REJECTED') ),
    RejectionCode  text CHECK (type IN ('INSUFFICIENT_FUNDS', 'BALANCE_OVERFLOW')),
    created_at timestamptz NOT NULL DEFAULT now());


