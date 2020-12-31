CREATE TABLE metrics
(
    id            SERIAL PRIMARY KEY,
    name          TEXT                     NOT NULL,
    status_code   INTEGER                  NOT NULL,
    response_size BIGINT                   NOT NULL,
    response_time INTEGER                  NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    FOREIGN KEY (name) REFERENCES configs (name)
);
