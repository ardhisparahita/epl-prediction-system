# Research Specification — EPL Prediction System

Dokumen ini menjadi acuan bersama untuk eksperimen machine learning, backend, ML Service, frontend, dan evaluasi sistem pada skripsi.

## 1. Identitas Penelitian

**Judul rekomendasi dosen:**

> Rancang Bangun Sistem Prediksi Hasil Pertandingan Liga Inggris Berbasis Web Menggunakan Random Forest dan Logistic Regression

**Objek penelitian:** pertandingan English Premier League (Liga Inggris).

**Sumber data:** Football-Data.co.uk, berkas kompetisi Premier League dengan kode `E0` untuk setiap musim.

**Cakupan musim:** 2018/2019 sampai 2026/2027, mengikuti ketersediaan dan kelengkapan data pada saat pengumpulan. Musim 2026/2027 merupakan musim berjalan dan bukan dataset uji historis yang dianggap telah lengkap.

## 2. Tujuan Sistem

Sistem menyediakan prediksi hasil pertandingan Liga Inggris dalam tiga kelas, membandingkan keluaran Random Forest dan Logistic Regression, menampilkan probabilitas setiap kelas, serta menyimpan prediksi agar dapat ditelusuri dan dievaluasi setelah hasil aktual tersedia.

Sistem juga menyediakan pengelolaan data pertandingan dan tim, informasi pertandingan historis yang digunakan untuk membentuk fitur, serta halaman evaluasi model.

## 3. Definisi Target Prediksi

Target adalah hasil akhir pertandingan pada waktu normal, dengan label internal:

| Label | Definisi | Tampilan di website |
|---|---|---|
| `H` | Gol tim kandang lebih besar daripada gol tim tandang | Kandang menang |
| `D` | Gol tim kandang sama dengan gol tim tandang | Seri |
| `A` | Gol tim tandang lebih besar daripada gol tim kandang | Tandang menang |

Label diturunkan dari skor akhir (`FTHG` dan `FTAG`) dan diperiksa konsistensinya dengan kolom hasil (`FTR`). Skor akhir dan statistik pertandingan yang sedang diprediksi tidak boleh menjadi fitur masukan model.

## 4. Dataset dan Pemeriksaan Data

### 4.1 Berkas sumber

Satu berkas CSV `E0` digunakan untuk setiap musim yang tersedia. Nama berkas mengikuti kode musim, misalnya `E0_1819.csv` dan `E0_2627.csv`.

### 4.2 Kolom yang diperlukan

Kolom minimal untuk penelitian:

- `Date`
- `Time` jika tersedia dan valid pada berkas sumber
- `HomeTeam`
- `AwayTeam`
- `FTHG`
- `FTAG`
- `FTR`
- `HST`
- `AST`

Kolom lain dapat dipertahankan untuk audit atau kebutuhan administrasi, tetapi tidak otomatis menjadi fitur model.

### 4.3 Pemeriksaan kualitas data

Pipeline data harus:

1. Memastikan identitas tim dan hasil pertandingan tersedia.
2. Memastikan skor kandang dan tandang tersedia untuk pertandingan yang digunakan sebagai data berlabel.
3. Memeriksa apakah `FTR` konsisten dengan `FTHG` dan `FTAG`.
4. Mengubah tanggal ke tipe tanggal dengan urutan hari-bulan-tahun sesuai format sumber.
5. Menggunakan tanggal dan waktu kickoff (`Date` dan `Time`) untuk pengurutan kronologis apabila waktu tersedia dan dapat dibaca.
6. Memeriksa duplikasi pertandingan, nilai kosong, dan konsistensi nama tim antarmusim.
7. Mencatat jumlah baris yang dibuang atau tidak dapat digunakan beserta alasannya.
8. Memeriksa ketersediaan `HST` dan `AST` pada setiap musim. Nilai kosong tidak boleh otomatis dianggap nol.

Jika `Time` tidak tersedia atau tidak valid, keterbatasan ketepatan waktu harus dicatat. Waktu kickoff tidak boleh dibuat-buat. Aturan pengurutan untuk pertandingan dengan tanggal yang sama harus diterapkan secara konsisten dan dapat diaudit.

## 5. Pembentukan Fitur

### 5.1 Keputusan yang telah ditetapkan

Keputusan pengguna untuk implementasi metodologi:

- Riwayat lima pertandingan dihitung dari **semua pertandingan tim**, baik ketika bermain kandang maupun tandang.
- Tim harus memiliki minimal **tiga pertandingan sebelumnya yang telah selesai** agar pertandingan target dapat diproses.
- Maksimal lima pertandingan sebelumnya digunakan. Jika tim baru memiliki tiga atau empat pertandingan sebelumnya, gunakan seluruh pertandingan yang tersedia (tiga atau empat).
- Fitur hanya memakai informasi yang telah tersedia sebelum kickoff pertandingan target.

### 5.2 Empat fitur dasar per tim

Untuk setiap tim dihitung:

1. **Rata-rata gol cetak (`gf_5`)**: rata-rata gol yang dicetak dalam maksimal lima pertandingan sebelumnya.
2. **Rata-rata gol kebobolan (`ga_5`)**: rata-rata gol yang diterima dalam maksimal lima pertandingan sebelumnya.
3. **Rata-rata poin (`pts_5`)**: rata-rata poin dalam maksimal lima pertandingan sebelumnya; menang = 3, seri = 1, kalah = 0.
4. **Rata-rata tembakan tepat sasaran (`sot_5`)**: rata-rata tembakan tepat sasaran dari pertandingan sebelumnya yang memiliki nilai SOT valid.

Untuk SOT, pembagi adalah jumlah pertandingan dengan data SOT valid, bukan selalu lima. Nilai SOT kosong diabaikan dalam perhitungan rata-rata dan tidak diubah menjadi nol.

### 5.3 Komposisi 12 fitur model

Setiap baris pertandingan memiliki empat fitur tim kandang, empat fitur tim tandang, dan empat fitur selisih (kandang dikurangi tandang):

| Urutan | Nama fitur | Definisi |
|---:|---|---|
| 1 | `H_gf_5` | Rata-rata gol cetak tim kandang |
| 2 | `H_ga_5` | Rata-rata gol kebobolan tim kandang |
| 3 | `H_pts_5` | Rata-rata poin tim kandang |
| 4 | `H_sot_5` | Rata-rata SOT valid tim kandang |
| 5 | `A_gf_5` | Rata-rata gol cetak tim tandang |
| 6 | `A_ga_5` | Rata-rata gol kebobolan tim tandang |
| 7 | `A_pts_5` | Rata-rata poin tim tandang |
| 8 | `A_sot_5` | Rata-rata SOT valid tim tandang |
| 9 | `diff_gf_5` | `H_gf_5 - A_gf_5` |
| 10 | `diff_ga_5` | `H_ga_5 - A_ga_5` |
| 11 | `diff_pts_5` | `H_pts_5 - A_pts_5` |
| 12 | `diff_sot_5` | `H_sot_5 - A_sot_5` |

Urutan dan nama fitur harus sama pada proses training, penyimpanan artefak, dan inference. Jumlah pertandingan pembentuk fitur (`n_prev`) dan jumlah SOT valid sebaiknya disimpan sebagai metadata/audit, bukan ditambahkan sebagai fitur model tanpa perubahan spesifikasi.

### 5.4 Pencegahan data leakage

- Riwayat tim harus diurutkan secara kronologis.
- Fitur pertandingan target dihitung dari pertandingan tim yang sudah selesai sebelum kickoff target.
- Implementasi rolling harus menggunakan operasi setara `shift(1)` sebelum rolling window.
- Statistik pertandingan target (`FTHG`, `FTAG`, `HST`, `AST`, dan statistik lain dari pertandingan target) hanya boleh digunakan untuk label atau evaluasi setelah pertandingan selesai, bukan sebagai fitur prediksi.
- Untuk prediksi berjalan, hanya hasil pertandingan yang telah selesai dan tersedia sebelum waktu prediksi yang boleh memperbarui riwayat tim.

### 5.5 Aturan kelayakan fitur dan nilai SOT yang tidak tersedia

Syarat minimal tiga pertandingan berlaku untuk jumlah pertandingan historis tim. SOT dihitung dari nilai valid dalam jendela maksimal lima pertandingan.

Jika tidak ada satu pun nilai SOT valid dalam jendela tersebut, fitur `sot_5` tidak dapat dihitung. Sampel training yang memiliki fitur wajib tidak lengkap harus dicatat dan tidak boleh diam-diam diisi nol. Untuk inference, sistem harus menolak atau memberi peringatan bahwa fitur historis belum mencukupi. Jumlah sampel yang terdampak harus dilaporkan.

## 6. Pembagian Data Berdasarkan Musim

| Kelompok | Musim | Fungsi |
|---|---|---|
| Training | 2018/2019–2023/2024 | Melatih kandidat model |
| Validation | 2024/2025 | Memilih konfigurasi model |
| Final test | 2025/2026 | Evaluasi final satu kali |
| Live prediction | 2026/2027 | Demonstrasi prediksi berjalan dan evaluasi setelah hasil aktual tersedia |

Pembagian tidak dilakukan secara acak. Riwayat pertandingan lintas musim boleh digunakan untuk membentuk fitur apabila pertandingan tersebut terjadi sebelum pertandingan target. Namun, label setiap musim tetap mengikuti kelompok data di atas.

## 7. Pelatihan dan Pemilihan Model

### 7.1 Algoritma

- Random Forest Classifier.
- Logistic Regression untuk klasifikasi tiga kelas.

Kedua algoritma harus menggunakan baris data, label, dan 12 fitur yang sama pada eksperimen komparatif.

### 7.2 Kandidat konfigurasi awal

Kandidat berikut mengikuti rancangan Bab III yang telah disusun dan perlu dikunci sebelum eksperimen final dijalankan:

**Random Forest**

- `n_estimators`: 200, 400
- `max_depth`: 3, 4, 6, 8
- `min_samples_leaf`: 5, 10, 20
- `class_weight`: `None`, `balanced`
- `random_state`: 42

**Logistic Regression**

- `C`: 0.01, 0.1, 1, 10
- `class_weight`: `None`, `balanced`
- `StandardScaler` diterapkan melalui Pipeline dan hanya di-fit pada data yang digunakan untuk melatih kandidat.
- `max_iter` ditetapkan cukup tinggi dan konsisten agar proses konvergensi dapat diperiksa.

Solver dan parameter teknis lain harus dicatat dalam konfigurasi eksperimen. Nilai kandidat tidak boleh diubah setelah hasil final test dilihat.

### 7.3 Prosedur pemilihan dan pengujian

1. Latih kandidat model dengan data training 2018/2019–2023/2024.
2. Evaluasi kandidat pada validation 2024/2025.
3. Pilih konfigurasi setiap algoritma berdasarkan Macro F1-score; Log Loss digunakan sebagai metrik pendukung untuk kualitas probabilitas.
4. Setelah konfigurasi ditetapkan, latih ulang masing-masing model menggunakan gabungan data training dan validation (2018/2019–2024/2025).
5. Evaluasi final satu kali menggunakan data 2025/2026.
6. Jangan mengubah fitur, parameter, atau keputusan model berdasarkan hasil final test.

Validation tidak boleh dilaporkan sebagai final test. Final test tidak boleh digunakan untuk memilih konfigurasi.

## 8. Baseline dan Metrik Evaluasi

### 8.1 Baseline

Baseline utama adalah **majority-class baseline**: selalu memprediksi kelas yang paling sering muncul pada data pelatihan yang relevan. Untuk evaluasi final, baseline harus dihitung dari data yang tersedia bagi proses pelatihan ulang final, yaitu training + validation, agar pembanding konsisten dengan model final test.

Jika Log Loss baseline turut dilaporkan, gunakan probabilitas konstan berdasarkan distribusi kelas pada data pelatihan yang relevan. Definisi ini harus dijelaskan terpisah dari baseline prediksi kelas mayoritas.

### 8.2 Metrik

- Accuracy.
- Macro Precision.
- Macro Recall.
- Macro F1-score.
- Confusion Matrix dengan urutan kelas `H`, `D`, `A`.
- Log Loss untuk menilai kualitas probabilitas.

Probabilitas harus dipetakan ke urutan kelas yang benar (`H`, `D`, `A`) sebelum disimpan atau ditampilkan. Evaluasi setiap model dilakukan pada set data yang sama.

Feature importance Random Forest dan koefisien Logistic Regression dapat digunakan sebagai analisis tambahan, bukan pengganti metrik evaluasi.

## 9. Model untuk Prediksi Berjalan 2026/2027

### 9.1 Keputusan pengguna: fixed model

Model untuk prediksi berjalan musim 2026/2027 **dibekukan** setelah dilatih menggunakan data sampai musim 2025/2026. Model ini tidak menjalani retraining selama periode evaluasi prediksi berjalan yang ditetapkan.

Model live merupakan artefak produksi tersendiri. Model ini berbeda dari model final-test yang dilatih menggunakan data sampai 2024/2025 dan diuji pada 2025/2026.

### 9.2 Fitur tetap diperbarui

Walaupun parameter dan artefak model dibekukan, fitur historis setiap tim tetap diperbarui setelah pertandingan selesai. Pembaruan fitur hanya menggunakan hasil pertandingan yang sudah selesai dan tersedia sebelum prediksi berikutnya dibuat.

Dengan demikian:

- model/parameter tetap;
- riwayat pertandingan dan fitur input dapat berubah;
- versi model yang digunakan setiap prediksi harus dicatat.

### 9.3 Pencatatan dan evaluasi live

Setiap prediksi live harus menyimpan setidaknya:

- pertandingan yang diprediksi;
- model dan versi model;
- waktu prediksi (`predicted_at`);
- waktu kickoff yang dijadwalkan;
- probabilitas `H`, `D`, dan `A`;
- kelas prediksi;
- snapshot 12 fitur yang dipakai;
- hasil aktual setelah pertandingan selesai.

Prediksi hanya masuk evaluasi live apabila dibuat sebelum kickoff dan hasil aktualnya sudah tersedia. Prediksi yang dibuat setelah kickoff atau pertandingan yang belum memiliki hasil aktual tidak boleh dimasukkan ke evaluasi tersebut. Pertandingan 2026/2027 yang telah selesai sebelum sistem mulai digunakan bukan prediksi live; data tersebut dapat menjadi riwayat fitur.

Tanggal cutoff pembekuan model, versi artefak, dan tanggal pengambilan dataset harus dicatat.

## 10. Kebutuhan Fungsional Sistem

Ruang lingkup fungsi minimum mengikuti review dosen:

1. Pemilihan tim kandang dan tim tandang.
2. Pemilihan jadwal pertandingan yang akan diprediksi atau input pertandingan yang belum berlangsung.
3. Pengambilan dan pembentukan fitur historis secara otomatis.
4. Prediksi dari Random Forest dan Logistic Regression.
5. Tampilan probabilitas H/D/A dari masing-masing model.
6. Informasi pertandingan sebelumnya yang digunakan untuk membentuk fitur.
7. Halaman evaluasi model.
8. Pengelolaan data serta penanganan kondisi ketika riwayat tim belum mencukupi.

Pengguna tidak diminta memasukkan statistik lima pertandingan sebelumnya secara manual.

## 11. Arsitektur Teknologi

| Komponen | Teknologi yang direncanakan | Tanggung jawab |
|---|---|---|
| Frontend | React + TypeScript | Antarmuka pengguna dan admin |
| Backend API | Go + Fiber v3 | REST API, validasi, business logic, pengelolaan data, dan komunikasi layanan |
| ORM | GORM | Akses database dari backend |
| Database | PostgreSQL | Menyimpan tim, pertandingan, prediksi, model/version metadata, dan data evaluasi yang diperlukan |
| ML Service | Python + FastAPI + scikit-learn | Memuat model, membentuk/validasi input fitur, menghasilkan probabilitas |
| Infrastructure | Docker Compose | Menjalankan layanan lokal secara terkoordinasi |

Pembentukan fitur pada training dan inference harus konsisten. Implementasi dapat berbagi modul Python agar rumus serta urutan fitur tidak diduplikasi secara berbeda.

## 12. Prinsip Desain Data untuk Database

Desain ERD dan migration dibuat setelah spesifikasi ini ditinjau. Entitas konseptual awal:

- **teams**: identitas klub dan status keaktifan per kebutuhan sistem;
- **matches**: tim kandang, tim tandang, kickoff, musim, skor aktual, hasil, dan status pertandingan;
- **model_versions**: algoritma, versi, tanggal training/cutoff data, skema fitur, dan identitas artefak;
- **predictions**: pertandingan, model/version, waktu prediksi, kelas prediksi, probabilitas H/D/A;
- **prediction feature snapshot**: 12 nilai fitur yang dipakai pada prediksi, dapat disimpan sebagai JSONB atau struktur ter-normalisasi setelah ERD diputuskan;
- **users/admin**: hanya jika autentikasi dan peran admin ditetapkan sebagai kebutuhan sistem.

Constraint yang perlu dipertimbangkan:

- satu pertandingan tidak boleh memiliki tim kandang dan tandang yang sama;
- pasangan pertandingan dan kickoff perlu divalidasi terhadap duplikasi;
- probabilitas H/D/A harus berada pada rentang 0–1 dan jumlahnya mendekati 1;
- prediksi harus terkait dengan versi model yang jelas;
- hasil aktual hanya diisi setelah pertandingan selesai dan berasal dari skor yang tervalidasi.

Entitas konseptual di atas belum merupakan keputusan final kolom maupun migration. ERD harus dibuat dan ditinjau sebelum SQL migration ditulis.

## 13. Pengembangan dan Pengujian Sistem

Sistem dikembangkan menggunakan metode Waterfall sesuai rancangan penelitian: analisis kebutuhan, perancangan, implementasi, dan pengujian.

Black Box Testing digunakan untuk menguji kesesuaian fungsi aplikasi dengan kebutuhan. Pengujian fungsional ini dilaporkan terpisah dari evaluasi kinerja model machine learning.

Pengujian tambahan yang relevan:

- validasi input tim dan jadwal;
- penolakan prediksi untuk pertandingan yang sudah dimulai/selesai;
- validasi minimal tiga pertandingan historis per tim;
- penanganan SOT yang tidak tersedia;
- kesesuaian urutan 12 fitur;
- probabilitas kedua model berjumlah mendekati 1;
- pencatatan waktu prediksi dan versi model;
- evaluasi live hanya untuk prediksi pra-kickoff yang hasil aktualnya sudah tersedia.

## 14. Keputusan yang Masih Perlu Dikunci

Hal-hal berikut belum ditetapkan secara eksplisit oleh keputusan pengguna atau review dosen, sehingga perlu diperiksa sebelum implementasi final:

1. **Timestamp kickoff:** apakah kolom `Time` tersedia dan cukup lengkap untuk semua musim; jika tidak, aturan untuk pertandingan bertanggal sama harus didokumentasikan.
2. **SOT tanpa nilai valid:** spesifikasi menghitung rata-rata dari nilai valid. Jika seluruh lima pertandingan sebelumnya tidak memiliki SOT valid, sistem tidak dapat menghasilkan fitur tersebut. Kebijakan akhir adalah menolak/mengeluarkan sampel dengan fitur wajib yang tidak tersedia, atau menetapkan kebijakan imputasi yang disetujui dan dilaporkan. Jangan mengisi nol secara otomatis.
3. **Ambang kelengkapan SOT:** minimal tiga pertandingan historis berlaku untuk jumlah pertandingan sebelumnya; jumlah minimum SOT valid di dalam jendela belum ditetapkan oleh review. Catat jumlah valid dan verifikasi distribusinya sebelum menetapkan aturan tambahan.
4. **Solver dan parameter teknis Logistic Regression:** tetapkan sebelum eksperimen final dan catat.
5. **Manajemen akses admin:** tentukan apakah login/role admin termasuk ruang lingkup fungsional dan apakah tabel users diperlukan.
6. **Sumber jadwal mendatang:** tetapkan apakah jadwal dimasukkan admin atau diimpor dari sumber data yang ditentukan.
7. **Timezone kickoff:** tetapkan standar penyimpanan waktu (disarankan timestamp konsisten dan konversi tampilan yang eksplisit); jangan mencampur waktu lokal tanpa informasi zona.

## 15. Referensi Acuan Internal

Spesifikasi ini disusun berdasarkan:

- Dokumen dosen: *Review Pengajuan Judul dan Abstraksi Awal Skripsi — Ardhis Parahita, 21 Oktober 2026*.
- Keputusan pengguna: lima pertandingan terakhir mencakup semua status kandang/tandang; minimal tiga pertandingan sebelumnya; model live 2026/2027 dibekukan dengan data sampai musim 2025/2026.
- Rancangan Bab I dan Bab III skripsi yang telah dibagikan sebelumnya.

Jika dokumen metodologi skripsi dan spesifikasi teknis berbeda, perubahan harus diselaraskan pada keduanya sebelum eksperimen final dan implementasi sistem dinyatakan selesai.
