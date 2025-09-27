```markdown
# Soal Tes Take-home — Fullstack CMS

Ringkasan singkat

- Tugas: Bangun sebuah Content Management System (CMS) fullstack production-ready.
- Waktu rekomendasi: 3–5 hari kerja

Tujuan

- Menilai kemampuan kandidat dalam merancang dan membangun aplikasi web end-to-end: backend, admin UI, API publik, keamanan, testing, observability, dan operasional.

Deliverables (WAJIB dikumpulkan):

1. Repository kode sumber dengan struktur rapi.
2. README lengkap berisi:
   - Cara menjalankan di development. `(make dev)`
   - Cara menjalankan test suite. `(-)`
   - Instruksi migrasi DB dan seed data. `(make migrate-up)`
   - Daftar endpoint API + contoh request/response. `(ada file postman di folder docs)`
   - Akun demo (username/password) dan peran. `("email": "admin@example.com", "password": "admin123" - Role Admin)`
   - Singkat trade-offs/pilihan arsitektural. `(trade-offs: JSONB, REST, FS Media)`
3. Skrip / mekanisme migrasi database.
4. Test suite: unit tests, integration tests, dan minimal 1 e2e test.
5. Contoh data (seed) dan akun pengguna demo.
6. Instruksi deployment (container/manifest). Demo deploy publik sangat dihargai.

Requirement Fungsional (MUST)

1. Content modeling & CRUD

   - Dukung definisi content type dinamis (pages, blog posts, custom types).
   - Field minimal: string, rich text/markdown, number, boolean, datetime, relation.
   - Admin UI untuk membuat content type dan entry (create/read/update/delete).
   - API publik untuk content types dan entries.

2. Rich editor & media

   - Editor WYSIWYG atau markdown.
   - Upload media (images/files) tersimpan dan punya URL publik.
   - Minimal: transformasi gambar (resize/thumbnail).

3. Routing & URL

   - Slug management, custom permalinks, preview draft (preview hanya untuk authenticated user).

4. Pengguna & RBAC

   - Minimal roles: Admin, Editor, Viewer.
   - Permission granular: create/edit/publish/delete pada resources.

5. Versioning & Audit

   - Simpan riwayat versi konten; audit log (siapa, kapan, ringkasan perubahan).
   - Fitur rollback ke versi sebelumnya.

6. Publishing workflow & scheduling

   - Draft vs Published; scheduling untuk publish di masa mendatang.

7. API publik

   - REST atau GraphQL yang mendukung pagination, filtering, sorting, field selection.
   - Dokumentasi API (OpenAPI/Swagger atau docs).

8. Security dasar

   - Proteksi CSRF, sanitasi input untuk XSS, validasi upload, penggunaan query parameterized.
   - Penanganan secrets (tidak commit secret ke repo).

9. Testing

   - Unit tests untuk logika penting.
   - Integration tests untuk API.
   - Minimal 1 e2e test: contoh flow (login → buat post → publish → cek visible via API).

10. Deployment & Operasional (Done All)

- Mekanisme build & deploy (container image atau paket). (Done: file Makefile / command: "make run")
- Script migrasi DB, strategi backup & restore (dokumen atau script). (Done: file Makefile mengenai migrate)
- Health checks. (Done: http://localhost:4000/healthz)

Requirement Non-fungsional

- Caching untuk konten publik (HTTP cache headers, CDN-friendly).
- Observability: structured logging (JSON), basic metrics (request count, latencies), health endpoint.
- Maintainability: kode modular, dokumentasi arsitektural, konfigurasi environment yang terpisah.

Fitur Bonus (opsional — poin ekstra)

- Desain untuk horizontal scaling (stateless services + external storage untuk media).
- Multi-tenant support.
- Plugin/extension system.
- Full-text search terintegrasi.
- Contoh frontend (headless + sample site).
- SSO / OAuth2.
- Content localization (i18n).
- Realtime collaboration.

Acceptance Criteria (yang harus bisa direproduksi reviewer)

- Kode dapat dibuild/run lewat script yang jelas (Makefile / scripts / docker-compose).

Struktur repository yang direkomendasikan

- /README.md
- /docs/ (API docs, design decisions)
- /server/ (backend service menggunakan Golang Gorila)
- /admin-ui/ (admin frontend menggunakan Next.js)
- /web/ (sample public frontend — opsional)
- /migrations/
- /scripts/ (build/start/migrate/backup/restore)
- /tests/ (unit/integration/e2e)

Objektif Penilaian (total 100 poin)

- Arsitektur & desain — 25
- Kualitas kode — 20
- Keamanan — 15
- Testing — 15
- Dokumentasi & Reproducibility — 10
- Fungsional completeness — 10
- Bonus — up to 10
```
