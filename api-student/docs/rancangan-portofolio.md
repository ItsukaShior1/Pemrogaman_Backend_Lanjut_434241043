# Rancangan Portofolio (Tugas Mandiri Modul 10)

Dokumen ini adalah bahan laporan `Tugas10_NIM.pdf`: rancangan dokumen,
potongan kode penting, dan jawaban analisis.

---

## C.1 Rancangan Dokumen

### Collection

`student_portfolios` di database `praktikum_backend` (MongoDB).

Satu mahasiswa di PostgreSQL (`students.id`) memiliki **tepat satu** dokumen
portofolio di MongoDB. Relasi ini bukan foreign key — MongoDB tidak mengenal
PostgreSQL — melainkan dicek oleh service sebelum menulis, dan dijaga index
unik di database.

### Bentuk dokumen

```json
{
  "_id": ObjectId("6ae1f12deeb4d55fd42ac318"),
  "student_id": 1,
  "headline": "Backend Developer - Go & Fiber",
  "bio": "Mahasiswa Teknik Informatika yang sedang belajar membangun REST API dengan Go, Fiber, PostgreSQL, dan MongoDB.",
  "projects": [
    {
      "_id": ObjectId("6ae1f12d045e57798a8d5ffe"),
      "title": "Sistem Informasi Akademik",
      "technologies": ["go", "fiber", "postgresql"],
      "url": "https://github.com/budi/siak",
      "year": 2025
    }
  ],
  "certificates": [
    {
      "_id": ObjectId("6ae1f12d045e57798a8d5fff"),
      "name": "Certified Kubernetes Application Developer",
      "issuer": "CNCF",
      "date": ISODate("2025-06-01T00:00:00Z")
    }
  ],
  "created_at": ISODate("2026-10-28T12:57:17.935Z"),
  "updated_at": ISODate("2026-10-28T12:57:17.935Z")
}
```

Catatan per field:

| Field | Tipe | Alasan |
|---|---|---|
| `student_id` | int | Kunci ke `students.id` di PostgreSQL. Disimpan sebagai int, bukan ObjectID, karena yang dirujuk adalah baris tabel. |
| `headline` | string | Satu kalimat profil, selalu dibaca bersama portofolio. |
| `bio` | string | Deskripsi diri, selalu dibaca bersama portofolio. |
| `projects` | array of object | Disarangkan (embedded): selalu dibaca bersama induknya, dibatasi 20, tidak dipakai dokumen lain. |
| `certificates` | array of object | Disarangun dengan alasan yang sama, dibaca bersama portofolio. |
| `created_at`, `updated_at` | date | `updated_at` berubah setiap kali portofolio atau proyeknya berubah. |

Waktu dibulatkan ke milidetik sebelum disimpan, karena tipe `date` BSON hanya
menyimpan sampai milidetik sedangkan `time.Time` di Go sampai nanodetik.

### Index

| Nama | Keys | Jenis | Alasan |
|---|---|---|---|
| `portfolio_student_unique` | `student_id: 1` | **unique** | Menegakkan "satu mahasiswa satu portofolio" di level database, bukan hanya di kode. |
| `portfolio_technologies` | `projects.technologies: 1` | multikey | Mencari portofolio yang memakai suatu teknologi; satu entri indeks per elemen array. |
| `portfolio_project_ids` | `projects._id: 1` | multikey | Mempercepat pencarian proyek saat `DELETE .../projects/:projectId` menuliskan filter `projects._id`. |

Bukti index unik bekerja (percobaan membuat portofolio kedua untuk
`student_id: 1`):

```
E11000 duplicate key error collection: praktikum_backend.student_portfolios
index: portfolio_student_unique dup key: { student_id: 1 }
```

### Batas 20 proyek

Batas diperiksa **di dalam filter** operasi yang sama dengan penambahan, bukan
dibaca dulu di Go lalu ditulis. Pola `projects.19` berasal dari pola
`replies.49` di Modul 10 Langkah 6:

```go
bson.M{"student_id": studentID, "projects.19": bson.M{"$exists": false}}
```

Array yang indeks 19-nya belum ada berarti isinya belum 20 elemen. Karena
`findOneAndUpdate` bersifat atomik per dokumen, dua permintaan serentak tidak
dapat sama-sama lolos pada slot terakhir.

### Normalisasi teknologi

Daftar teknologi disimpan dalam huruf kecil, tanpa spasi tepi, tanpa duplikat.
Tanpa itu, "Go", "go", dan " GO " akan terhitung sebagai tiga teknologi
berbeda pada statistik, padahal maksudnya sama. Aturan ini sama persis dengan
`NormalizeTags` pada fitur ulasan di Modul 10.

---

## C.2 Endpoint

| Method | Path | Sumber data | Keterangan |
|---|---|---|---|
| GET | `/api/v1/students/:id/portfolio` | MongoDB | 404 bila mahasiswa tidak ada di PostgreSQL, 404 bila portofolio belum dibuat, 403 bila bukan pemiliknya. |
| PUT | `/api/v1/students/:id/portfolio` | MongoDB | Upsert: membuat bila belum ada, mengganti seluruh isi bila sudah ada. |
| POST | `/api/v1/students/:id/portfolio/projects` | MongoDB | `$push` satu proyek, maks 20 secara atomik. 409 bila sudah 20. |
| DELETE | `/api/v1/students/:id/portfolio/projects/:projectId` | MongoDB | `$pull` satu proyek. 404 bila portofolio atau proyeknya tidak ada. |
| GET | `/api/v1/portfolios/stats/technologies` | MongoDB | Aggregation dua kali `$unwind`, 10 teknologi terpopuler. |

Hak akses mengikuti model `students` yang sudah ada: pemilik portofolio
(`students.owner_id`) boleh akses, selain itu butuh permission
`student:read:any` / `student:update:any`. Tidak ada permission baru.

### Bentuk request

```jsonc
// PUT /students/:id/portfolio
{
  "headline": "Backend Developer - Go & Fiber",
  "bio": "deskripsi diri",
  "projects": [
    { "title": "Judul", "technologies": ["Go", "Fiber"], "url": "https://...", "year": 2025 }
  ],
  "certificates": [
    { "name": "Nama sertifikat", "issuer": "Penerbit", "date": "2025-06-01" }
  ]
}

// POST /students/:id/portfolio/projects
{ "title": "Judul", "technologies": ["Go"], "url": "https://...", "year": 2026 }
```

### Potongan kode penting

Filter atomis penambahan proyek (repository):

```go
limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

err = r.coll.FindOneAndUpdate(ctx,
    bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
    bson.M{
        "$push": bson.M{"projects": doc},
        "$set":  bson.M{"updated_at": now()},
    },
    options.FindOneAndUpdate().SetReturnDocument(options.After),
).Decode(&updated)
```

Penghapusan proyek dengan `$pull`:

```go
res, err := r.coll.UpdateOne(ctx,
    bson.M{"student_id": studentID, "projects._id": oid},
    bson.M{
        "$pull": bson.M{"projects": bson.M{"_id": oid}},
        "$set":  bson.M{"updated_at": now()},
    },
)
```

Statistik teknologi dengan dua kali `$unwind`:

```go
pipeline := mongo.Pipeline{
    {{Key: "$unwind", Value: "$projects"}},
    {{Key: "$unwind", Value: "$projects.technologies"}},
    {{Key: "$group", Value: bson.M{
        "_id":   "$projects.technologies",
        "total": bson.M{"$sum": 1},
    }}},
    {{Key: "$sort", Value: bson.D{
        {Key: "total", Value: -1},
        {Key: "_id", Value: 1},
    }}},
    {{Key: "$limit", Value: int64(limit)}},
}
```

---

## C.4 Analisis Singkat

### 10. Mengapa projects disarangkan, dan kapan keputusannya dibalik

`projects` disarungkan karena tiga syarat embedding terpenuhi: proyek selalu
dibaca bersama portofolionya (halaman profil tidak pernah menampilkan daftar
proyek tanpa deskripsi pemiliknya), jumlahnya dibatasi 20 sehingga tidak akan
menabrak batas dokumen 16 MB, dan satu proyek tidak pernah dipakai bersama
portofolio mahasiswa lain. Menyimpan proyek di collection terpisah justru
menambah satu query pada setiap pembacaan portofolio tanpa memberi keuntungan
apa pun. Keputusan ini harus dibalik bila satu kebutuhan berubah: bila suatu
saat proyek boleh disunting dan dibahas **secara kolaboratif** oleh banyak
mahasiswa, atau bila jumlah proyek tidak lagi dibatasi sehingga satu
portofolio bisa tumbuh tak terbatas, proyek harus dipindah ke collection
tersendiri yang dirujuk lewat `project_id`. Contoh perubahan kebutuhan yang
paling nyata: portofolio dipublikasikan ke lowongan kerja dan tiap proyek
butuh halaman próprio dengan ulasan, jumlah suka, dan komentar — datanya
tidak lagi milik satu dokumen saja.

### 11. Bila mahasiswa dihapus dari PostgreSQL, apa yang terjadi pada portofolionya

Tidak ada foreign key lintas database, jadi portofolio **tetap tinggal** di
MongoDB sebagai data yatim. `students.owner_id` memakai `ON DELETE RESTRICT`,
sehingga biasanya mahasiswa tidak bisa dihapus selama masih memiliki data
lain yang merujuk, tetapi portofolio di MongoDB tidak termasuk database itu
dan tidak menahan apa pun. Konsekuensinya: `GET /students/:id/portfolio`
menolak dengan 404 karena mahasiswanya sudah tidak ada di PostgreSQL, tetapi
dokumennya masih ada di `student_portfolios` dan masih bisa ditemukan lewat
query langsung, sehingga data pribadi mahasiswa yang sudah tidak aktif tetap
tersimpan.

Dua cara menanganinya:

1. **Penghapusan berkala (reconciliation job).** Sebuah job terjadwal
   mengambil semua `student_id` di MongoDB, membandingkan dengan `students.id`
   di PostgreSQL, lalu menghapus portofolio yang induknya sudah hilang.
   Kelemahannya: ada jeda antara penghapusan mahasiswa dan penghapusan
   portofolionya, sehingga data yatim tetap ada selama selang waktu itu —
   dan bila job gagal atau tidak dijadwalkan, data yatim bisa menumpuk
   permanen. Jalur ini juga hanya bisa berjalan bila MongoDB hidup.
2. **Penghapusan dari sisi service (event-driven).** Endpoint
   `DELETE /students/:id` memanggil repository MongoDB untuk menghapus
   portofolio sebelum atau sesudah baris mahasiswa dihapus dari PostgreSQL.
   Kelemahannya: tidak ada transaksi lintas database, jadi bila langkah kedua
   gagal, portofolio sudah hilang tetapi mahasiswanya masih ada (atau
   sebaliknya). Ada juga kemungkinan request dibatalkan di tengah jalan.
   Dibuat lebih aman dengan pola "hapus MongoDB dulu, baru PostgreSQL" dan
   retry, atau dengan menuliskan eventsourcing-style outbox.

Cara pertama lebih tahan terhadap kegagalan sedangkan cara kedua lebih
segera. Kombinasi keduanya — penghapusan sinkron di service untuk kasus
normal, ditambah reconciliation job untuk membersihkan sisa — adalah pilihan yang
paling seimbang.

### 12. Endpoint mana yang tetap berjalan bila MongoDB mati

Dengan pemisahan PostgreSQL/MongoDB, matinya MongoDB hanya mematikan endpoint
yang menyentuh MongoDB. Endpoint PostgreSQL tetap berjalan normal:

| Endpoint | Sumber | Saat MongoDB mati |
|---|---|---|
| `GET /api/v1/health` | keduanya | 503 "MongoDB tidak dapat dihubungi" (2 detik) |
| `GET /api/v1/students` | PostgreSQL | 200 (0,2 detik) |
| `GET /api/v1/students/:id` | PostgreSQL | 200 (0,2 detik) |
| `GET /api/v1/users`, `/auth/login` | PostgreSQL | 200 |
| `GET /api/v1/courses` | PostgreSQL | 200 |
| `GET /api/v1/students/:id/portfolio` | MongoDB | 500 (5 detik) |
| `POST /api/v1/students/:id/portfolio/projects` | MongoDB | 500 (5 detik) |
| `GET /api/v1/portfolios/stats/technologies` | MongoDB | 500 (5 detik) |
| `GET /api/v1/courses/:id/reviews` | MongoDB | 500 (5 detik) |

Perilaku ini memang yang diinginkan, karena satu bagian yang mati tidak
menjatuhkan seluruh aplikasi. Namun ada dua hal yang perlu diperbaiki:
pertama, endpoint MongoDB sebaiknya menjawab **503 Service Unavailable**,
bukan 500, karena penyebabnya bukan kesalahan kode melainkan ketergantungan
yang sedang tidak tersedia; kedua, `GET /health` sudah menjawab 503 dengan
pesan yang menyebut MongoDB, sehingga load balancer atau uptime monitor dapat
membedakan "aplikasi rusak" dari "satu dependensi mati". Setelah MongoDB
dinyalakan lagi, aplikasi pulih sendiri tanpa perlu restart, karena driver
mongo establishing koneksi baru secara otomatis pada permintaan berikutnya.
