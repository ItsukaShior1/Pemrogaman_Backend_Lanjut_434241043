#!/usr/bin/env bash
# Menjalankan dan menyimpan seluruh output bukti portofolio.
# Output tampil di terminal sekaligus disimpan ke folder docs/bukti/ untuk
# dipakai generator laporan.
#
# Pakai: bash scripts/kumpulkan-bukti.sh
set -u

OUT="docs/bukti"
BASE="http://localhost:3000/api/v1"
JSON="Content-Type: application/json"
mkdir -p "$OUT"

ADMIN=$(curl -s -X POST "$BASE/auth/login" -H "$JSON" \
  -d '{"username":"admin01","password":"rahasia123"}' \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')
AUTH="Authorization: Bearer $ADMIN"

# User biasa, dipakai untuk membuktikan kasus 403 (bukan pemilik portofolio).
LAIN=$(curl -s -X POST "$BASE/auth/login" -H "$JSON" \
  -d '{"username":"testmahasiswa","password":"RahasiaKuat123!"}' \
  | sed -E 's/.*"access_token":"([^"]+)".*/\1/')

save() { # save <nama-file> <curl args...>
  local name="$1"; shift
  {
    echo "\$ curl $*" | sed -e "s|$ADMIN|\$TOKEN|g" -e "s|$LAIN|\$TOKEN|g"
    echo
    curl -s -w "\n[HTTP %{http_code}]\n" "$@" \
      | sed -e "s/$ADMIN/\$TOKEN/g" -e "s/$LAIN/\$TOKEN/g"
  } | tee "$OUT/$name.txt"
  echo "  -> $OUT/$name.txt"
}

judul() { printf '\n=== %s ===\n' "$1"; }

echo "Menyiapkan portofolio contoh (student 1) untuk tangkapan layar Compass..."
curl -s -o /dev/null -X PUT "$BASE/students/1/portfolio" -H "$AUTH" -H "$JSON" -d '{
  "headline": "Backend Developer - Go, Fiber, PostgreSQL & MongoDB",
  "bio": "Mahasiswa Teknik Informatika semester 5 yang sedang membangun REST API polyglot persistence: data mahasiswa tetap di PostgreSQL, sedangkan portofolio disimpan di MongoDB. Tertarik pada arsitektur berlapis, concurrency control, dan pemrograman backend dengan Go.",
  "projects": [
    {"title": "Sistem Informasi Akademik (SIA)", "technologies": ["Go", "Fiber", "PostgreSQL", "React"], "url": "https://github.com/icadgtng/siak", "year": 2026},
    {"title": "API Student Service", "technologies": ["Go", "Fiber", "MongoDB", "Docker"], "url": "https://github.com/icadgtng/api-student", "year": 2026},
    {"title": "Sistem Antrean Klinik", "technologies": ["Go", "PostgreSQL", "Redis", "WebSocket"], "url": "https://github.com/icadgtng/antrean-klinik", "year": 2025},
    {"title": "Website Portfolio Responsif", "technologies": ["HTML", "CSS", "JavaScript"], "url": "https://icadgtng.github.io/portfolio", "year": 2025}
  ],
  "certificates": [
    {"name": "Certified Kubernetes Application Developer (CKAD)", "issuer": "The Linux Foundation", "date": "2025-06-01"},
    {"name": "Go: The Complete Developer Guide", "issuer": "Udemy", "date": "2025-03-15"},
    {"name": "MongoDB Associate Developer", "issuer": "MongoDB Inc.", "date": "2026-02-20"}
  ]
}'

mongosh --quiet "mongodb://localhost:27017/praktikum_backend" --eval "
const d = db.student_portfolios.findOne({student_id: 1});
print(JSON.stringify(d, null, 2));
" > "$OUT/00-contoh-dokumen.json"
echo "  -> $OUT/00-contoh-dokumen.json"

echo "Mengumpulkan bukti..."

judul "Health check (PostgreSQL + MongoDB)"
save 14-health-check "$BASE/health"

judul "Endpoint portofolio"
save 01-get-portfolio      "$BASE/students/2/portfolio" -H "$AUTH"
save 02-put-portfolio      -X PUT "$BASE/students/2/portfolio" -H "$AUTH" -H "$JSON" \
  -d '{"headline":"Backend Developer - Go, Fiber, PostgreSQL & MongoDB","bio":"Mahasiswa Teknik Informatika semester 5 yang sedang membangun REST API polyglot persistence.","projects":[{"title":"Sistem Informasi Akademik (SIA)","technologies":["Go","Fiber","PostgreSQL","React"],"url":"https://github.com/icadgtng/siak","year":2026}],"certificates":[{"name":"MongoDB Associate Developer","issuer":"MongoDB Inc.","date":"2026-02-20"}]}'

PROJECT_ID=$(curl -s "$BASE/students/2/portfolio" -H "$AUTH" \
  | grep -o '"projects":\[.*\],"certificates"' | grep -o '"id":"[a-f0-9]\{24\}"' | head -1 | cut -d'"' -f4)
save 03-post-project      -X POST "$BASE/students/2/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"API Student Service","technologies":["Go","MongoDB","Docker"],"url":"https://github.com/icadgtng/api-student","year":2026}'

save 04-stats-teknologi   "$BASE/portfolios/stats/technologies" -H "$AUTH"

NEW_ID=$(curl -s "$BASE/students/2/portfolio" -H "$AUTH" \
  | grep -o '"projects":\[.*\],"certificates"' | grep -o '"id":"[a-f0-9]\{24\}"' | head -1 | cut -d'"' -f4)
save 05-delete-project    -X DELETE "$BASE/students/2/portfolio/projects/$NEW_ID" -H "$AUTH"

judul "Kasus gagal"
save 06-gagal-student-tidak-ada  "$BASE/students/99999/portfolio" -H "$AUTH"
save 07-gagal-nosql-injection    -X PUT "$BASE/students/2/portfolio" -H "$AUTH" -H "$JSON" \
  -d '{"headline":{"$ne":""},"bio":{"$gt":""}}'
save 08-gagal-validasi           -X POST "$BASE/students/2/portfolio/projects" -H "$AUTH" -H "$JSON" \
  -d '{"title":"","technologies":[],"url":"bukan-url","year":1999}'
save 09-gagal-tanpa-token        "$BASE/students/2/portfolio"
save 10-gagal-id-proyek-invalid  -X DELETE "$BASE/students/2/portfolio/projects/bukan-objectid" -H "$AUTH"
save 11-gagal-tidak-memiliki     "$BASE/students/2/portfolio" -H "Authorization: Bearer $LAIN"

judul "Bukti batas 20 proyek (18 proyek + 10 permintaan serentak)"
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
      -d "{\"title\":\"Serangan $i\",\"technologies\":[\"Go\"],\"url\":\"https://example.com/s$i\",\"year\":2026}" \
      > "/tmp/bukti_shot/code$i.txt"
  ) &
done
wait

{
  echo "Isi 18 proyek pada portofolio student_id=3, lalu kirim 10 permintaan POST"
  echo "  /students/3/portfolio/projects secara serentak."
  echo
  echo "Status code masing-masing permintaan:"
  cat /tmp/bukti_shot/code*.txt | sort | uniq -c
  echo
  echo "Jumlah proyek pada dokumen setelah semuanya selesai:"
  mongosh --quiet "mongodb://localhost:27017/praktikum_backend" \
    --eval "print('  jumlah proyek:', db.student_portfolios.findOne({student_id:3}).projects.length)"
  echo
  echo "Contoh respons yang ditolak (409):"
  grep -l "sudah mencapai batas" /tmp/bukti_shot/body*.json | head -1 | xargs cat
} | tee "$OUT/12-bukti-batas-20-proyek.txt"
echo "  -> $OUT/12-bukti-batas-20-proyek.txt"

judul "Bukti explain()"
mongosh --quiet "mongodb://localhost:27017/praktikum_backend" --eval "
function indexOf(plan) {
  if (!plan) return '(tidak ada)';
  if (plan.indexName) return plan.indexName;
  if (plan.inputStage) return indexOf(plan.inputStage);
  if (plan.inputStages) {
    for (const s of plan.inputStages) { const n = indexOf(s); if (n !== '(tidak ada)') return n; }
  }
  return '(COLLSCAN - index tidak dipakai)';
}
const doc = db.student_portfolios.findOne({student_id: 1});
const satuProyek = doc.projects[0]._id;

print('=== Index pada collection student_portfolios ===');
db.student_portfolios.getIndexes().forEach(i => printjson(i));
print('');
print('=== explain(\"executionStats\") untuk pencarian satu portofolio ===');
const e1 = db.student_portfolios.find({student_id: 1}).explain('executionStats');
print('winning plan      :', JSON.stringify(e1.queryPlanner.winningPlan));
print('index dipakai     :', indexOf(e1.queryPlanner.winningPlan));
print('totalKeysExamined :', e1.executionStats.totalKeysExamined);
print('totalDocsExamined :', e1.executionStats.totalDocsExamined);
print('nReturned         :', e1.executionStats.nReturned);
print('');
print('=== explain(\"executionStats\") untuk pencarian proyek (projects._id) ===');
const e2 = db.student_portfolios.find({'projects._id': satuProyek}).explain('executionStats');
print('index dipakai     :', indexOf(e2.queryPlanner.winningPlan));
print('totalKeysExamined :', e2.executionStats.totalKeysExamined);
print('totalDocsExamined :', e2.executionStats.totalDocsExamined);
print('nReturned         :', e2.executionStats.nReturned);
print('');
print('=== explain(\"executionStats\") untuk pencarian teknologi (multikey) ===');
const e3 = db.student_portfolios.find({'projects.technologies': 'go'}).explain('executionStats');
print('index dipakai     :', indexOf(e3.queryPlanner.winningPlan));
print('isMultiKey        :', e3.queryPlanner.winningPlan.inputStage.isMultiKey);
print('totalKeysExamined :', e3.executionStats.totalKeysExamined);
print('totalDocsExamined :', e3.executionStats.totalDocsExamined);
print('nReturned         :', e3.executionStats.nReturned);
print('');
print('=== Bukti index unik: mencoba membuat portofolio kedua untuk student_id 1 ===');
try {
  db.student_portfolios.insertOne({student_id: 1, headline: 'duplikat', bio: 'x', projects: [], certificates: []});
  print('DUPLIKAT DITERIMA ( seharusnya ditolak)');
} catch (e) { print('DITOLAK oleh index ->', e.message); }
" | tee "$OUT/13-bukti-explain.txt"
echo "  -> $OUT/13-bukti-explain.txt"

echo ""
echo "Selesai. Semua bukti ada di $OUT/"
