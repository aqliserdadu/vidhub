# Video Downloader - API Documentation

## Base URL
```
http://localhost:8080/api
```

## Endpoints

### 1. Get Video Information

Fetch metadata about a video from a given URL.

**Endpoint:**
```http
GET /api/video/info?url=<video_url>
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| url | string | Yes | Valid video URL from supported platform |

**Example Request:**
```bash
curl "http://localhost:8080/api/video/info?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

**Success Response (200 OK):**
```json
{
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "title": "Rick Astley - Never Gonna Give You Up (Official Video)",
  "duration": 213,
  "thumbnail_url": "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
  "uploader": "Rick Astley Official",
  "formats": [
    {
      "format_id": "137",
      "format": "1920x1080 - ....",
      "ext": "mp4",
      "resolution": "1920x1080",
      "video_codec": "h264",
      "audio_codec": "aac",
      "file_size": 52428800,
      "fps": 30,
      "quality": "FHD",
      "official_name": "FHD (1920x1080) - h264 + aac"
    },
    {
      "format_id": "22",
      "format": "1280x720 - ....",
      "ext": "mp4",
      "resolution": "1280x720",
      "video_codec": "h264",
      "audio_codec": "aac",
      "file_size": 31457280,
      "fps": 30,
      "quality": "HD",
      "official_name": "HD (1280x720) - h264 + aac"
    },
    {
      "format_id": "251",
      "format": "tiny default audio",
      "ext": "webm",
      "resolution": "",
      "video_codec": "none",
      "audio_codec": "opus",
      "file_size": 5242880,
      "fps": 0,
      "quality": "Audio",
      "official_name": "Audio - none + opus"
    }
  ]
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "invalid_domain",
  "message": "URL domain is not allowed",
  "code": 400
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "error": "fetch_failed",
  "message": "Failed to fetch video information: ...",
  "code": 500
}
```

---

### 2. Start Download

Initiate a video download with a specific format.

**Endpoint:**
```http
POST /api/download
Content-Type: application/json
```

**Request Body:**
```json
{
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "format_id": "137"
}
```

**Request Parameters:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| url | string | Yes | Video URL |
| format_id | string | Yes | Format ID from video info endpoint |

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format_id": "137"
  }'
```

**Success Response (200 OK):**
```json
{
  "id": "1702910400000000000",
  "title": "Rick_Astley_Never_Gonna_Give_You_Up.mp4",
  "download_link": "/api/download/1702910400000000000",
  "expires_at": 1702996800
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "file_too_large",
  "message": "File size exceeds maximum limit of 300MB",
  "code": 400
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "error": "download_failed",
  "message": "Download failed: ...",
  "code": 500
}
```

---

### 3. Download File

Stream the downloaded video file.

**Endpoint:**
```http
GET /api/download/:id
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Download ID from start download response |

**Example Request:**
```bash
curl -O http://localhost:8080/api/download/1702910400000000000
```

**Success Response (200 OK):**
- Binary file stream with `Content-Disposition: attachment` header
- File will be downloaded to client

**Error Response (404 Not Found):**
```json
{
  "error": "not_found",
  "message": "File not found or has expired",
  "code": 404
}
```

---

### 4. Health Check

Check if the service is running and healthy.

**Endpoint:**
```http
GET /api/health
```

**Success Response (200 OK):**
```json
{
  "status": "healthy",
  "service": "video-downloader"
}
```

---

## Rate Limiting

- **Limit per IP**: 30 requests per minute
- **Burst allowance**: 60 requests
- **Reset period**: 1 minute

Exceeding limits will result in HTTP 429 response.

## Time Zones

All timestamps are in Unix Epoch (seconds since 1970-01-01 00:00:00 UTC).

**Example:** `expires_at: 1702996800` = 2023-12-19 20:00:00 UTC

## Supported Formats Quality

The API returns formats in these quality categories:

| Quality | Description | Example |
|---------|-------------|---------|
| FHD | Full HD (1920x1080) | 1080p video + audio |
| HD | HD (1280x720) | 720p video + audio |
| 4K | 4K (3840x2160) | Blocked by server (max 300MB) |
| Audio | Audio only | MP3, AAC, Opus |
| 480p | Standard (854x480) | 480p video + audio |
| 360p | Low (640x360) | 360p video + audio |

## File Availability

- **TTL (Time to Live)**: 24 hours
- **Storage location**: `./downloads/`
- **Automatic cleanup**: Expired files are deleted automatically
- **Max file size**: 300MB

Files are automatically deleted after 24 hours of creation.

## Error Codes

| Code | Name | Description |
|------|------|-------------|
| 400 | Bad Request | Invalid request format or parameters |
| 404 | Not Found | File not found or expired |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error during processing |
| 513 | Service Unavailable | Python worker unreachable |

## Supported Domains

By default, these domains are supported:
- `youtube.com`
- `youtu.be`
- `vimeo.com`

To add more domains, modify the `ALLOWED_DOMAINS` environment variable.

## Examples with cURL

### Get video info
```bash
curl "http://localhost:8080/api/video/info?url=https://www.youtube.com/watch?v=..." | jq .
```

### Download video
```bash
# Step 1: Get formats
FORMATS=$(curl -s "http://localhost:8080/api/video/info?url=https://www.youtube.com/watch?v=...")
FORMAT_ID=$(echo $FORMATS | jq -r '.formats[0].format_id')

# Step 2: Start download
RESPONSE=$(curl -s -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"https://www.youtube.com/watch?v=...\",\"format_id\":\"$FORMAT_ID\"}")

DOWNLOAD_ID=$(echo $RESPONSE | jq -r '.id')

# Step 3: Download file
curl -O "http://localhost:8080/api/download/$DOWNLOAD_ID"
```

### Check health
```bash
curl http://localhost:8080/api/health | jq .
```

---

## Response Headers

All responses include:
- `Content-Type`: `application/json` or `application/octet-stream`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: SAMEORIGIN`

## Request Timeout

- **Video Info**: 60 seconds
- **Download**: 300 seconds (5 minutes)
- **Read Timeout**: 300 seconds

---

## Changelog

### Version 1.0.0 (2026-02-06)
- Initial release
- Support for multiple video platforms
- Quality selection (FHD, HD, Audio)
- File size and 4K restrictions
- 24-hour TTL with auto cleanup
- Production-ready logging and error handling
