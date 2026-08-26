# api-students

REST API sederhana untuk mengelola data mahasiswa, dibangun dengan
**Go + Fiber v2** sebagai lanjutan dari tugas pertemuan 1
(struct `Student` naik pangkat dari sekadar tipe data menjadi entitas REST API).

> Data disimpan di memori. Setiap kali server dimatikan, seluruh data
> hilang. Penyimpanan persisten (basis data) baru dibahas pada pertemuan 3.

---

## Daftar Isi

- [Menjalankan](#menjalankan)
- [Kontrak API per Endpoint](#kontrak-api-per-endpoint)
  - [1. `GET /api/v1/students` — Daftar](#1-get-apiv1students--daftar)
  - [2. `GET /api/v1/students/:id` — Satu Mahasiswa](#2-get-apiv1studentsid--satu-mahasiswa)
  - [3. `POST /api/v1/students` — Buat Baru](#3-post-apiv1students--buat-baru)
  - [4. `PUT /api/v1/students/:id` — Ganti Seluruh](#4-put-apiv1studentsid--ganti-seluruh)
  - [5. `PATCH /api/v1/students/:id` — Ubah Sebagian](#5-patch-apiv1studentsid--ubah-sebagian)
  - [6. `DELETE /api/v1/students/:id` — Hapus](#6-delete-apiv1studentsid--hapus)
- [Query String](#query-string)
- [Bentuk Respons](#bentuk-respons)
- [Status HTTP yang Dipakai](#status-http-yang-dipakai)
- [Struktur Proyek](#struktur-proyek)

---

## Menjalankan

```bash
cd ../api-student
go mod tidy
go run .
```

Server berjalan di `http://localhost:3000`. Coba:

```bash
curl -i http://localhost:3000/api/v1/health
```

---

## Kontrak API per Endpoint

Tiap endpoint di bawah punya kolom: **metode + path**, **parameter**,
**contoh body permintaan**, **status yang mungkin dikembalikan**, dan
**contoh respons**. Dokumentasi ini adalah sumber kebenaran bagi klien
— Swagger UI pada pertemuan 13 akan menggenerasi tabel yang sama
secara otomatis dari anotasi kode.

---

### 1. `GET /api/v1/students` — Daftar

Mengembalikan daftar mahasiswa dengan paginasi, pencarian,
pengurutan, dan penyaringan.

| Kolom              | Nilai                                                                              |
| ------------------ | ---------------------------------------------------------------------------------- |
| **Metode + Path**  | `GET /api/v1/students`                                                             |
| **Parameter**      | Query string (lihat [Query String](#query-string))                                  |
| **Body request**   | —                                                                                  |
| **Status sukses**  | `200 OK`                                                                           |
| **Status gagal**   | — (tidak ada skenario gagal pada daftar)                                            |

**Contoh permintaan:**

```
GET /api/v1/students?page=1&limit=2&sort=grade&order=desc HTTP/1.1
Host: localhost:3000
```

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "daftar student berhasil diambil",
  "data": [
    {
      "id": 3,
      "nim": "2201110003",
      "name": "Dewi Anggraini",
      "grade": 88.75,
      "is_active": false,
      "created_at": "2026-08-26T19:15:30+07:00"
    },
    {
      "id": 1,
      "nim": "2201110001",
      "name": "Budi S. Edited",
      "grade": 85.5,
      "is_active": true,
      "created_at": "2026-08-26T19:16:02+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 2,
    "total": 3,
    "total_pages": 2
  }
}
```

---

### 2. `GET /api/v1/students/:id` — Satu Mahasiswa

Mengembalikan satu mahasiswa berdasarkan ID, atau 404 bila tidak ada.

| Kolom              | Nilai                                                          |
| ------------------ | -------------------------------------------------------------- |
| **Metode + Path**  | `GET /api/v1/students/:id`                                     |
| **Parameter**      | `:id` — integer positif (path parameter)                       |
| **Body request**   | —                                                              |
| **Status sukses**  | `200 OK`                                                       |
| **Status gagal**   | `400 Bad Request`, `404 Not Found`                             |

**Contoh permintaan:**

```
GET /api/v1/students/1 HTTP/1.1
Host: localhost:3000
```

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student ditemukan",
  "data": {
    "id": 1,
    "nim": "2201110001",
    "name": "Budi Santoso",
    "grade": 80,
    "is_active": true,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

**Contoh respons 400** (`:id` bukan angka):

```json
{
  "success": false,
  "message": "id harus berupa angka positif"
}
```

**Contoh respons 404** (tidak ada mahasiswa dengan id tersebut):

```json
{
  "success": false,
  "message": "student tidak ditemukan"
}
```

---

### 3. `POST /api/v1/students` — Buat Baru

Membuat mahasiswa baru. Wajib mengirim semua field. Server mengembalikan
**201 Created** disertai header `Location` yang menunjuk ke alamat sumber daya baru.

| Kolom              | Nilai                                                          |
| ------------------ | -------------------------------------------------------------- |
| **Metode + Path**  | `POST /api/v1/students`                                        |
| **Parameter**      | —                                                              |
| **Body request**   | `Content-Type: application/json`, semua field wajib            |
| **Status sukses**  | `201 Created` (+ header `Location`)                            |
| **Status gagal**   | `400 Bad Request`, `409 Conflict`, `415 Unsupported Media Type`, `422 Unprocessable Entity` |

**Contoh body permintaan:**

```json
{
  "nim": "2201110001",
  "name": "Budi Santoso",
  "grade": 80.0,
  "is_active": true
}
```

**Contoh respons 201:**

```http
HTTP/1.1 201 Created
Location: /api/v1/students/1
Content-Type: application/json

{
  "success": true,
  "message": "student berhasil dibuat",
  "data": {
    "id": 1,
    "nim": "2201110001",
    "name": "Budi Santoso",
    "grade": 80,
    "is_active": true,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

**Contoh respons 409** (NIM duplikat):

```json
{
  "success": false,
  "message": "NIM sudah dipakai"
}
```

**Contoh respons 422** (validasi gagal):

```json
{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "nim": "wajib diisi",
    "grade": "harus bernilai antara 0 dan 100"
  }
}
```

---

### 4. `PUT /api/v1/students/:id` — Ganti Seluruh

Mengganti **seluruh isi** mahasiswa. Karena PUT bermakna replace, semua
field wajib dikirim — field yang tidak dikirim dianggap kosong (zero value).

| Kolom              | Nilai                                                          |
| ------------------ | -------------------------------------------------------------- |
| **Metode + Path**  | `PUT /api/v1/students/:id`                                     |
| **Parameter**      | `:id` — integer positif                                        |
| **Body request**   | `Content-Type: application/json`, semua field wajib            |
| **Status sukses**  | `200 OK`                                                       |
| **Status gagal**   | `400 Bad Request`, `404 Not Found`, `409 Conflict`, `415 Unsupported Media Type`, `422 Unprocessable Entity` |

**Contoh body permintaan** (ganti seluruh):

```json
{
  "nim": "2201110001",
  "name": "Budi S. Edited",
  "grade": 90.0,
  "is_active": false
}
```

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student berhasil diganti seluruhnya",
  "data": {
    "id": 1,
    "nim": "2201110001",
    "name": "Budi S. Edited",
    "grade": 90,
    "is_active": false,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

**Contoh respons 422** (field tidak lengkap):

```json
{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "nim": "wajib diisi pada PUT"
  }
}
```

---

### 5. `PATCH /api/v1/students/:id` — Ubah Sebagian

Mengubah **sebagian isi** mahasiswa. Hanya field yang dikirim yang berubah;
field yang tidak dikirim dibiarkan apa adanya. Karena alasan ini, semua
field pada struct `PatchStudentRequest` bertipe pointer (`*string`,
`*float64`, `*bool`) — supaya server bisa membedakan "tidak dikirim"
(`nil`) dari "dikirim bernilai default" (`false`, `0`, dst).

| Kolom              | Nilai                                                          |
| ------------------ | -------------------------------------------------------------- |
| **Metode + Path**  | `PATCH /api/v1/students/:id`                                   |
| **Parameter**      | `:id` — integer positif                                        |
| **Body request**   | `Content-Type: application/json`, minimal 1 field              |
| **Status sukses**  | `200 OK`                                                       |
| **Status gagal**   | `400 Bad Request`, `404 Not Found`, `409 Conflict`, `415 Unsupported Media Type`, `422 Unprocessable Entity` |

**Contoh body permintaan** (ubah hanya `grade`):

```json
{ "grade": 95.5 }
```

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student berhasil diperbarui sebagian",
  "data": {
    "id": 1,
    "nim": "2201110001",
    "name": "Budi S. Edited",
    "grade": 95.5,
    "is_active": false,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

Field `nim`, `name`, dan `is_active` **tidak berubah** meskipun tidak dikirim.

**Contoh respons 400** (body kosong):

```json
{
  "success": false,
  "message": "tidak ada field yang diubah"
}
```

---

### 6. `DELETE /api/v1/students/:id` — Hapus

Menghapus mahasiswa. Respons **tidak berisi body** — hanya header.

| Kolom              | Nilai                                                |
| ------------------ | ---------------------------------------------------- |
| **Metode + Path**  | `DELETE /api/v1/students/:id`                        |
| **Parameter**      | `:id` — integer positif                              |
| **Body request**   | —                                                    |
| **Status sukses**  | `204 No Content`                                     |
| **Status gagal**   | `400 Bad Request`, `404 Not Found`                   |

**Contoh permintaan:**

```
DELETE /api/v1/students/2 HTTP/1.1
Host: localhost:3000
```

**Contoh respons 204:**

```http
HTTP/1.1 204 No Content
X-Request-Id: ...
```

(Tidak ada body.)

---

## Query String

Berlaku untuk endpoint **`GET /api/v1/students`** (daftar).

| Parameter   | Tipe    | Wajib | Nilai bawaan | Keterangan                                                                  |
| ----------- | ------- | ----- | ------------ | --------------------------------------------------------------------------- |
| `page`      | integer | tidak | `1`          | Halaman ke berapa. Minimal 1 (di bawah 1 dipaksa jadi 1).                   |
| `limit`     | integer | tidak | `10`         | Baris per halaman. Batas atas **100** (lihat catatan di bawah).             |
| `search`    | string  | tidak | (kosong)     | Pencarian pada field `name`, case-insensitive.                              |
| `sort`      | string  | tidak | `id`         | Hanya field pada **whitelist**: `id`, `nim`, `name`, `grade`, `created_at`. |
| `order`     | string  | tidak | `asc`        | `asc` atau `desc`. Nilai lain diabaikan.                                   |
| `is_active` | boolean | tidak | (kosong)     | Filter `true` / `false`. Kosong = tidak menyaring.                          |
| `grade_min` | float   | tidak | (kosong)     | `grade >= grade_min` (inklusif).                                            |
| `grade_max` | float   | tidak | (kosong)     | `grade <= grade_max` (inklusif).                                            |

**Catatan batas atas `limit` = 100.** Klien dapat mengirim `?limit=99999999`
untuk membuat server kewalahan; server memotong nilai tersebut ke100.

**Catatan daftar putih `sort`.** Menerima nama field apa pun dari klien
dan menempelkannya ke query basis data adalah jalan termudah menuju
SQL injection. Saat ini data di memori, sehingga belum berbahaya —
tetapi ketika dipindahkan ke PostgreSQL pada pertemuan 3, whitelist
menjadi pengaman utama. Saat ini input di luar whitelist diabaikan
dan query di-default-kan ke `id`.

---

## Bentuk Respons

Seluruh endpoint memakai satu amplop baku, baik untuk respons berhasil
maupun gagal. Konsistensi bentuk membuat klien tidak perlu menebak
struktur per endpoint.

| Tipe respons      | Bentuk                                                                       |
| ----------------- | ---------------------------------------------------------------------------- |
| Tunggal / object  | `{ "success": true, "message": "...", "data": { ... } }`                     |
| Daftar            | `{ "success": true, "message": "...", "data": [ ... ], "meta": { ... } }`    |
| Gagal umum        | `{ "success": false, "message": "..." }`                                    |
| Gagal validasi    | `{ "success": false, "message": "validasi gagal", "errors": { "field": "..." } }` |

---

## Status HTTP yang Dipakai

| Status   | Nama                       | Situasi                                                                                  |
| -------- | -------------------------- | ---------------------------------------------------------------------------------------- |
| `200 OK` | Berhasil                   | GET / daftar, GET /:id, PUT, PATCH berhasil                                              |
| `201 Created` | Sumber daya dibuat    | POST berhasil — header `Location` ikut menunjuk ke alamat resource baru                  |
| `204 No Content` | Berhasil tanpa body | DELETE berhasil — tidak ada body yang perlu dikembalikan                                  |
| `400 Bad Request` | Permintaan salah    | `:id` bukan angka, atau body bukan JSON yang valid                                       |
| `404 Not Found` | Tidak ditemukan       | `:id` tidak ada, atau endpoint tidak dikenal                                              |
| `409 Conflict` | Konflik data         | NIM yang dipakai sudah dipakai mahasiswa lain                                            |
| `415 Unsupported Media Type` | Format tidak didukung | `Content-Type` bukan `application/json` (middleware `requireJSON` menolak)               |
| `422 Unprocessable Entity` | Validasi gagal      | Isi permintaan dipahami tapi tidak lolos validasi (mis. `grade` di luar 0–100) — dengan rincian per field |

---

## Struktur Proyek

```
api-students/
├── go.mod              # module api-students + dependensi Fiber v2
├── go.sum              # checksum dependensi
├── main.go             # bootstrap Fiber + middleware + route
├── model.go            # struct entitas, DTO request, amplop respons
├── helper.go           # helper respons + parser query string + requireJSON
├── handler.go          # handler tiap endpoint
├── README.md           # kontrak API (file ini)
└── screenshots/        # tangkapan layar pengujian (diisi manual)
```

Keempat berkas Go tetap berada dalam `package main` sehingga saling
melihat tanpa import tambahan. Pemisahan tanggung jawab yang lebih
ketat (handler → service → repository) baru dibahas pada pertemuan 3.

---

## Sumber Bantuan

Bagian yang dibantu AI: penulisan struktur awal `model.go`, `helper.go`,
`handler.go`, dan `main.go` mengikuti pola dari modul Pertemuan 2;
seluruh perilaku, validasi, dan endpoint dijelaskan di atas ditulis
sendiri dan diuji secara manual dengan `curl`.