#  VidHub - Video Downloader & Media Archiver

**Platform universal untuk mengarsipkan dan mengkonversi media dari berbagai sumber streaming**

---

##  Daftar Isi

- [Deskripsi Proyek](#deskripsi-proyek)
- [Fitur Utama](#fitur-utama)
- [Teknologi yang Digunakan](#teknologi-yang-digunakan)
- [Struktur Proyek](#struktur-proyek)
- [Instalasi & Setup](#instalasi--setup)
- [Cara Menggunakan](#cara-menggunakan)
- [API & Endpoints](#api--endpoints)
- [Fungsi Utama & Modul](#fungsi-utama--modul)
- [Konfigurasi](#konfigurasi)
- [Tips & Catatan Penting](#tips--catatan-penting)
- [Troubleshooting](#troubleshooting)

---

##  Deskripsi Proyek

### Tujuan Proyek
VidHub adalah aplikasi web untuk mengunduh, mengkonversi, dan mengarsipkan video dari berbagai platform streaming (YouTube, Facebook, TikTok, Instagram, Twitter, Vimeo, dll.) ke format MP4, WEBM, atau MP3.

### Masalah yang Diselesaikan
1. **Arsip Media Offline** - Menyimpan konten video/audio penting untuk akses offline
2. **Konversi Format** - Mengubah format video ke berbagai resolusi dan format audio
3. **Edukasi & Presentasi** - Menyediakan tool untuk guru dan siswa mengunduh materi pembelajaran
4. **Personal Backup** - Membackup konten kreasi sendiri atau referensi penting
5. **Privacy & Security** - Proses on-premise tanpa mengirim data ke server eksternal

### Use Cases
-  **Pendidik**: Menyiapkan materi pembelajaran offline untuk kelas
-  **Siswa**: Mengunduh tutorial dan referensi pelajaran
-  **Content Creator**: Membackup karya sendiri
-  **Peneliti**: Mengarsipkan sumber referensi video

---

##  Fitur Utama

### 1. **Multi-Platform Support**
- YouTube & YouTube Music
- Facebook & Facebook Watch
- TikTok
- Instagram
- Twitter/X
- Vimeo

### 2. **Multiple Output Formats**
- **Video**: MP4, WEBM, MKV
- **Audio**: MP3, M4A, WAV
- **Kualitas**: FHD (1080p), HD (720p), SD (480p), Audio Only

### 3. **Sistem Pembersihan Otomatis**
- File TTL (Time To Live) yang dapat dikonfigurasi
- Cleanup routine yang berjalan di background
- Default: File dihapus setelah 24 jam
- Configurable interval: `STORAGE_CLEANUP_INTERVAL` & `FILE_TTL_SECONDS`

### 4. **Rate Limiting & Quota Protection**
- **Rate Limiting**: Cegah abuse dengan limit permintaan per menit
- **Download Quota**: Batasi total download per IP per hari
- **Reset Schedule**: Quota reset otomatis setiap hari pada jam tertentu

### 5. **Security Features**
- Domain whitelist untuk mencegah download dari platform tidak disetujui
- Request timeout otomatis
- File size validation
- Filename length restriction

### 6. **User-Friendly Interface**
- Modern glassmorphism UI design
- Real-time format filtering
- Metadata preview (judul, durasi, uploader)
- Error messages yang user-friendly
- Loading indicators & progress feedback

---

##  Teknologi yang Digunakan

### Backend
- **Language**: Go 1.21
- **Framework**: Gin Web Framework
- **Logging**: Uber Zap Logger
- **Config Management**: Environment Variables

### Frontend
- **HTML5 & CSS3**: Modern responsive design
- **JavaScript (Vanilla)**: No framework, lightweight
- **UI Library**: Bootstrap Icons
- **Alert/Dialog**: SweetAlert2

### Media Processing
- **Python 3.9+**: Media conversion engine
- **yt-dlp**: Video metadata & download (replacement untuk youtube-dl)
- **FFmpeg**: Audio/video conversion
- **Flask**: Python worker API

### DevOps & Deployment
- **Docker**: Containerization
- **Docker Compose**: Multi-service orchestration
- **Nginx**: Reverse proxy (optional)
- **Systemd**: Service management

---

##  Struktur Proyek

```
videoDownload/
│
├── backend/                      # Go Backend Server
│   ├── main.go                  # Entry point
│   ├── go.mod                   # Go dependencies
│   ├── config/                  # Configuration loading
│   │   └── config.go           # Config struct & env loader
│   ├── main.go         # Server entry point
│   ├── internal/
│   │   ├── handler/            # HTTP request handlers
│   │   │   ├── video_handler.go      # GET /api/video/info
│   │   │   └── download_handler.go   # POST /api/download, GET /api/download/:id
│   │   ├── service/            # Business logic
│   │   │   ├── video_service.go      # Video metadata extraction
│   │   │   ├── download_service.go   # Download orchestration
│   │   │   ├── quota_service.go      # Quota tracking per IP
│   │   │   └── ratelimit_service.go  # Rate limiting
│   │   ├── storage/            # File storage management
│   │   │   └── manager.go      # File tracking & auto-cleanup
│   │   └── model/              # Data structures
│   │       ├── models.go       # Domain models
│   │       └── config.go       # Config models
│   ├── pkg/                    # Shared packages
│   │   ├── logger/            # Zap logger wrapper
│   │   ├── middleware/        # HTTP middleware
│   │   └── validator/         # Input validation
│   └── server                  # Compiled binary
│
├── frontend/                    # Web Interface
│   └── index.html              # Single-page application
│
├── python-worker/              # Python Media Processing Service
│   ├── worker.py              # Flask app for media operations
│   ├── requirements.txt        # Python dependencies
│   └── downloads/             # Temp media files
│
├── docker-compose.yml          # Multi-service orchestration
├── Dockerfile                  # Backend container
├── Dockerfile.python           # Python worker container
├── nginx.conf                  # Reverse proxy config (optional)
│
├── downloads/                  # Downloaded files storage
├── log/                        # Application logs
│
└── Documentation/              # Project documentation
    ├── API.md                 # REST API documentation
    └── QUICKSTART.md          # Quick start guide

```

### Deskripsi Folder Utama

| Folder | Fungsi |
|--------|--------|
| **backend/** | Go REST API server, business logic, file management |
| **frontend/** | Web UI (single HTML file), user interface |
| **python-worker/** | Media processing service menggunakan yt-dlp & FFmpeg |
| **downloads/** | Temporary storage untuk file yang diunduh |
| **log/** | Application logs (Go backend & Python worker) |

---

##  Instalasi & Setup

### Persyaratan Sistem

**Hardware:**
- RAM minimal: 2GB (4GB recommended)
- Storage: Minimal 10GB (atau lebih sesuai kebutuhan)
- CPU: 2 cores (4 cores recommended)

**Software:**
- Docker 20.10+ & Docker Compose 2.0+ (untuk deployment)
- Git (untuk clone project)

### Opsi 1: Docker Compose (Recommended)


```bash
# 1. Clone project
git clone <repo-url>
cd videoDownload

# 2. build dan jalankan container
docker-compose build --no-cache 
docker-compose up -d

# 3. cek logs container
docker logs video-downloader-backend
docker logs video-downloader-worker

# 4. Access aplikasi
# Frontend: http://localhost:8080
# Backend API: http://localhost:8080/api
```

**Ports yang di-expose:**
- `8080` - Frontend & Backend API
- `5000` - Python Worker (internal only, tidak exposed)

---

### Opsi 2: Local Development Setup

**Persyaratan:**
```bash
# Install Go 1.21
# Download dari: https://golang.org/dl/

# Install Python 3.9+
sudo apt-get update
sudo apt-get install python3 python3-pip python3-venv

# Install FFmpeg (untuk audio conversion)
sudo apt-get install ffmpeg

# Install sistem dependencies lainnya
sudo apt-get install build-essential libssl-dev libffi-dev python3-dev
```

**Setup Backend:**
```bash
cd backend

# Download dependencies
go mod download
go mod tidy

# Build (optional)
go build -o server main.go

# Run
go run main.go
# atau
./server

# Server akan listen di http://localhost:8080
```

**Setup Python Worker (Terminal berbeda):**
```bash
cd python-worker

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Run worker
python worker.py
# atau
python3 -u worker.py

# Worker akan listen di http://localhost:5000
```


##  Cara Menggunakan

### 1. Cara Pakai Aplikasi (User)

**Step 1: Buka aplikasi**
```
1. Buka browser ke http://localhost:8080
2. Interface akan tampil dengan input form
```

**Step 2: Input URL media**
```
1. Cari video yang ingin diunduh (YouTube, TikTok, dll)
2. Copy link video
3. Paste di form "Tempel tautan video untuk diarsip..."
4. Atau klik button Clipboard icon untuk paste dari clipboard
```

**Step 3: Proses metadata**
```
1. Klik button "Proses Media"
2. Tunggu beberapa detik hingga metadata dimuat
3. Sistem akan tampilkan:
   - Thumbnail video
   - Judul dan durasi
   - Uploader/sumber
```

**Step 4: Pilih format**
```
1. Gunakan quality filter:
   - "Semua Format" - tampilkan semua format tersedia
   - "High Quality (1080p)" - hanya format FHD
   - "Standar (720p)" - hanya format HD
   - "Audio Only (MP3)" - hanya audio

2. Klik format kartu yang ingin diunduh
   - Kartu akan highlight dengan border biru
   - Check mark akan muncul di corner
```

**Step 5: Download**
```
1. Klik button "Unduh ke Perangkat"
2. Sistem akan memproses (tunggu status dialog)
3. File akan otomatis download ke perangkat Anda
4. Notifikasi sukses akan muncul
```

---

### 2. Contoh Input & Output

**Contoh 1: Download Video YouTube**

| Aspek | Nilai |
|-------|-------|
| **Input URL** | https://www.youtube.com/watch?v=dQw4w9WgXcQ |
| **Quality Filter** | HD (720p) |
| **Selected Format** | H.264 Video + AAC Audio (500 MB) |
| **Output** | Rick_Astley_Never_Gonna_Give_You_Up_720p.mp4 |
| **Time Taken** | ~1-2 menit |

**Contoh 2: Download Audio dari Musik**

| Aspek | Nilai |
|-------|-------|
| **Input URL** | https://www.youtube.com/watch?v=music_id |
| **Quality Filter** | Audio Only (MP3) |
| **Selected Format** | MP3 Audio 192kbps (5 MB) |
| **Output** | Song_Title_Audio.mp3 |
| **Time Taken** | ~30 detik |

---

### 3. Command Line / API Usage

**Get Video Information (Query Metadata):**
```bash
curl -X GET "http://localhost:8080/api/video/info?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Response (JSON):
{
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "title": "Rick Astley - Never Gonna Give You Up",
  "duration": 213,
  "thumbnail_url": "https://...",
  "uploader": "Rick Astley",
  "formats": [
    {
      "format_id": "18",
      "ext": "mp4",
      "resolution": "480p",
      "video_codec": "h264",
      "audio_codec": "aac",
      "file_size": 50000000,
      "quality": "SD"
    },
    // ... format lainnya
  ]
}
```

**Download Video:**
```bash
curl -X POST "http://localhost:8080/api/download" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format_id": "18",
    "quality": "SD",
    "file_size": 50000000
  }'

# Response (JSON):
{
  "id": "1707407648000000000",
  "title": "Rick_Astley_Never_Gonna_Give_You_Up.mp4",
  "download_link": "/api/download/1707407648000000000",
  "expires_at": 1707494048
}
```

**Download File:**
```bash
# Browser akan otomatis download file
curl -X GET "http://localhost:8080/api/download/1707407648000000000" \
  -o downloaded_video.mp4
```

**Health Check:**
```bash
curl -X GET "http://localhost:8080/api/health"

# Response:
{"status":"healthy"}
```

---

## 🔌 API & Endpoints

### REST API Endpoints

#### 1. **GET /api/health**
**Deskripsi**: Health check endpoint

```
Method: GET
URL: http://localhost:8080/api/health
Response Status: 200 OK
Response Body: {"status":"healthy"}
```

---

#### 2. **GET /api/video/info**
**Deskripsi**: Ambil metadata video (judul, durasi, available formats)

```
Method: GET
URL: http://localhost:8080/api/video/info?url=<URL_ENCODED_VIDEO_URL>
Query Parameters:
  - url (required): URL video yang sudah di-encode
Response Status: 200 OK
Response Body:
{
  "url": "string",
  "title": "string",
  "duration": 123,
  "thumbnail_url": "string",
  "uploader": "string",
  "formats": [
    {
      "format_id": "string",
      "ext": "string",
      "resolution": "string",
      "file_size": 123000,
      "quality": "FHD|HD|SD|Audio"
    }
  ]
}

Error Response (400):
{
  "error": "invalid_domain",
  "message": "URL domain tidak didukung",
  "code": 400
}

Error Response (500):
{
  "error": "download_failed",
  "message": "Gagal mengambil metadata",
  "code": 500
}
```

---

#### 3. **POST /api/download**
**Deskripsi**: Download & konversi video ke format yang dipilih

```
Method: POST
URL: http://localhost:8080/api/download
Content-Type: application/json
Request Body:
{
  "url": "string (required)",
  "format_id": "string (required)",
  "quality": "string (optional)",
  "file_size": 123000 (optional)
}

Response Status: 200 OK
Response Body:
{
  "id": "string (download ID)",
  "title": "string (filename)",
  "download_link": "string (/api/download/{id})",
  "expires_at": 1707494048 (unix timestamp)
}

Error Response (413 - File Too Large):
{
  "error": "file_too_large",
  "message": "File melebihi limit 100MB",
  "code": 413
}

Error Response (402 - Quota Exceeded):
{
  "error": "quota_exhausted",
  "message": "Kuota harian sudah habis",
  "code": 402
}

Error Response (429 - Too Many Requests):
{
  "error": "rate_limit_exceeded",
  "message": "Terlalu banyak permintaan",
  "code": 429
}
```

---

#### 4. **GET /api/download/:id**
**Deskripsi**: Download file yang sudah diproses

```
Method: GET
URL: http://localhost:8080/api/download/{download_id}
Path Parameters:
  - id (required): Download ID dari response POST /api/download
Response Status: 200 OK
Response Body: Binary file data
Headers:
  - Content-Disposition: attachment; filename="filename.mp4"
  - Content-Type: application/octet-stream

Error Response (404):
{
  "error": "not_found",
  "message": "File tidak ditemukan atau sudah expired",
  "code": 404
}
```

---

### Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | Success | Video diunduh berhasil |
| 400 | Bad Request | URL invalid atau format tidak sesuai |
| 402 | Payment Required | Quota harian sudah habis |
| 404 | Not Found | File expired atau tidak ada |
| 413 | Payload Too Large | File melampaui size limit |
| 429 | Too Many Requests | Rate limit terlampaui |
| 500 | Server Error | Kesalahan server atau processing |

---

## 🔧 Fungsi Utama & Modul

### Backend (Go)

#### 1. **VideoHandler** (`internal/handler/video_handler.go`)
**Fungsi**: Menangani request untuk metadata video

**Methods:**
- `GetVideoInfo(c *gin.Context)` - Query metadata dari URL
- `HealthCheck(c *gin.Context)` - Health check endpoint

**Flow:**
```
User Input URL
    ↓
VideoHandler.GetVideoInfo()
    ↓
VideoService.GetVideoInfo() (calls Python worker)
    ↓
Returns: title, duration, thumbnail, formats
```

---

#### 2. **DownloadHandler** (`internal/handler/download_handler.go`)
**Fungsi**: Menangani download & serving file

**Methods:**
- `StartDownload(c *gin.Context)` - Initiate download process
- `GetFile(c *gin.Context)` - Serve downloaded file

**Features:**
- Validate file size
- Check download quota
- Apply rate limiting
- Track file expiry

---

#### 3. **VideoService** (`internal/service/video_service.go`)
**Fungsi**: Komunikasi dengan Python worker untuk metadata

**Methods:**
- `GetVideoInfo(url string)` - Fetch metadata dari Python worker
- `CheckConnection()` - Verify worker availability

**Implementation:**
```go
// Call Python worker REST API
http.Get("http://python-worker:5000/api/video/info?url=...")
```

---

#### 4. **DownloadService** (`internal/service/download_service.go`)
**Fungsi**: Orchestration proses download & file management

**Methods:**
- `Download(req *DownloadRequest)` - Main download logic
- `GetDownloadFile(id string)` - Retrieve file info
- `GetFileSize(id string)` - Get file size for quota tracking

**Process:**
```
1. Validate URL format
2. Call Python worker /download endpoint
3. Read response stream
4. Save file to disk
5. Track in storage manager
6. Return download link
```

---

#### 5. **StorageManager** (`internal/storage/manager.go`)
**Fungsi**: File storage management & automatic cleanup

**Methods:**
- `SaveFile(id, file)` - Register file untuk di-track
- `GetFile(id)` - Retrieve tracked file metadata
- `EnsureDownloadDir()` - Create download directory
- `GetTrackedFilesCount()` - Get monitoring info
- `GetExpiredFilesCount()` - Get cleanup status
- `ManualCleanup()` - Force cleanup (testing)

**Features:**
-  Automatic cleanup dengan TTL
-  Background cleanup routine
-  Safe logging with nil checks
-  Non-blocking stop signal
-  Proper error handling

**Cleanup Flow:**
```
File downloaded & saved
    ↓
SaveFile() → Set ExpiresAt = now + FileTTLSeconds
    ↓
cleanupRoutine() runs every STORAGE_CLEANUP_INTERVAL
    ↓
For each tracked file:
  IF now > ExpiresAt:
    → os.Remove(FilePath)
    → delete from tracking map
    ↓
File deleted from disk & memory
```

---

#### 6. **QuotaService** (`internal/service/quota_service.go`)
**Fungsi**: Track & enforce download quota per IP

**Methods:**
- `CheckQuota(ip, extraBytes)` - Check if IP has quota available
- `AddUsage(ip, sizeMB)` - Record download usage
- `GetQuotaInfo(ip)` - Get current quota status
- `ResetQuotas()` - Reset expired quotas

**Features:**
- Per-IP daily quota tracking
- Configurable reset time (default: midnight)
- In-memory storage (fast)
- Automatic cleanup for old entries

---

#### 7. **RateLimitService** (`internal/service/ratelimit_service.go`)
**Fungsi**: Prevent abuse dengan rate limiting

**Methods:**
- `IsAllowed(ip)` - Check if request allowed
- `CheckTicket(ip)` - Get remaining requests
- `Reset()` - Reset counters

**Configuration:**
```
RATELIMIT_ENABLED: true
RATELIMIT_REQUESTS_PER_MINUTE: 60
RATELIMIT_BURST_SIZE: 10
RATELIMIT_CLEANUP_INTERVAL: 1800 (seconds)
```

---

### Python Worker

#### 1. **VideoInfoEndpoint** (`/api/video/info`)
**Fungsi**: Extract video metadata menggunakan yt-dlp

```python
@app.route('/api/video/info', methods=['GET'])
def get_video_info():
    url = request.args.get('url')
    # Use yt-dlp to extract:
    # - video title
    # - duration
    # - thumbnail
    # - available formats
    # - uploader
    return formats_json
```

---

#### 2. **DownloadEndpoint** (`/api/download`)
**Fungsi**: Download & convert video ke format yang dipilih

```python
@app.route('/api/download', methods=['POST'])
def download_video():
    data = request.get_json()
    url = data['url']
    format_id = data['format_id']
    
    # Use yt-dlp to download
    # Save to DOWNLOAD_DIR
    # Return file via send_file()
    return send_file(filepath)
```

---

#### 3. **Filename Truncation**
**Fungsi**: Ensure filenames tidak terlalu panjang

```python
def truncate_filename(filename, max_len=200):
    """Truncate filename ke MAX_FILENAME_LENGTH sambil preserve extension"""
    base, ext = os.path.splitext(filename)
    if len(filename) > max_len:
        available_len = max_len - len(ext)
        base = base[:available_len]
        filename = base + ext
    return filename
```

---

### Frontend

#### 1. **VidHubTool Class** (`index.html` - JavaScript)
**Fungsi**: Main frontend controller

**Key Methods:**
```javascript
// Initialize elements & event listeners
initializeElements()
attachEventListeners()

// Fetch video metadata
async fetchVideoInfo()

// Display video info
displayVideoInfo()

// Render format cards
renderFormats()
getUniqueFormats()

// Filter formats by quality
filterFormats(quality)
updateDownloadButtonState()

// Handle download
async startDownload()
getUserFriendlyErrorMessage(status, data)

// Utility functions
formatDuration(seconds)
formatBytes(bytes)
isValidURL(url)
```

---

##  Konfigurasi

### Environment Variables

#### Backend (Go)

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `SERVER_PORT` | 8080 | Port server listen |
| `SERVER_HOST` | 0.0.0.0 | Host server bind (0.0.0.0 = all interfaces) |
| `SERVER_TIMEOUT` | 300 | Request timeout (seconds) |
| `DOWNLOAD_DIR` | ./downloads | Folder download files |
| `MAX_VIDEO_SIZE_MB` | 100 | Max file size (MB) |
| `MAX_FILENAME_LENGTH` | 200 | Max filename length (chars) |
| `STORAGE_CLEANUP_INTERVAL` | 3600 | Cleanup interval (seconds) = 1 jam |
| `FILE_TTL_SECONDS` | 86400 | File expire time (seconds) = 24 jam |
| `PYTHON_WORKER_HOST` | python-worker | Python worker hostname |
| `PYTHON_WORKER_PORT` | 5000 | Python worker port |
| `PYTHON_WORKER_TIMEOUT` | 60 | Api call timeout (seconds) |
| `LOG_LEVEL` | info | Log level: debug, info, warn, error |
| `LOG_FILE` | ./log/app.log | Log file path |
| `ALLOWED_DOMAINS` | youtube.com,youtu.be,... | Allowed video domains |
| `QUOTA_ENABLED` | true | Enable quota limiting |
| `QUOTA_DAILY_LIMIT_MB` | 100 | Daily quota per IP (MB) |
| `QUOTA_RESET_HOUR` | 0 | Quota reset hour (0-23) |
| `QUOTA_RESET_MINUTE` | 0 | Quota reset minute (0-59) |
| `RATELIMIT_ENABLED` | true | Enable rate limiting |
| `RATELIMIT_REQUESTS_PER_MINUTE` | 60 | Max requests per minute |
| `RATELIMIT_BURST_SIZE` | 10 | Burst allowance |
| `RATELIMIT_CLEANUP_INTERVAL` | 1800 | Cleanup interval (seconds) |

#### Python Worker

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PYTHON_WORKER_PORT` | 5000 | Port worker listen |
| `PYTHON_WORKER_HOST` | 0.0.0.0 | Host worker bind |
| `DOWNLOAD_DIR` | ./downloads | Folder temporary files |
| `MAX_VIDEO_SIZE_MB` | 300 | Max file size (MB) |
| `MAX_FILENAME_LENGTH` | 200 | Max filename length |
| `LOG_LEVEL` | INFO | Log level: DEBUG, INFO, WARNING, ERROR |
| `ALLOWED_DOMAINS` | youtube.com,youtu.be,... | Allowed domains |

### Contoh .env File

```bash
# Backend Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_TIMEOUT=300

# Storage & Cleanup
DOWNLOAD_DIR=/app/downloads
MAX_VIDEO_SIZE_MB=100
MAX_FILENAME_LENGTH=200
STORAGE_CLEANUP_INTERVAL=3600      # Cleanup per 1 jam
FILE_TTL_SECONDS=86400             # Expire setelah 24 jam

# Python Worker
PYTHON_WORKER_HOST=python-worker
PYTHON_WORKER_PORT=5000
PYTHON_WORKER_TIMEOUT=60

# Security
ALLOWED_DOMAINS=youtube.com,youtu.be,vimeo.com,facebook.com,m.facebook.com,fb.watch,tiktok.com,instagram.com,twitter.com,x.com
REQUEST_TIMEOUT=60

# Quota (Pembatasan)
QUOTA_ENABLED=true
QUOTA_DAILY_LIMIT_MB=100           # 100MB per IP per hari
QUOTA_RESET_HOUR=0                 # Reset jam 00:00
QUOTA_RESET_MINUTE=0

# Rate Limiting (Anti-abuse)
RATELIMIT_ENABLED=true
RATELIMIT_REQUESTS_PER_MINUTE=60   # Max 60 requests/menit
RATELIMIT_BURST_SIZE=10
RATELIMIT_CLEANUP_INTERVAL=1800

# Logging
LOG_LEVEL=info
LOG_FILE=/app/log/app.log
LOG_ROTATION_SIZE=104857600
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=14
```

### Konfigurasi untuk Different Scenarios

**Development Environment:**
```bash
STORAGE_CLEANUP_INTERVAL=60        # Cleanup per menit (testing)
FILE_TTL_SECONDS=300               # Expire setelah 5 menit
LOG_LEVEL=debug
QUOTA_ENABLED=false
RATELIMIT_ENABLED=false
```

**Production Environment:**
```bash
STORAGE_CLEANUP_INTERVAL=3600      # Cleanup per jam
FILE_TTL_SECONDS=86400             # Expire setelah 1 hari
LOG_LEVEL=info
QUOTA_ENABLED=true
QUOTA_DAILY_LIMIT_MB=500           # Lebih generous
RATELIMIT_ENABLED=true
RATELIMIT_REQUESTS_PER_MINUTE=100
```

**High-Security Environment:**
```bash
STORAGE_CLEANUP_INTERVAL=1800      # Cleanup tiap 30 menit
FILE_TTL_SECONDS=3600              # Expire setelah 1 jam
MAX_VIDEO_SIZE_MB=50               # Lebih ketat
QUOTA_ENABLED=true
QUOTA_DAILY_LIMIT_MB=50            # Sangat terbatas
RATELIMIT_ENABLED=true
RATELIMIT_REQUESTS_PER_MINUTE=30   # Ketat
```

---

##  Tips & Catatan Penting

### 1. **File Cleanup System**

 **Bagaimana Cleanup Bekerja:**
- File yang diunduh di-track dalam memory map
- Setiap file punya `ExpiresAt` timestamp
- Background routine cleanup berjalan setiap `STORAGE_CLEANUP_INTERVAL` detik
- File expired akan dihapus otomatis dari disk

 **Important Notes:**
- Cleanup adalah **in-memory**, bukan persistent database
- Jika server restart, tracking info hilang (file baru akan dibuat)
- File lama di disk tidak otomatis di-cleanup tanpa restart
- Untuk production, pertimbangkan backup strategy untuk files penting

**Recommended Values:**
- Development: `FILE_TTL_SECONDS=300` (5 menit)
- Production: `FILE_TTL_SECONDS=86400` (1 hari)
- High-volume: `FILE_TTL_SECONDS=3600` (1 jam)

---

### 2. **Quota & Rate Limiting**

 **Rate Limiting:**
- Cegah abuse & DDoS attacks
- Per-IP basis (berdasarkan IP address)
- Jika melalui proxy, pastikan `X-Forwarded-For` header properly set
- Default 60 requests/menit dengan burst size 10

 **Download Quota:**
- Batasi total volume download per IP per hari
- Reset otomatis setiap hari pada `QUOTA_RESET_HOUR:QUOTA_RESET_MINUTE`
- Cocok untuk public/shared service
- Default 100MB per IP per hari

 **Considerations:**
- Jika server di belakang proxy, IP detection bisa salah
- Shared WiFi/office akan membagi quota satu IP
- Audio-only downloads lebih hemat kuota dibanding video

---

### 3. **Error Handling**

 **Frontend Error Messages:**
User-friendly error messages untuk berbagai kasus:
- **Quota penuh**: "Kuota unduhan harian sudah terpenuhi..."
- **File terlalu besar**: "File media ini terlalu besar (max 100MB)..."
- **Rate limit**: "Terlalu banyak permintaan dalam waktu singkat..."
- **Platform tidak didukung**: "Sumber media ini tidak didukung..."

Semua error diterima sebagai `title` dan `text` untuk user clarity.

---

### 4. **Backend Module Architecture**

```
Handler (menerima request)
    ↓
Service (business logic)
    ↓
Repository/Manager (data access)
    ↓
Middleware (apply policies)
```

Clean separation of concerns memudahkan maintenance & testing.

---

### 5. **Security Best Practices**

 **Domain Whitelisting:**
```
ALLOWED_DOMAINS=youtube.com,youtu.be,vimeo.com,...
```
Hanya domain ini yang bisa di-download.

 **File Size Validation:**
```
MAX_VIDEO_SIZE_MB=100
```
Reject files yang terlalu besar.

 **Filename Sanitization:**
- Max length: `MAX_FILENAME_LENGTH=200`
- Remove invalid characters
- Prevent path traversal

 **Request Timeouts:**
```
REQUEST_TIMEOUT=60 (seconds)
SERVER_TIMEOUT=300 (seconds)
```
Prevent hanging requests.

---

### 6. **Performance Optimization**

**Recommendations:**
1. Monitor disk space regularly
2. Tune `STORAGE_CLEANUP_INTERVAL` based on volume
3. Use SSD untuk download folder (faster I/O)
4. Monitor memory usage (in-memory tracking)
5. Consider rate limiting berbasis bandwidth

**Monitoring Metrics:**
- Total tracked files: `StorageManager.GetTrackedFilesCount()`
- Expired files waiting cleanup: `StorageManager.GetExpiredFilesCount()`
- API response time
- Worker availability

---

### 7. **Common Issues & Solutions**

| Masalah | Penyebab | Solusi |
|---------|----------|--------|
| Quota error padahal aman | IP Forwarding salah | Set X-Forwarded-For header |
| Download timeout | Network/File size besar | Increase `SERVER_TIMEOUT` |
| "Source not supported" | Domain tidak di-whitelist | Add ke `ALLOWED_DOMAINS` |
| Python worker error | FFmpeg tidak installed | `apt-get install ffmpeg` |

---

##  Troubleshooting

### Backend Tidak Bisa Connect ke Python Worker

**Error Message:**
```
Failed to contact Python worker
error: connection refused
```

**Solutions:**
```bash
# 1. Check Python worker status
docker-compose ps video-downloader-worker

# 2. Check logs
docker logs video-downloader-worker

# 3. Verify port 5000 is open
netstat -tulpn | grep 5000

# 4. Test connection
curl http://localhost:5000/api/health

# 5. Restart both services
docker-compose restart python-worker backend
```

---

### Disk Space Penuh

**Symptoms:**
```
Failed to write file
No space left on device
```

**Solutions:**
```bash
# 1. Check disk usage
df -h

# 2. Manual cleanup old files
cd downloads/
find . -mtime +7 -delete  # Delete files older than 7 days

# 3. Fine-tune cleanup settings
STORAGE_CLEANUP_INTERVAL=1800      # More frequent
FILE_TTL_SECONDS=3600               # Shorter TTL

# 4. Monitor ongoing
watch -n 5 'du -sh downloads/'
```

---

### Memory Usage Tinggi

**Cause**: In-memory file tracking banyak

**Solutions:**
```bash
# 1. Check tracked files
curl http://localhost:8080/debug/storage  # (if exposed)

# 2. Force cleanup
POST /admin/cleanup endpoint (if available)

# 3. Restart to reset memory
docker-compose restart backend

# 4. Reduce cleanup interval
STORAGE_CLEANUP_INTERVAL=1800       # More aggressive
```

---

### Rate Limit/Quota Error Padahal User Baru

**Cause**: IP detection salah (proxy/shared network)

**Solutions:**
```bash
# 1. Check IP detection in logs
curl -v http://localhost:8080/api/health

# 2. If behind proxy, ensure:
# - X-Forwarded-For header properly set
# - Nginx/Reverse proxy configured correctly

# 3. Disable quota for testing
QUOTA_ENABLED=false
RATELIMIT_ENABLED=false

# 4. Increase limits
QUOTA_DAILY_LIMIT_MB=500
RATELIMIT_REQUESTS_PER_MINUTE=200
```

---

### "Sumber Media Tidak Didukung"

**Cause**: Domain tidak di-whitelist

**Solutions:**
```bash
# 1. Check allowed domains
echo $ALLOWED_DOMAINS

# 2. Add new domain
ALLOWED_DOMAINS=youtube.com,youtu.be,tiktok.com,newsite.com

# 3. Restart backend
docker-compose restart backend

# 4. Test again
```

---

##  Monitoring & Logging

### Check Logs

```bash
# Backend logs
docker logs -f video-downloader-backend

# Python worker logs
docker logs -f video-downloader-worker

# View specific log level
docker logs --tail 100 video-downloader-backend | grep ERROR

# Save logs to file
docker logs video-downloader-backend > backend.log 2>&1
```

### Health Check

```bash
# Check both services
curl http://localhost:8080/api/health

# Is backend running?
curl -v http://localhost:8080/api/health

# Is Python worker reachable?
curl -v http://localhost:5000/api/health (internal)
```

---

##  License & Credit

**Project**: VidHub - Video Downloader & Media Archiver
**Version**: 1.0.0
**Created**: February 2026

**Technologies Used:**
- Go (Backend)
- Python (Media Processing)
- JavaScript (Frontend)
- yt-dlp (Video extraction)
- FFmpeg (Format conversion)

---

##  Contributing & Support

### Feature Requests
- Buat issue dengan label `enhancement`
- Jelaskan use case & expected behavior

### Bug Reports
- Sertakan error message & logs
- Jelaskan steps untuk reproduce
- Mention environment (OS, Docker version, etc)

### Questions
- Check documentation first
- Review existing issues
- Ask in discussion forum

---

##  Additional Resources

- [API.md](API.md) - Detailed API documentation

---

**Last Updated**: February 8, 2026  
**Status**: Production Ready ✅  
**Documentation Version**: 1.0.0
