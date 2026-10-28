#!/usr/bin/env bash
# Uji manual endpoint portofolio. Output langsung tampil di terminal.
# Pakai: bash scripts/bukti-portofolio.sh
set -u

BASE="http://localhost:3000/api/v1"
JSON="Content-Type: application/json"

ADMIN=$(curl -s -X POST "$BASE/auth/login" -H "$JSON" \
  -d '{"username":"admin01","password":"rahasia123"}' \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')
AUTH="Authorization: Bearer $ADMIN"

# GET portofolio
curl -s -w "\n[%{http_code}]\n" "$BASE/students/2/portfolio" -H "$AUTH"

# PUT portofolio (membuat atau mengganti)
curl -s -w "\n[%{http_code}]\n" -X PUT "$BASE/students/2/portfolio" -H "$AUTH" -H "$JSON" -d '{
  "headline": "Backend Developer - Go & Fiber",
  "bio": "Mahasiswa Teknik Informatika yang sedang belajar membangun REST API dengan Go, Fiber, PostgreSQL, dan MongoDB.",
  "projects": [
    {"title": "Sistem Informasi Akademik", "technologies": ["Go", "Fiber", "PostgreSQL"], "url": "https://github.com/budi/siak", "year": 2025}
  ],
  "certificates": [
    {"name": "Certified Kubernetes Application Developer", "issuer": "CNCF", "date": "2025-06-01"}
  ]
}'

# POST proyek
curl -s -w "\n[%{http_code}]\n" -X POST "$BASE/students/2/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"API Student Service","technologies":["Go","MongoDB","Docker"],"url":"https://github.com/budi/api-student","year":2026}'

# GET statistik teknologi
curl -s -w "\n[%{http_code}]\n" "$BASE/portfolios/stats/technologies" -H "$AUTH"

# DELETE proyek
PROJ_JSON=$(curl -s "$BASE/students/2/portfolio" -H "$AUTH" | grep -o '"projects":\[.*\],"certificates"')
PROJECT_ID=$(echo "$PROJ_JSON" | grep -o '"id":"[a-f0-9]\{24\}"' | head -1 | cut -d'"' -f4)
echo "projectId = $PROJECT_ID"
curl -s -w "\n[%{http_code}]\n" -X DELETE "$BASE/students/2/portfolio/projects/$PROJECT_ID" -H "$AUTH"

# Gagal: mahasiswa tidak ada di PostgreSQL
curl -s -w "\n[%{http_code}]\n" "$BASE/students/99999/portfolio" -H "$AUTH"

# Gagal: NoSQL injection, operator di dalam body
curl -s -w "\n[%{http_code}]\n" -X PUT "$BASE/students/2/portfolio" -H "$AUTH" -H "$JSON" \
  -d '{"headline":{"$ne":""},"bio":{"$gt":""}}'

# Gagal: validasi input
curl -s -w "\n[%{http_code}]\n" -X POST "$BASE/students/2/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"","technologies":[],"url":"bukan-url","year":1999}'

# Gagal: token tidak dikirim
curl -s -w "\n[%{http_code}]\n" "$BASE/students/2/portfolio"

# Gagal: id proyek bukan ObjectID
curl -s -w "\n[%{http_code}]\n" -X DELETE "$BASE/students/2/portfolio/projects/bukan-objectid" -H "$AUTH"

# Batas 20 proyek: isi 18 proyek, lalu kirim 10 permintaan sekaligus
BODY='{"headline":"Backend Developer","bio":"Uji batas proyek.","projects":['
for i in $(seq 1 18); do
  [ "$i" -gt 1 ] && BODY="$BODY,"
  BODY="$BODY{\"title\":\"Proyek $i\",\"technologies\":[\"Go\",\"MongoDB\"],\"url\":\"https://example.com/p$i\",\"year\":2025}"
done
BODY="$BODY],\"certificates\":[]}"
curl -s -o /dev/null -X PUT "$BASE/students/3/portfolio" -H "$AUTH" -H "$JSON" -d "$BODY"

rm -rf /tmp/bukti_shot && mkdir -p /tmp/bukti_shot
for i in $(seq 1 10); do
  (
    curl -s -o "/tmp/bukti_shot/body$i.json" -w "%{http_code}\n" \
      -X POST "$BASE/students/3/portfolio/projects" -H "$AUTH" -H "$JSON" \
      -d "{\"title\":\"Serangan $i\",\"technologies\":[\"Go\"],\"url\":\"https://example.com/serang$i\",\"year\":2026}" \
      > "/tmp/bukti_shot/code$i.txt"
  ) &
done
wait

cat /tmp/bukti_shot/code*.txt | sort | uniq -c
mongosh --quiet "mongodb://localhost:27017/praktikum_backend" \
  --eval "print('jumlah proyek:', db.student_portfolios.findOne({student_id:3}).projects.length)"
