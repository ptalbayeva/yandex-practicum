CREATE TABLE shorten_urls (
                        id SERIAL PRIMARY KEY,
                        uuid VARCHAR(50) NOT NULL,
                        original VARCHAR(255) NOT NULL,
                        shorten VARCHAR(255) NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_shorten_urls_original ON shorten_urls(original);

CREATE INDEX idx_shorten_urs_shorten ON shorten_urls(shorten);