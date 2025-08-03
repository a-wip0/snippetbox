-- Add user_id to snippets to track ownership
ALTER TABLE snippets
    ADD COLUMN user_id INTEGER NOT NULL,
    ADD CONSTRAINT fk_snippets_user_id
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE;

-- Insert seed user
INSERT INTO users (name, email, hashed_password, created)
VALUES (
    'Alice Jones',
    'alice@example.com',
    '$2a$12$NuTjWXm3KKntReFwyBVHyuf/to.HEwTy.eS206TNfkGfr6HzGJSWG',
    '2022-01-01 10:00:00'
);
