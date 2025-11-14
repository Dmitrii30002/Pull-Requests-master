CREATE TABLE IF NOT EXISTS pull_requests (
    id       varchar(255) PRIMARY KEY,
	name     varchar(255) NOT NULL,
	author_id varchar(255) NOT NULL,
	status   varchar(10) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	merged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_pull_requests_author
    FOREIGN KEY author_id
    REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_pull_requests_id ON pull_requests(id);