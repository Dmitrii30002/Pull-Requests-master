CREATE TABLE IF NOT EXISTS pr_reviewrs (
    user_id VARCHAR(255) NOT NULL,
    pr_id VARCHAR(255) NOT NULL,

    PRIMARY KEY (user_id, pr_id),
    CONSTRAINT fk_pr_reviewrs_user
    FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_pr_reviewrs_pr
    FOREIGN KEY (pr_id)
    REFERENCES pull_requests(id) ON DELETE CASCADE
);