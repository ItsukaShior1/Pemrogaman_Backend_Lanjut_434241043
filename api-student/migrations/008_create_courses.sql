-- Migration 008: kelas dan KRS
-- psql -U postgres -d praktikum_backend -f migrations/008_create_courses.sql
--
-- Data kelas tetap di PostgreSQL karena isinya berstruktur tetap dan perlu
-- constraint. Ulasan kelasnya nanti disimpan di MongoDB.
CREATE TABLE IF NOT EXISTS courses (
    id         SERIAL       PRIMARY KEY,
    code       VARCHAR(20)  NOT NULL,
    name       VARCHAR(150) NOT NULL,
    semester   VARCHAR(20)  NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS courses_code_key ON courses (code);

CREATE TABLE IF NOT EXISTS enrollments (
    course_id   INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (course_id, user_id)
);

CREATE INDEX IF NOT EXISTS enrollments_user_idx ON enrollments (user_id);

INSERT INTO courses (code, name, semester) VALUES
    ('IF101', 'Pemrograman Backend Lanjut', '2025/2026'),
    ('IF102', 'Basis Data', '2025/2026')
ON CONFLICT (code) DO NOTHING;

-- Peserta kelas IF101. User lain sengaja tidak 등록 agar ada yang mendapat
-- 403 NOT_ENROLLED saat mencoba menulis ulasan.
INSERT INTO enrollments (course_id, user_id)
SELECT c.id, u.id
FROM courses c
CROSS JOIN users u
WHERE c.code = 'IF101' AND u.username IN ('budi01', 'admin01')
ON CONFLICT DO NOTHING;

INSERT INTO enrollments (course_id, user_id)
SELECT c.id, u.id
FROM courses c
CROSS JOIN users u
WHERE c.code = 'IF102' AND u.username IN ('user01', 'user02')
ON CONFLICT DO NOTHING;
