# API Documentation

## 1. Overview

Dokumentasi ini menjelaskan REST API pada EPL Prediction System.

Go Backend berfungsi sebagai API utama yang menerima request prediksi, mengambil data historis pertandingan dari PostgreSQL, dan meneruskan data ke Python ML Service.

### Base URL

Development:

`http://localhost:8080`

### Content Type

`application/json`

## 2. Available Endpoints

| Method | Endpoint              | Description                           |
| ------ | --------------------- | ------------------------------------- |
| GET    | `/api/v1/health`      | Memeriksa status API                  |
| GET    | `/api/v1/health/db`   | Memeriksa koneksi database            |
| POST   | `/api/v1/predictions` | Melakukan prediksi hasil pertandingan |

## 3. Prediction API

### Endpoint

`POST /api/v1/predictions`

Endpoint ini digunakan untuk memprediksi hasil pertandingan EPL menggunakan model Logistic Regression dan Random Forest.

Go Backend mengambil riwayat pertandingan secara otomatis dari PostgreSQL.

### Request Headers

| Header       | Value            |
| ------------ | ---------------- |
| Content-Type | application/json |

### Request Body

Pengguna hanya perlu mengirimkan tiga parameter:

| Field        | Type   | Required | Description                                   |
| ------------ | ------ | -------- | --------------------------------------------- |
| home_team    | string | Yes      | Nama tim kandang                              |
| away_team    | string | Yes      | Nama tim tandang                              |
| kickoff_date | string | Yes      | Tanggal pertandingan dengan format YYYY-MM-DD |

### Request Example

```json
{
  "home_team": "Arsenal",
  "away_team": "Chelsea",
  "kickoff_date": "2026-11-15"
}
```

### Processing Flow

Setelah menerima request, Go Backend melakukan proses berikut:

1. Memvalidasi nama tim dan tanggal pertandingan.
2. Mengambil maksimal lima pertandingan terakhir tim kandang dari PostgreSQL.
3. Mengambil maksimal lima pertandingan terakhir tim tandang dari PostgreSQL.
4. Memastikan masing-masing tim memiliki minimal tiga pertandingan sebelumnya.
5. Menghapus duplikasi pertandingan dari kedua riwayat.
6. Mengirimkan data historis ke Python ML Service.
7. Menerima hasil prediksi dari kedua model.
8. Mengembalikan hasil prediksi kepada pengguna.

Semua pertandingan yang digunakan harus memiliki tanggal sebelum tanggal kickoff.

### Successful Response

**HTTP Status:** `200 OK`

Example:

```json
{
  "home_team": "Arsenal",
  "away_team": "Chelsea",
  "kickoff_date": "2026-11-15",
  "feature_count": 12,
  "history_window": 5,
  "minimum_history_matches": 3,
  "probabilities": {
    "logistic_regression": {
      "H": 0.40889,
      "D": 0.334658,
      "A": 0.256452
    },
    "random_forest": {
      "H": 0.386105,
      "D": 0.313836,
      "A": 0.300059
    }
  },
  "model_version": "epl-result-live-2026-2027-v1"
}
```

### Response Fields

| Field                   | Type    | Description                            |
| ----------------------- | ------- | -------------------------------------- |
| home_team               | string  | Nama tim kandang                       |
| away_team               | string  | Nama tim tandang                       |
| kickoff_date            | string  | Tanggal pertandingan                   |
| feature_count           | integer | Jumlah fitur yang digunakan            |
| history_window          | integer | Maksimal pertandingan historis per tim |
| minimum_history_matches | integer | Minimal pertandingan historis per tim  |
| probabilities           | object  | Probabilitas hasil dari kedua model    |
| model_version           | string  | Versi model machine learning           |

### Probability Classes

| Class | Description |
| ----- | ----------- |
| H     | Home Win    |
| D     | Draw        |
| A     | Away Win    |

Nilai probabilitas berada dalam rentang 0 sampai 1. Semakin tinggi nilainya, semakin besar probabilitas kelas hasil pertandingan tersebut menurut model.

## 4. Error Responses

API menggunakan HTTP status code untuk menunjukkan kondisi request.

| HTTP Status               | Description                                                    |
| ------------------------- | -------------------------------------------------------------- |
| 400 Bad Request           | Format JSON tidak valid                                        |
| 422 Unprocessable Entity  | Input tidak valid atau riwayat pertandingan tidak mencukupi    |
| 500 Internal Server Error | Terjadi kegagalan saat mengambil data historis dari database   |
| 502 Bad Gateway           | Python ML Service tidak tersedia atau gagal memberikan respons |
| 503 Service Unavailable   | Prediction Service belum diinisialisasi                        |

### 4.1 Invalid JSON

**HTTP Status:** `400 Bad Request`

```json
{
  "status": "error",
  "message": "Invalid JSON request body"
}
```

### 4.2 Insufficient Match History

**HTTP Status:** `422 Unprocessable Entity`

```json
{
  "status": "error",
  "message": "insufficient match history for home team \"Unknown FC\": found 0, minimum required 3"
}
```

### 4.3 Invalid Kickoff Date

**HTTP Status:** `422 Unprocessable Entity`

```json
{
  "status": "error",
  "message": "kickoff_date must use YYYY-MM-DD format"
}
```

### 4.4 Database Error

**HTTP Status:** `500 Internal Server Error`

```json
{
  "status": "error",
  "message": "Failed to retrieve match history"
}
```

### 4.5 ML Service Unavailable

**HTTP Status:** `502 Bad Gateway`

```json
{
  "status": "error",
  "message": "ML prediction service is unavailable"
}
```

## 5. Python ML Service API

Python ML Service menyediakan endpoint internal yang digunakan oleh Go Backend.

### Health Check

`GET /health`

Digunakan untuk memeriksa status Python ML Service.

### Prediction Endpoint

`POST /api/v1/predictions`

Endpoint ini menerima data pertandingan dan riwayat pertandingan yang telah diperoleh oleh Go Backend.

Berbeda dengan endpoint publik Go Backend, request internal ML Service juga menyertakan `historical_matches`.

Go Backend bertanggung jawab menyediakan riwayat pertandingan tersebut sehingga pengguna tidak perlu mengirimkannya secara manual.
