CREATE TABLE IF NOT EXISTS users (
    id varchar(255) PRIMARY KEY,
    username varchar(255) NOT NULL,
    is_active boolean NOT NULL,
    team_name varchar(255) NOT NULL,
    CONSTRAINT fk_users_teams 
    FOREIGN KEY (team_name) 
    REFERENCES teams(name) ON DELETE CASCADE
);

CREATE INDEX idx_users_id ON users(id);
CREATE INDEX idx_users_username ON users(username);

