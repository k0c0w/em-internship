CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    service_name TEXT NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    is_deleted BIT NOT NULL DEFAULT 0::BIT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP
);
