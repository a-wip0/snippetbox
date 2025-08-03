-- Create the `snippets` table
CREATE TABLE snippets (
    id      INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title   VARCHAR(100) NOT NULL,
    content TEXT         NOT NULL,
    created DATETIME     NOT NULL,
    expires DATETIME     NOT NULL
);

-- Create the `users` table
CREATE TABLE users (
    id              INTEGER      NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    hashed_password CHAR(60)     NOT NULL,
    created         DATETIME     NOT NULL
);

-- Add unique constraint for email
ALTER TABLE users
    ADD CONSTRAINT users_uc_email UNIQUE (email);

-- Index for sorting/filtering snippets by creation time
CREATE INDEX idx_snippets_created ON snippets (created);
