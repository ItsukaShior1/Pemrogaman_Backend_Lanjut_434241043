INSERT INTO users (username, email, password_hash, role)
VALUES (
    'budi01',
    'budi01@lab.id',
    '$2a$10$zJByUi46nnJgO/YPfX.NFujUWQyV3lYys.NkiB5nKHlMGdpNS5E9i',
    'user'
)
ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, email, password_hash, role)
VALUES (
    'admin01',
    'admin01@lab.id',
    '$2a$10$cKlzEyfbKXJwJN3gTOdRYuC2EExr71ueZtWkz4KoaJV/v3wWLmFhO',
    'admin'
)
ON CONFLICT (username) DO NOTHING;

SELECT id, username, email, role, substring(password_hash, 1, 7) AS hash_prefix, length(password_hash) AS hash_len
FROM users
ORDER BY id;
