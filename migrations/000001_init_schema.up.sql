CREATE TABLE users (
    user_id     TEXT        PRIMARY KEY,
    username    TEXT        NOT NULL UNIQUE,                  
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE teams (
    team_name   TEXT        PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE team_members (
    team_name   TEXT NOT NULL REFERENCES teams(team_name)   ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(user_id)     ON DELETE CASCADE,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_name, user_id)
);

CREATE TABLE pull_requests (
    pull_request_id     TEXT        PRIMARY KEY, -- исправить
    pull_request_name   TEXT        NOT NULL,
    author_id           TEXT        NOT NULL REFERENCES users(user_id),
    -- author_team_name           TEXT NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at           TIMESTAMPTZ NULL
);

CREATE TABLE pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL REFERENCES users(user_id)         ON DELETE CASCADE,
    -- reviewer_team_name  TEXT NOT NULL,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_team_members_team_name ON team_members(team_name);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX idx_reviewers_pull_request_id ON pull_request_reviewers(pull_request_id);
CREATE INDEX idx_reviewers_user_id ON pull_request_reviewers(user_id);
