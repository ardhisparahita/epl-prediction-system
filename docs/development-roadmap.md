# Development Roadmap — EPL Prediction System

Roadmap ini menyesuaikan urutan kerja project dengan review dosen dan keputusan metodologi yang sudah ditetapkan. Backend infrastructure yang sudah dibuat tetap dipertahankan; pengembangan fitur domain dan migration menunggu spesifikasi serta skema data dikunci.

## Status saat ini

| Pekerjaan | Status |
|---|---|
| GitHub Monorepo `epl-prediction-system` | Selesai |
| Branch `main` dan `develop` | Sudah dibuat |
| Go module dan Fiber v3 bootstrap | Selesai |
| Health check endpoint | Selesai |
| Environment configuration | Berhasil dimuat |
| PostgreSQL Docker Compose | Sudah ditambahkan ke `develop` |
| GORM PostgreSQL connection | Berhasil dijalankan |
| Research specification | Sedang difinalisasi |
| ERD dan SQL migration | Belum dimulai |

## Prinsip kerja

1. Setiap pekerjaan baru dimulai dari branch terbaru `develop`.
2. Buat branch sesuai lingkup pekerjaan sebelum coding.
3. Kerjakan satu lingkup perubahan, jalankan format/lint/test yang relevan, lalu commit.
4. Push branch dan buat Pull Request ke `develop`.
5. Merge setelah perubahan ditinjau dan tes lulus.
6. Jangan commit `.env`, password, dataset mentah, atau file lokal/hasil sementara yang tidak diperlukan.
7. Perubahan metodologi harus diperbarui pada `docs/research-specification.md` dan naskah skripsi.

## Fase 1 — Research specification dan audit eksperimen

### 1.1 Dokumentasi spesifikasi

- [ ] Simpan `docs/research-specification.md`.
- [ ] Simpan `docs/development-roadmap.md`.
- [ ] Tinjau keputusan fitur: semua venue, maksimal lima laga, minimal tiga laga sebelumnya.
- [ ] Tinjau penanganan SOT kosong dan jumlah SOT valid.
- [ ] Kunci penggunaan timestamp kickoff dan timezone.
- [ ] Kunci parameter teknis Logistic Regression.
- [ ] Kunci cara admin mengelola jadwal pertandingan.

Branch yang disarankan: `docs/research-specification`.

### 1.2 Audit dan finalisasi eksperimen Python

- [ ] Audit sembilan CSV dan catat jumlah baris serta kolom per musim.
- [ ] Periksa duplikasi, konsistensi `FTR` dengan skor, nama tim, tanggal, waktu, dan missing value.
- [ ] Bentuk target H/D/A.
- [ ] Bentuk 12 fitur dengan `shift(1)` dan rolling window maksimal lima pertandingan.
- [ ] Pastikan minimal tiga pertandingan sebelumnya untuk masing-masing tim.
- [ ] Hitung SOT dari nilai valid; jangan ubah missing menjadi nol.
- [ ] Pisahkan training 2018/19–2023/24, validation 2024/25, test 2025/26, dan live 2026/27.
- [ ] Pilih konfigurasi menggunakan validation, bukan final test.
- [ ] Refit model uji menggunakan training + validation, lalu evaluasi satu kali pada test.
- [ ] Laporkan baseline, accuracy, macro precision/recall/F1, confusion matrix, dan log loss.
- [ ] Buat artefak model live terpisah: training sampai 2025/26, lalu freeze.
- [ ] Buat fungsi inference yang menggunakan tepat 12 fitur dalam urutan yang sama dengan training.
- [ ] Uji beberapa pasangan tim dan verifikasi probabilitas H/D/A.

Branch yang disarankan: `feat/ml-experiment-finalization`.

## Fase 2 — Data architecture

### 2.1 Perancangan data

- [ ] Buat ERD berdasarkan kebutuhan sistem dan hasil audit data.
- [ ] Tetapkan relasi teams dan matches.
- [ ] Tetapkan status pertandingan dan aturan penyimpanan kickoff.
- [ ] Tetapkan struktur model version dan identitas artefak.
- [ ] Tetapkan struktur prediction dan snapshot 12 fitur.
- [ ] Putuskan kebutuhan tabel users/admin.
- [ ] Tinjau constraint, unique index, foreign key, dan indeks pencarian.
- [ ] Cocokkan ERD dengan API contract.

Branch yang disarankan: `docs/database-design`.

### 2.2 SQL migration

- [ ] Pilih dan pin tool migration SQL.
- [ ] Buat migration awal untuk tabel domain yang telah disetujui.
- [ ] Buat indeks dan constraint.
- [ ] Uji migration up pada database kosong.
- [ ] Uji migration down dan up kembali.
- [ ] Pastikan migration dapat dijalankan berulang sesuai perilaku tool.

Branch yang disarankan: `feat/database-migration`.

## Fase 3 — Backend domain

Urutan implementasi setiap domain: model → repository → service → handler → routes → tests.

### 3.1 Teams

- [ ] Model dan validasi Team.
- [ ] Repository CRUD dan pencarian.
- [ ] Service aturan nama/identitas tim.
- [ ] Handler dan routes.
- [ ] Unit dan integration tests.

Branch: `feat/team-management`.

### 3.2 Matches dan jadwal

- [ ] Model Match dan status pertandingan.
- [ ] Repository daftar pertandingan berdasarkan tanggal, musim, dan tim.
- [ ] Service validasi pasangan tim, kickoff, dan hasil.
- [ ] Handler input/impor jadwal serta pembaruan hasil aktual.
- [ ] Pastikan pertandingan yang sudah dimulai/selesai tidak dapat diprediksi sebagai pertandingan mendatang.
- [ ] Tests.

Branch: `feat/match-management`.

### 3.3 Model version dan predictions

- [ ] Model metadata versi model.
- [ ] Repository penyimpanan prediksi dan snapshot fitur.
- [ ] Service validasi kelayakan prediksi.
- [ ] Simpan satu hasil per model yang digunakan untuk satu pertandingan dan versi model.
- [ ] Simpan `predicted_at`, kickoff, probabilitas H/D/A, predicted class, dan feature snapshot.
- [ ] Cegah prediksi ganda yang tidak disengaja atau definisikan perilaku pembuatan versi prediksi ulang.
- [ ] Tests.

Branch: `feat/prediction-management`.

### 3.4 Evaluation

- [ ] Ambil hanya prediksi yang dibuat sebelum kickoff.
- [ ] Ambil hanya pertandingan dengan hasil aktual yang tersedia.
- [ ] Hitung metrik dengan urutan kelas H/D/A yang konsisten.
- [ ] Bedakan evaluasi historis final test dari evaluasi prediksi live 2026/27.
- [ ] Tests.

Branch: `feat/model-evaluation`.

## Fase 4 — ML Service

- [ ] Siapkan FastAPI service.
- [ ] Muat artefak model yang sudah dibekukan dan metadata versinya.
- [ ] Validasi schema input dan urutan 12 fitur.
- [ ] Kembalikan probabilitas H/D/A dan versi model.
- [ ] Terapkan timeout dan penanganan error untuk pemanggil backend.
- [ ] Uji inference terhadap hasil notebook.
- [ ] Dokumentasikan kontrak request/response.

Branch: `feat/ml-service`.

## Fase 5 — Frontend

- [ ] Setup React + TypeScript.
- [ ] Dashboard ringkasan pertandingan dan prediksi.
- [ ] Daftar jadwal dan detail pertandingan.
- [ ] Form pemilihan tim atau pertandingan mendatang.
- [ ] Tampilan dua model dan probabilitas H/D/A.
- [ ] Tampilan riwayat pertandingan yang membentuk fitur.
- [ ] Halaman evaluasi model.
- [ ] Antarmuka pengelolaan data/admin sesuai keputusan scope.
- [ ] Validasi kondisi histori tidak mencukupi.

Branch disarankan per fitur frontend.

## Fase 6 — Integrasi, pengujian, dan deployment

- [ ] Integrasikan frontend, Go API, PostgreSQL, dan FastAPI melalui Docker Compose.
- [ ] Tambahkan konfigurasi health check untuk semua service.
- [ ] Uji alur end-to-end pembuatan prediksi.
- [ ] Jalankan Black Box Testing sesuai kebutuhan fungsional.
- [ ] Uji pencegahan prediksi setelah kickoff.
- [ ] Uji evaluasi live hanya untuk prediksi pra-kickoff dengan hasil aktual.
- [ ] Dokumentasikan setup lokal, environment, migration, training, dan deployment.
- [ ] Periksa secret, data, dan artefak sebelum repository dipublikasikan.

## Urutan terdekat yang harus dikerjakan

1. Masukkan dan review dua dokumen di folder `docs/`.
2. Audit serta finalisasi eksperimen Python sesuai spesifikasi.
3. Buat ERD dan data dictionary.
4. Setelah ERD disetujui, mulai SQL migration.
5. Lanjutkan model, repository, service, handler, routes, dan testing.

**Catatan:** fondasi Fiber, konfigurasi environment, PostgreSQL Docker, dan GORM yang sudah dibuat tidak perlu diulang. Roadmap ini menunda domain migration sampai kebutuhan data dan aturan prediksi cukup jelas.
