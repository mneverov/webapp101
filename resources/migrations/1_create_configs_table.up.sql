CREATE TABLE configs
(
    name              TEXT PRIMARY KEY,
    url               TEXT UNIQUE NOT NULL     DEFAULT '',
    scraping_interval TEXT        NOT NULL     DEFAULT '1m',
    deleted_at        TIMESTAMP WITH TIME ZONE DEFAULT NULL
);
