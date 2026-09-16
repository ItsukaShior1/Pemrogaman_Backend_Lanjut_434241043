CREATE TABLE IF NOT EXISTS prestasi (
    id_prestasi         SERIAL       PRIMARY KEY,
    student_id INT          NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    champion   INT        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO prestasi (student_id, title, champion)
VALUES (1, 'Juara 1 Lomba Pemrograman', 1);