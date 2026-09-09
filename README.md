# logger_stdout_library

Structured trace logger untuk Go service, terinspirasi dari konsep *span* ala OpenTelemetry. Library ini dirancang untuk mencatat satu flow proses utama (`MainLog`) beserta pemanggilan ke service/sistem eksternal di dalamnya (`NativeCall`), dengan output berupa JSON terstruktur yang siap dikonsumsi oleh log aggregator (Datadog, ELK, OpenSearch, dll).

## Daftar Isi

- [Fitur Utama](#fitur-utama)
- [Instalasi](#instalasi)
- [Konsep Dasar](#konsep-dasar)
- [Konfigurasi Output (.env)](#konfigurasi-output-env)
- [Penggunaan Dasar](#penggunaan-dasar)
- [Menangani Panic](#menangani-panic)
- [Distributed Trace Correlation](#distributed-trace-correlation)
- [Struktur JSON Output](#struktur-json-output)
- [Referensi API](#referensi-api)
- [Versioning](#versioning)

## Fitur Utama

- Structured JSON logging untuk setiap flow service.
- Nested logging untuk pemanggilan service eksternal (`NativeCall`) di dalam satu flow utama.
- Distributed trace correlation lewat `traceId` yang sama antar service, dengan `spanId` unik per service.
- Output fleksibel: `stdout` atau `file`, dikontrol lewat environment variable.
- Rotasi file log harian secara otomatis.
- Panic recovery bawaan (`SafeFinishOnPanic`) supaya log tetap tercatat walau terjadi error tak terduga.
- Aman digunakan secara concurrent (thread-safe).

## Instalasi

```bash
go get github.com/wmusermii/logger_stdout_library@v0.1.3
```

Lalu import di kode Go kamu:

```go
import logger "github.com/wmusermii/logger_stdout_library"
```

## Konsep Dasar

Library ini punya dua konsep utama:

| Konsep | Fungsi | Method |
|---|---|---|
| `MainLog` | Merepresentasikan satu flow proses utama (misal: satu request masuk ke service) | `BeginMain`, `UpdateMain`, `FinishMain` |
| `NativeCall` | Merepresentasikan satu pemanggilan ke service/sistem lain di dalam flow utama | `BeginNative`, `UpdateNative`, `FinishNative` |

Alur penggunaan yang umum:

1. Flow utama mulai berjalan → panggil `BeginMain`.
2. Di tengah flow, ada pemanggilan ke service lain → panggil `BeginNative`.
3. Pemanggilan service lain selesai (sukses/gagal) → panggil `FinishNative`.
4. Ada perubahan data di tengah flow utama → panggil `UpdateMain` (opsional).
5. Ulangi langkah 2–3 untuk setiap pemanggilan service lain berikutnya.
6. Flow utama selesai → panggil `FinishMain`. Ini titik dimana log benar-benar ditulis ke output (stdout/file).

## Konfigurasi Output (.env)

Konfigurasi dibaca dari environment variable berikut:

| Variable | Wajib | Deskripsi |
|---|---|---|
| `LOG_OUTPUT` | Tidak (`default: stdout`) | Tujuan output log. Nilai valid: `stdout` atau `file`. Bersifat *exclusive* — tidak bisa keduanya sekaligus. |
| `LOG_DIR` | Ya, jika `LOG_OUTPUT=file` | Path direktori tujuan file log. Bisa berupa folder lokal atau path ke *shared/mounted network drive* (NFS/SMB) yang sudah di-mount di server. |
| `LOG_SERVICE_NAME` | Ya, jika `LOG_OUTPUT=file` | Nama yang dipakai sebagai basis nama file log. |

Contoh isi file `.env`:

```env
LOG_OUTPUT=file
LOG_DIR=/mnt/shared-logs/order-service
LOG_SERVICE_NAME=order-service
```

Format nama file yang dihasilkan: `{LOG_SERVICE_NAME}-{YYYY-MM-DD}.log`, dan otomatis berganti file setiap hari (rotasi harian).

Library ini **wajib** diinisialisasi sekali di awal aplikasi sebelum `BeginMain` pertama kali dipakai:

```go
package main

import (
    "log"

    "github.com/joho/godotenv"
    logger "github.com/wmusermii/logger_stdout_library"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("file .env tidak ditemukan, pakai environment variable sistem")
    }

    if err := logger.Init(); err != nil {
        log.Fatalf("gagal inisialisasi logger: %v", err)
    }
    defer logger.Close()

    // ... jalankan service seperti biasa
}
```

Jika `logger.Init()` tidak dipanggil sama sekali, library tetap berjalan normal dengan output default ke `stdout` — tidak ada *breaking change* untuk service yang belum membutuhkan fitur file logging.

## Penggunaan Dasar

```go
package main

import (
    logger "github.com/wmusermii/logger_stdout_library"
)

func main() {
    // 1. Mulai flow utama
    m := logger.BeginMain(logger.BeginMainParams{
        SeverityText:     "INFO",
        EventName:        "ProcessOrder",
        EventCategory:    "business",
        ServiceName:      "order-service",
        ServiceVersion:   "1.0.0",
        OperationName:    "CreateOrder",
        OperationVersion: "v1",
        LocalIP:          "10.0.0.5",
        Channel:          "mobile-app",
        Layer:            "handler",
        ClientIP:         "203.0.113.10",
        Body: map[string]interface{}{
            "orderId": "ORD-001",
        },
    })

    // 2. Panggil service lain
    nc := m.BeginNative(logger.BeginNativeParams{
        Type: "HTTP",
        URL:  "https://inventory-service/api/check-stock",
    })

    // ... proses pemanggilan service lain di sini ...

    // 3. Pemanggilan service lain selesai
    nc.FinishNative("200", "OK")

    // 4. Update data di tengah flow (opsional)
    m.UpdateMain(logger.UpdateMainParams{
        Body: map[string]interface{}{
            "orderId": "ORD-001",
            "stock":   "available",
        },
    })

    // 6. Flow utama selesai, log ditulis ke output
    m.FinishMain("200", "order created successfully")
}
```

## Menangani Panic

Untuk memastikan log tetap tercatat walau terjadi *panic* (misal nil pointer dereference) di tengah flow, gunakan `SafeFinishOnPanic` lewat `defer` tepat setelah `BeginMain`:

```go
func processOrder() {
    m := logger.BeginMain(logger.BeginMainParams{
        EventName:   "ProcessOrder",
        ServiceName: "order-service",
    })
    defer m.SafeFinishOnPanic()

    // ... kode yang berpotensi panic ...
}
```

Jika panic terjadi, `SafeFinishOnPanic` akan otomatis memanggil `FinishMain` dengan `responseCode: "500"` dan pesan panic di `responseMessage`, lalu me-*re-throw* panic tersebut supaya tetap tertangkap oleh crash reporter atau monitoring lain di level yang lebih atas.

## Distributed Trace Correlation

Untuk menghubungkan log antar service dalam satu request yang sama, teruskan `traceId` yang sama ke setiap service, tapi biarkan `spanId` di-generate baru di masing-masing service:

```go
// Di service pertama (misal API Gateway)
traceID := logger.GenTraceID()

m := logger.BeginMain(logger.BeginMainParams{
    TraceID:     traceID, // fix traceId secara eksplisit
    ServiceName: "api-gateway",
    EventName:   "IncomingRequest",
})

// traceID diteruskan ke service berikutnya, misal lewat HTTP header:
// req.Header.Set("X-Trace-Id", traceID)
```

```go
// Di service kedua (misal Order Service), setelah menerima traceId dari header
receivedTraceID := "..." // diambil dari header X-Trace-Id

m := logger.BeginMain(logger.BeginMainParams{
    TraceID:     receivedTraceID, // reuse, JANGAN generate baru
    ServiceName: "order-service",
    EventName:   "ProcessOrder",
})
```

Dengan pola ini, semua log dari berbagai service dalam satu request akan memiliki `traceId` yang identik, sehingga bisa dikelompokkan menjadi satu kesatuan trace di log aggregator, sementara `spanId` tetap unik untuk membedakan setiap service.

## Struktur JSON Output

Setiap flow utama menghasilkan satu baris JSON dengan struktur berikut:

```json
{
  "traceId": "string",
  "spanId": "string",
  "severityText": "string",
  "severityNumber": "string",
  "timestamp": "string",
  "event": {
    "eventName": "string",
    "eventCategory": "string"
  },
  "body": "any",
  "resource": {
    "serviceName": "string",
    "serviceVersion": "string",
    "operationName": "string",
    "operationVersion": "string",
    "localIp": "string"
  },
  "attributes": {
    "channel": "string",
    "layer": "string",
    "clientIp": "string"
  },
  "responseCode": "string",
  "responseMessage": "string",
  "startTime": "string",
  "endTime": "string",
  "duration": "string",
  "nativeCalls": [
    {
      "type": "string",
      "URL": "string",
      "startTime": "string",
      "endTime": "string",
      "duration": 0,
      "responseCode": "string",
      "responseMessage": "string"
    }
  ]
}
```

`severityText` menerima salah satu dari nilai berikut: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, yang otomatis dipetakan ke `severityNumber` (1–6).

## Referensi API

### `Init() error`
Membaca konfigurasi dari environment variable (`LOG_OUTPUT`, `LOG_DIR`, `LOG_SERVICE_NAME`) dan menyiapkan output logger. Wajib dipanggil sekali di awal aplikasi.

### `InitWithConfig(cfg Config) error`
Sama seperti `Init`, tapi konfigurasi disuplai secara manual (berguna untuk unit test).

### `Close() error`
Menutup file handle yang sedang aktif (relevan untuk mode `file`). Sebaiknya dipanggil lewat `defer` di `main()`.

### `BeginMain(p BeginMainParams) *MainLog`
Memulai satu flow log baru. Jika `TraceID`/`SpanID` tidak diisi, akan di-generate otomatis.

### `(*MainLog) UpdateMain(p UpdateMainParams)`
Memperbarui data pada flow utama yang sedang berjalan. Hanya field yang diisi yang akan diperbarui.

### `(*MainLog) FinishMain(responseCode, responseMessage string)`
Menandai flow utama selesai dan menulis log ke output yang telah dikonfigurasi.

### `(*MainLog) SafeFinishOnPanic()`
Dipanggil lewat `defer` untuk menjamin log tetap tertulis walau terjadi panic di tengah flow.

### `(*MainLog) BeginNative(p BeginNativeParams) *NativeCallHandle`
Mencatat awal pemanggilan ke service/sistem eksternal.

### `(*NativeCallHandle) UpdateNative(url string)`
Memperbarui data pemanggilan service eksternal yang sedang berjalan.

### `(*NativeCallHandle) FinishNative(responseCode, responseMessage string)`
Menandai pemanggilan service eksternal selesai.

### `GenTraceID() string` / `GenSpanID() string`
Helper untuk generate `traceId`/`spanId` secara manual, berguna saat perlu mengontrol trace correlation antar service.

## Versioning

Library ini mengikuti [Semantic Versioning](https://semver.org/). Rilis tersedia dalam bentuk Git tag, misal:

```bash
go get github.com/wmusermii/logger_stdout_library@v0.1.3
```

Lihat daftar rilis lengkap di halaman **Releases**/**Tags** repository ini.

## Lisensi

Internal use — sesuaikan dengan kebijakan organisasi.