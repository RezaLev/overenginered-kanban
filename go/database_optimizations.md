# Dokumentasi Optimasi PostgreSQL (Skala 5 Juta Data)

Sebelum kita beralih ke OpenSearch (Query Model), berikut adalah dokumentasi rekam jejak optimasi yang telah kita lakukan pada database PostgreSQL untuk menangani beban 10 juta baris data (*records*) pada aplikasi Kanban Todo.

Optimasi ini berhasil menurunkan waktu *load* dari **10+ detik** menjadi **hitungan milidetik (< 500ms)**.

---

## 1. B-Tree Composite Indexing (Untuk Sorting & Pagination)
Pada awalnya, saat mencoba mengambil halaman pertama (LIMIT 10) dan mengurutkannya dari yang terbaru, PostgreSQL harus membaca dan menyortir seluruh 5 juta data di memori/disk (terjadi `Sort Method: external merge Disk`) yang memakan waktu belasan detik.

**Solusi:** 
Kita membuat sebuah *Composite Index* yang memetakan pola pengambilan data secara persis.
```sql
CREATE INDEX IF NOT EXISTS idx_todos_status_id ON todos (status ASC, id DESC);
```
**Hasil:** PostgreSQL dapat melakukan **Index-Only Scan**. Database tidak perlu lagi melakukan *sorting* manual, melainkan cukup menelusuri index yang sudah terurut dan langsung mengambil 10 data teratas.

---

## 2. Trigram Indexing (Untuk Full-Text Search)
Secara *default*, query pencarian seperti `WHERE title ILIKE '%makan%'` akan memaksa PostgreSQL melakukan *Sequential Scan* (membaca baris 1 sampai 5.000.000 satu per satu). Ini sangat lambat.

**Solusi:**
Kita mengaktifkan ekstensi `pg_trgm` dan memasang *Generalized Inverted Index (GIN)*.
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS trgm_idx_todos_title ON todos USING gin (title gin_trgm_ops);
```
**Hasil:** Index GIN akan memecah teks menjadi potongan-potongan 3 huruf (*trigram*). Ketika user mencari teks, PostgreSQL hanya mencocokkan potongan huruf tersebut dari index, menghindari *Sequential Scan*.

---

## 3. Query Execution Optimization (Menghindari ILIKE Kosong)
Saat Kanban *board* dimuat pertama kali tanpa ada kata pencarian, query yang dieksekusi sebelumnya adalah `WHERE title ILIKE '%%' AND status = 1`. 
Meskipun query tersebut secara logika mengembalikan semua data, PostgreSQL tetap melakukan filter *string matching* pada jutaan baris, membuat waktu tunggu facet dan kolom mencapai **1 - 2.6 detik**.

**Solusi:**
Kita merombak *Query Builder* di level repository Go. Jika `searchQuery` kosong (`""`), clause `ILIKE` dihilangkan sepenuhnya.
```go
if searchQuery != "" {
    // Gunakan ILIKE
} else {
    // Skip ILIKE, langsung ambil berdasarkan status
}
```
**Hasil:** PostgreSQL bisa langsung menggunakan *Index-Only Scan* untuk `status`, menurunkan waktu query secara drastis sebesar 80% (dari 2.6 detik menjadi ~500ms).

---

## 4. MVCC Count Bypass (Statistik Tabel Internal)
Menghitung total baris `SELECT count(*) FROM todos` sangat lambat di PostgreSQL (mencapai ~800ms) karena arsitektur MVCC (*Multi-Version Concurrency Control*) memaksa database memverifikasi visibilitas setiap baris.

**Solusi:**
Untuk penghitungan total awal (saat user pertama kali membuka aplikasi tanpa filter), kita tidak melakukan `COUNT(*)` biasa. Kita mengambil estimasi yang di-cache langsung dari tabel sistem PostgreSQL `pg_class`.
```sql
SELECT reltuples::bigint FROM pg_class WHERE relname = 'todos';
```
**Hasil:** Mendapatkan jumlah total data (5 juta) turun dari **~800ms** menjadi **~0.016ms** (instan).

---

Dengan fondasi di atas, arsitektur *Command* (Tulis) PostgreSQL kita sudah sangat optimal dan *enterprise-ready*. Menggeser beban *Read/Query* ke OpenSearch akan menjadi pelengkap sempurna untuk CQRS.
