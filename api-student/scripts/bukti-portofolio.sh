#!/usr/bin/env bash
# Bukti portofolio (Tugas Mandiri Modul 10).
# Pakai: bash scripts/bukti-portofolio.sh
#
# Semua perintah curl di bawah bisa dijalankan apa adanya di Git Bash.
set -u

BASE="http://localhost:3000/api/v1"
JSON="Content-Type: application/json"

login() {
  curl -s -X POST "$BASE/auth/login" -H "$JSON" \
    -d "{\"username\":\"$1\",\"password\":\"$2\"}" \
    | sed -E 's/.*"access_token":"([^"]+)".*/\1/'
}

ADMIN=$(login admin01 rahasia123)
AUTH="Authorization: Bearer $ADMIN"

line() { printf '\n=== %s ===\n' "$1"; }

line "0. Health check (PostgreSQL + MongoDB)"
curl -s -w "\n[%{http_code}]\n" "$BASE/health"

line "1. GET /students/:id/portfolio"
curl -s -w "\n[%{http_code}]\n" "$BASE/students/1/portfolio" -H "$AUTH"

line "2. PUT /students/:id/portfolio (membuat portofolio)"
curl -s -w "\n[%{http_code}]\n" -X PUT "$BASE/students/1/portfolio" -H "$AUTH" -H "$JSON" -d '{
  "headline": "Backend Developer - Go & Fiber",
  "bio": "Mahasiswa Teknik Informatika yang sedang belajar membangun REST API dengan Go, Fiber, PostgreSQL, dan MongoDB.",
  "projects": [
    {"title": "Sistem Informasi Akademik", "technologies": ["Go", "Fiber", "PostgreSQL"], "url": "https://github.com/budi/siak", "year": 2025}
  ],
  "certificates": [
    {"name": "Certified Kubernetes Application Developer", "issuer": "CNCF", "date": "2025-06-01"}
  ]
}'

line "3. POST /students/:id/portfolio/projects"
curl -s -w "\n[%{http_code}]\n" -X POST "$BASE/students/1/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"API Student Service","technologies":["Go","MongoDB","Docker"],"url":"https://github.com/budi/api-student","year":2026}'

line "4. GET /portfolios/stats/technologies"
curl -s -w "\n[%{http_code}]\n" "$BASE/portfolios/stats/technologies" -H "$AUTH"

line "5. DELETE /students/:id/portfolio/projects/:projectId"
PROJ_JSON=$(curl -s "$BASE/students/1/portfolio" -H "$AUTH" | grep -o '"projects":\[.*\],"certificates"')
PROJECT_ID=$(echo "$PROJ_JSON" | grep -o '"id":"[a-f0-9]\{24\}"' | head -1 | cut -d'"' -f4)
echo "projectId = $PROJECT_ID"
curl -s -w "\n[%{http_code}]\n" -X DELETE "$BASE/students/1/portfolio/projects/$PROJECT_ID" -H "$AUTH"

line "GAGAL A. mahasiswa tidak ada di PostgreSQL"
curl -s -w "\n[%{http_code}]\n" "$BASE/students/99999/portfolio" -H "$AUTH"

line "GAGAL B. NoSQL injection: operator di dalam body"
curl -s -w "\n[%{http_code}]\n" -X PUT "$BASE/students/1/portfolio" -H "$AUTH" -H "$JSON" \
  -d '{"headline":{"$ne":""},"bio":{"$gt":""}}'

line "GAGAL C. validasi gagal"
curl -s -w "\n[%{http_code}]\n" -X POST "$BASE/students/1/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"","technologies":[],"url":"bukan-url","year":1999}'

line "GAGAL D. tanpa token"
curl -s -w "\n[%{http_code}]\n" "$BASE/students/1/portfolio"

line "GAGAL E. id proyek bukan ObjectID"
curl -s -w "\n[%{http_code}]\n" -X DELETE "$BASE/students/1/portfolio/projects/bukan-objectid" -H "$AUTH"

line "BUKTI BATAS 20 PROYEK: isi 18 proyek, lalu 10 permintaan serentak"
BODY='{"headline":"Backend Developer","bio":"Uji batas proyek.","projects":['
for i in $(seq 1 18); do
  [ "$i" -gt 1 ] && BODY="$BODY,"
  BODY="$BODY{\"title\":\"Proyek $i\",\"technologies\":[\"Go\",\"MongoDB\"],\"url\":\"https://example.com/p$i\",\"year\":2025}"
done
BODY="$BODY],\"certificates\":[]}"
curl -s -o /dev/null -w "isi 18 proyek [%{http_code}]\n" -X PUT "$BASE/students/3/portfolio" -H "$AUTH" -H "$JSON" -d "$BODY"

rm -rf /tmp/bukti_shot && mkdir -p /tmp/bukti_shot
for i in $(seq 1 10); do
  (
    curl -s -o "/tmp/bukti_shot/body$i.json" -w "%{http_code}\n" \
      -X POST "$BASE/students/3/portfolio/projects" -H "$AUTH" -H "$JSON" \
      -d "{\"title\":\"Serangan $i\",\"technologies\":[\"Go\"],\"url\":\"https://example.com/s$i\",\"year\":2026}" \
      > "/tmp/bukti_shot/code$i.txt"
  ) &
done
wait
echo "--- status code 10 permintaan ---"
cat /tmp/bukti_shot/code*.txt | sort | uniq -c
echo "--- jumlah proyek akhir ---"
mongosh --quiet "mongodb://localhost:27017/praktikum_backend" \
  --eval "print('jumlah proyek:', db.student_portfolios.findOne({student_id:3}).projects.length)"
