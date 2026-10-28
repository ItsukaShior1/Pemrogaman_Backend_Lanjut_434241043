// Bukti index dipakai pada collection student_portfolios.
// Jalankan: mongosh mongodb://localhost:27017/praktikum_backend scripts/bukti-index.js
//
// Perintah mongosh --eval di bawah juga bisa dipakai langsung di PowerShell:
//   mongosh "mongodb://localhost:27017/praktikum_backend" --eval 'db.students.countDocuments()'

function indexOf(plan) {
  if (!plan) return '(tidak ada)';
  if (plan.indexName) return plan.indexName;
  if (plan.inputStage) return indexOf(plan.inputStage);
  if (plan.inputStages) {
    for (const s of plan.inputStages) {
      const n = indexOf(s);
      if (n !== '(tidak ada)') return n;
    }
  }
  return '(COLLSCAN - index tidak dipakai)';
}

const dbName = 'praktikum_backend';
const coll = db.getSiblingDB(dbName).student_portfolios;

print('=== Index yang ada pada student_portfolios ===');
coll.getIndexes().forEach(i => printjson(i));
print('');

print('=== 1. Cari satu portofolio (student_id) ===');
const e1 = coll.find({ student_id: 1 }).explain('executionStats');
print('winning plan      :', JSON.stringify(e1.queryPlanner.winningPlan));
print('index dipakai     :', indexOf(e1.queryPlanner.winningPlan));
print('totalKeysExamined :', e1.executionStats.totalKeysExamined);
print('totalDocsExamined :', e1.executionStats.totalDocsExamined);
print('nReturned         :', e1.executionStats.nReturned);
print('');

print('=== 2. Cari satu proyek (projects._id) ===');
const satuProyek = coll.findOne({ student_id: 1 }).projects[0]._id;
const e2 = coll.find({ 'projects._id': satuProyek }).explain('executionStats');
print('index dipakai     :', indexOf(e2.queryPlanner.winningPlan));
print('totalKeysExamined :', e2.executionStats.totalKeysExamined);
print('totalDocsExamined :', e2.executionStats.totalDocsExamined);
print('nReturned         :', e2.executionStats.nReturned);
print('');

print('=== 3. Cari berdasarkan teknologi (multikey index) ===');
const e3 = coll.find({ 'projects.technologies': 'go' }).explain('executionStats');
print('index dipakai     :', indexOf(e3.queryPlanner.winningPlan));
print('isMultiKey        :', e3.queryPlanner.winningPlan.inputStage.isMultiKey);
print('totalKeysExamined :', e3.executionStats.totalKeysExamined);
print('totalDocsExamined :', e3.executionStats.totalDocsExamined);
print('nReturned         :', e3.executionStats.nReturned);
print('');

print('=== 4. Bukti index unik menahan (satu mahasiswa satu portofolio) ===');
try {
  coll.insertOne({ student_id: 1, headline: 'duplikat', bio: 'x', projects: [], certificates: [] });
  print('DUPLIKAT DITERIMA ( seharusnya ditolak)');
} catch (e) {
  print('DITOLAK oleh index ->', e.message);
}
