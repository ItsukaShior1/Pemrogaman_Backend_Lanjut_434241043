# api-students

REST API sederhana untuk mengelola data mahasiswa, dibangun dengan
**Go + Fiber v2** sebagai lanjutan dari tugas pertemuan 1 (struct `Student`
naik pangkat dari sekadar tipe data menjadi entitas REST API).

> Data disimpan di memori. Setiap kali server dimatikan, seluruh data
> hilang. Penyimpanan persisten (basis data) baru dibahas pada pertemuan 3.

---

## Menjalankan

```bash
cd ../api-student
go mod tidy
go run .
```

Server berjalan di `http://localhost:3000`.

---

## Endpoint Ringkas

| Metode | Endpoint                  | Perilaku                                                  |
| ------ | ------------------------- | --------------------------------------------------------- |
| GET    | `/api/v1/students`        | Daftar mahasiswa dengan paginasi, pencarian, sortir, saringan |
| GET    | `/api/v1/students/:id`    | Satu mahasiswa, atau 404                                  |
| POST   | `/api/v1/students`        | Buat mahasiswa baru (201 + header `Location`)             |
| PUT    | `/api/v1/students/:id`    | Ganti seluruh isi (semua field wajib)                     |
| PATCH  | `/api/v1/students/:id`    | Ubah sebagian (hanya field yang dikirim)                  |
| DELETE | `/api/v1/students/:id`    | Hapus mahasiswa (204 tanpa body)                          |
| GET    | `/api/v1/health`          | Pemeriksaan cepat server hidup                            |
| GET    | `/`                       | `Hello, World!` (pembanding dari pertemuan 1)             |

---

## Kontrak API (Lengkap)

### 1. `GET /api/v1/students` — Daftar

**Query string (semua opsional, dengan nilai bawaan aman):**

| Parameter   | Nilai bawaan | Keterangan                                                  |
| ----------- | ------------ | ----------------------------------------------------------- |
| `page`      | `1`          | Halaman ke berapa, minimal 1                                |
| `limit`     | `10`         | Baris per halaman, dibatasi maksimum 100                    |
| `search`    | (kosong)     | Pencarian pada `name`, tidak membedakan huruf besar/kecil   |
| `sort`      | `id`         | Hanya `id`, `nim`, `name`, `grade`, `created_at`            |
| `order`     | `asc`        | `asc` atau `desc` (lainnya diabaikan)                       |
| `is_active` | (kosong)     | `true` / `false`; kosong berarti tidak menyaring            |
| `grade_min` | (kosong)     | Syarat `grade >= grade_min`                                 |
| `grade_max` | (kosong)     | Syarat `grade <= grade_max`                                 |

**Batas atas `limit` adalah 100.** Alasannya: mencegah satu permintaan
mengambil data ratusan ribu baris yang bisa menghabiskan memori server
dan memotong respons. Pada pertemuan 3 ketika data dipindahkan ke basis
data, batas ini menjadi penting karena query tanpa batas atas dapat
mengunci tabel untuk waktu lama.

**Status yang mungkin dikembalikan:** `200`.

**Contoh respons:**

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
      "created_at": "2026-08-26T19:15:30.36+07:00"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 1, "total_pages": 1 }
}
```

---

### 2. `GET /api/v1/students/:id` — Satu Mahasiswa

**Parameter path:** `id` (integer positif).

**Status yang mungkin dikembalikan:** `200`, `400`, `404`.

| Status | Situasi                                  |
| ------ | ---------------------------------------- |
| `200`  | Ditemukan                                |
| `400`  | `id` bukan angka positif                 |
| `404`  | Tidak ada mahasiswa dengan id tersebut   |

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student ditemukan",
  "data": {
    "id": 1, "nim": "2201110001", "name": "Budi Santoso",
    "grade": 80, "is_active": true,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

**Contoh respons 404:**

```json
{ "success": false, "message": "student tidak ditemukan" }
```

---

### 3. `POST /api/v1/students` — Buat Mahasiswa Baru

**Header wajib:** `Content-Type: application/json`.

**Body permintaan (semua field wajib):**

```json
{
  "nim": "2201110001",
  "name": "Budi Santoso",
  "grade": 80.0,
  "is_active": true
}
```

**Status yang mungkin dikembalikan:** `201`, `400`, `409`, `415`, `422`.

| Status | Situasi                                                              |
| ------ | -------------------------------------------------------------------- |
| `201`  | Berhasil dibuat. Respons disertai header `Location`                   |
| `400`  | Body bukan JSON yang valid                                           |
| `409`  | NIM sudah dipakai mahasiswa lain                                     |
| `415`  | `Content-Type` bukan `application/json`                              |
| `422`  | Validasi isi gagal (mis. `nim` kosong, `grade` di luar 0–100)        |

**Contoh respons 201:**

```http
HTTP/1.1 201 Created
Location: /api/v1/students/1
Content-Type: application/json

{
  "success": true,
  "message": "student berhasil dibuat",
  "data": {
    "id": 1, "nim": "2201110001", "name": "Budi Santoso",
    "grade": 80, "is_active": true,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

**Contoh respons 422:**

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

### 4. `PUT /api/v1/students/:id` — Ganti Seluruh Isi

**Header wajib:** `Content-Type: application/json`.

**Body permintaan (semua field WAJIB; yang tidak dikirim dianggap kosong):**

```json
{
  "nim": "2201110001",
  "name": "Budi S. Edited",
  "grade": 90.0,
  "is_active": false
}
```

**Status yang mungkin dikembalikan:** `200`, `400`, `404`, `409`, `415`, `422`.

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student berhasil diganti seluruhnya",
  "data": {
    "id": 1, "nim": "2201110001", "name": "Budi S. Edited",
    "grade": 90, "is_active": false,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

> Perilaku PUT yang sesungguhnya: field yang tidak dikirim akan dianggap
> dikosongkan (zero value struct). Itulah mengapa semua field bertipe
> biasa dan wajib divalidasi.

---

### 5. `PATCH /api/v1/students/:id` — Ubah Sebagian

**Header wajib:** `Content-Type: application/json`.

**Body permintaan (hanya field yang ingin diubah):**

```json
{ "grade": 95.5 }
```

**Status yang mungkin dikembalikan:** `200`, `400`, `404`, `409`, `415`, `422`.

**Contoh respons 200:**

```json
{
  "success": true,
  "message": "student berhasil diperbarui sebagian",
  "data": {
    "id": 1, "nim": "2201110001", "name": "Budi S. Edited",
    "grade": 95.5, "is_active": false,
    "created_at": "2026-08-26T19:16:02+07:00"
  }
}
```

> Field `nim`, `name`, dan `is_active` pada contoh di atas tidak ikut
> berubah meskipun pemanggil tidak mengirimkannya. Inilah perbedaan
> perilaku yang nyata antara PUT dan PATCH.

---

### 6. `DELETE /api/v1/students/:id` — Hapus Mahasiswa

**Status yang mungkin dikembalikan:** `204`, `400`, `404`.

| Status | Situasi                              |
| ------ | ------------------------------------ |
| `204`  | Berhasil dihapus (respons tanpa body)|
| `400`  | `id` bukan angka positif             |
| `404`  | Tidak ada mahasiswa dengan id tersebut|

**Contoh respons 204:**

```http
HTTP/1.1 204 No Content
X-Request-Id: ...
```

(Tidak ada body.)

---

## Bentuk Respons Konsisten

Seluruh respons mengikuti satu amplop baku:

```json
// Tunggal / object
{ "success": true, "message": "...", "data": { ... } }

// Daftar
{ "success": true, "message": "...", "data": [ ... ], "meta": { ... } }

// Gagal umum
{ "success": false, "message": "..." }

// Gagal validasi (422)
{ "success": false, "message": "validasi gagal", "errors": { "field": "..." } }
```

---

## Status HTTP yang Dipakai

| Status | Situasi yang dibuktikan                                                 |
| ------ | ----------------------------------------------------------------------- |
| `200`  | Pengambilan daftar / satu data, PUT, PATCH berhasil                     |
| `201`  | POST berhasil — header `Location` ikut                                 |
| `204`  | DELETE berhasil — tanpa body                                            |
| `400`  | `id` bukan angka, atau body bukan JSON                                  |
| `404`  | Sumber daya tidak ditemukan / endpoint tidak dikenal                    |
| `409`  | NIM bentrok dengan mahasiswa lain                                       |
| `415`  | `Content-Type` bukan `application/json`                                 |
| `422`  | Validasi isi gagal, dengan rincian per field                            |

---

## Struktur Berkas

```
api-students/
├── go.mod
├── main.go     konfigurasi aplikasi, middleware, dan route
├── model.go    struct entitas, request, dan respons
├── helper.go   amplop respons, parser query string, dan requireJSON
└── handler.go  fungsi penangan tiap endpoint (list / get / POST / PUT / PATCH / DELETE)
```

---

## Sumber Bantuan

Bagian yang dibantu AI: penulisan struktur awal `model.go`, `helper.go`,
`handler.go`, dan `main.go` mengikuti pola dari modul Pertemuan 2;
seluruh perilaku, validasi, dan endpoint dijelaskan di atas ditulis
sendiri dan diuji secara manual dengan `curl`.