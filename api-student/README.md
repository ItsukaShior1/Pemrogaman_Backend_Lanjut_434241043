# api-student

REST API untuk mengelola data mahasiswa, dibangun dengan
**Go + Fiber v2 + PostgreSQL** sebagai lanjutan dari tugas
pertemuan 1–3.

Proyek ini direstrukturisasi mengikuti pola **Clean Architecture**
pada pertemuan 4: layer terpisah dengan tanggung jawab masing-masing,
business rules yang dapat diuji tanpa HTTP, dan logger JSON ke file
dengan rotasi.

---

## Struktur Folder

```
api-student/
├── app/
│   ├── model/                # entitas Student + DTO + amplop respons
│   │   ├── student.go
│   │   ├── portfolio.go      # portofolio (MongoDB), bebas tipe driver
│   │   └── review.go         # ulasan kelas (MongoDB), bebas tipe driver
│   ├── repository/           # akses data ke PostgreSQL + MongoDB
│   │   ├── student_repository.go
│   │   ├── portfolio_repository.go   # index, $push/$pull, aggregation
│   │   └── review_repository.go
│   └── service/              # orchestration HTTP + business rules
│       ├── student_service.go   # handler fiber.Ctx (terima request)
│       ├── student_rules.go     # business rules MURNI (no Fiber)
│       ├── student_rules_test.go # unit test untuk rules
│       ├── portfolio_service.go
│       └── portfolio_rules.go
├── config/
│   ├── app.go                # struct AppConfig
│   ├── env.go                # loader .env
│   └── logger.go             # logger JSON + rotasi file
├── database/
│   ├── postgres.go           # connection pool pgxpool
│   └── mongo.go              # client MongoDB + ping saat start
├── helper/
│   └── response.go           # amplop respons + parser query string
├── logs/                     # output log (tidak di-commit)
├── middleware/
│   └── require_json.go       # RequireJSON global
├── route/
│   └── route.go              # pendaftaran route
├── migrations/
│   └── 001_create_students.sql
├── .env                      # tidak di-commit
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── main.go                   # hanya urutan perakitan
└── README.md
```

---

## Pemetaan ke Layer Clean Architecture

| Folder / File | Layer CA | Tanggung Jawab |
|---|---|---|
| `app/model/` | Entities | Struct Student, DTO request/response, amplop WebResponse |
| `app/repository/` | Interface Adapters (Gateway) | Query ke PostgreSQL lewat pgxpool + query ke MongoDB lewat driver v2 |
| `app/service/student_rules.go` | Enterprise Business Rules | Validasi murni, ApplyPatch, tanpa Fiber |
| `app/service/student_service.go` | Application Business Rules | Orchestration: parse request → rules → repository → response |
| `helper/response.go` | Interface Adapters (Presenter) | Amplop respons + parser query string |
| `middleware/require_json.go` | Interface Adapters | Middleware global RequireJSON |
| `route/route.go` | Interface Adapters | Pendaftaran route + middleware Fiber |
| `config/` | Framework & Driver | env loader, logger, AppConfig |
| `database/` | Framework & Driver | pgxpool ke PostgreSQL + mongo client |
| `main.go` | Composition Root | Urutan perakitan, tidak ada handler |

---

## Business Rules yang Diuji (Tanpa HTTP)

Tiga fungsi inti di `app/service/student_rules.go`:

1. `ValidateCreate(req model.CreateStudentRequest) map[string]string`
   — validasi untuk POST /students.
2. `ValidateReplace(req model.ReplaceStudentRequest) map[string]string`
   — validasi untuk PUT /students/:id.
3. `ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string)`
   — penerapan perubahan untuk PATCH /students/:id.

File rules ini **tidak mengimpor Fiber sama sekali**. Ia hanya butuh
`strings` (stdlib) dan `api-student/app/model`. Konsekuensinya,
10 unit test pada `student_rules_test.go` berjalan tanpa server HTTP
dan tanpa database — selesai dalam < 2 detik.

```bash
go test ./app/service/... -v
```

---

## Logger

`config/logger.go` membuka `logs/app.log` dan menggabungkannya dengan
STDOUT (kecuali `APP_MODE=release`). Setiap request HTTP dicatat satu
baris JSON berisi `time`, `request_id`, `method`, `path`, `status`,
`duration_ms`, dan `ip`.

Rotasi: ketika file > 10 MB, file diputar menjadi
`logs/app.log.<unix>` dan hanya 5 backup terbaru yang disimpan.

Folder `logs/` tercantum di `.gitignore` sehingga tidak masuk git.

---

## Dua Database: PostgreSQL + MongoDB

Sejak Pertemuan 10 aplikasi ini memakai **polyglot persistence**: data yang
berstruktur tetap dan butuh constraint tetap di PostgreSQL, data yang bentuknya
bervariasi dan dibaca utuh disimpan di MongoDB.

| Data | Database | Repository |
| ---- | -------- | ---------- |
| Student, user, role, kelas, KRS | PostgreSQL | `app/repository/*_repository.go` |
| Portofolio mahasiswa | MongoDB `student_portfolios` | `app/repository/portfolio_repository.go` |
| Ulasan kelas | MongoDB `course_reviews` | `app/repository/review_repository.go` |

Aturan yang dipegang:

- Seluruh akses ke MongoDB melewati repository. Package `app/model` bebas dari
  tipe milik driver — `bson.ObjectID` hanya muncul di repository, di model
  id disimpan sebagai string.
- Index dibuat sekali saat aplikasi menyala lewat `EnsureIndexes`.
- Operation update yang punya batas (20 proyek, 50 balasan) memakai filter
  bersyarat di dalam operasi yang sama, bukan dibaca dulu lalu ditulis.
- `GET /api/v1/health` memeriksa **kedua** database; pesan kegagalannya
  menyebut yang mana yang mati.

Portofolio dan endpoint-nya dijelaskan di
[`docs/rancangan-portofolio.md`](docs/rancangan-portofolio.md).

---

## Menjalankan

```bash
cd "C:/Users/LENOVO/OneDrive/semester 5/Prak/Backend Prak/tugas01/api-student"

# Pastikan schema sudah ada
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
psql -U postgres -d praktikum_backend -f migrations/008_create_courses.sql
psql -U postgres -d praktikum_backend -f migrations/009_review_permission.sql

# Build
go build ./...

# Test
go test ./app/service/... -v

# Run
go run .
```

Server jalan di `http://localhost:3000`. MongoDB harus menyala juga; kalau
tidak, aplikasi menolak menyala saat start.

---

## Kontrak API Singkat

| Method | Path                                       | Keterangan                  |
| ------ | ------------------------------------------ | --------------------------- |
| GET    | `/`                                        | Hello world                 |
| GET    | `/api/v1/health`                           | Cek PostgreSQL + MongoDB   |
| GET    | `/api/v1/students`                         | Daftar (paginasi + filter)  |
| GET    | `/api/v1/students/:id`                     | Ambil satu                  |
| POST   | `/api/v1/students`                         | Buat baru                   |
| PUT    | `/api/v1/students/:id`                     | Ganti seluruh               |
| PATCH  | `/api/v1/students/:id`                     | Ubah sebagian               |
| DELETE | `/api/v1/students/:id`                     | Hapus                       |
| GET    | `/api/v1/students/:id/portfolio`           | Ambil portofolio (MongoDB)  |
| PUT    | `/api/v1/students/:id/portfolio`           | Buat / ganti portofolio     |
| POST   | `/api/v1/students/:id/portfolio/projects`  | Tambah proyek (maks 20)     |
| DELETE | `/api/v1/students/:id/portfolio/projects/:projectId` | Hapus proyek       |
| GET    | `/api/v1/portfolios/stats/technologies`    | 10 teknologi terpopuler      |
| GET    | `/api/v1/courses`                          | Daftar kelas                |
| GET    | `/api/v1/courses/:id`                      | Ambil satu kelas            |
| GET    | `/api/v1/courses/:id/reviews`              | Daftar ulasan satu kelas    |
| GET    | `/api/v1/courses/:id/reviews/summary`      | Ringkasan ulasan            |
| POST   | `/api/v1/courses/:id/reviews`              | Tulis ulasan                |
| GET    | `/api/v1/reviews/:id`                      | Ambil ulasan                |
| PATCH  | `/api/v1/reviews/:id`                      | Ubah ulasan sendiri         |
| DELETE | `/api/v1/reviews/:id`                      | Hapus ulasan                |
| POST   | `/api/v1/reviews/:id/replies`              | Tambah balasan (maks 50)    |
| POST   | `/api/v1/reviews/:id/helpful`              | Tandai membantu             |

Detail kontrak per endpoint (status, body, error) ada di laporan
PDF pertemuan 4.

---

## Bukti Portofolio

Skrip `scripts/bukti-portofolio.sh` menjalankan seluruh perintah curl yang
dibutuhkan laporan (termasuk uji 10 permintaan bentrok untuk batas 20 proyek):

```bash
bash scripts/bukti-portofolio.sh
```
