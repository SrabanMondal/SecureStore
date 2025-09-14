# SecureStore -- Secure File Storage & Sharing Backend

SecureStore is a production-ready backend service for secure file
storage and sharing. It combines PostgreSQL (metadata), MinIO/S3 (object
storage), AES-256-GCM encryption, and JWT-based authentication in a
modular, extensible design. The project is structured with clear
separation of concerns (handlers → services → repositories → utils),
supports presigned uploads/downloads, and includes background workers
for cleanup & reconciliation, ensuring fault-tolerant operation.

## 🚀 Features

### Authentication & Authorization

- JWT-based user sessions
- Bcrypt password hashing

### File Management

- Metadata in PostgreSQL, storage in MinIO
- Upload options: presigned (large files) or encrypted (AES-256-GCM)
- Download: redirect via presigned URL or decrypt & stream from backend
- Lifecycle states: pending, uploaded, deleting

### File Sharing

- Public share links with expiry
- Optional password protection (bcrypt-secured)
- Unified flow: direct download (unencrypted or presigned) or password-validated access

### Background Workers

- Cleanup of deleted files in MinIO + DB
- Reconcile pending uploads (ensure consistency between DB & storage)
- Auto-delete expired share links

### Extensibility

- Modular services for future plugins: virus scanning, compression, embedding, etc.
- Clear service/repository abstraction

## 📦 Architecture

```plain
    cmd/server/main.go          → entrypoint
    internal/
      config/                   → environment & config loader
      handlers/                 → HTTP route handlers (Auth, File, Share)
      services/                 → business logic (Auth, File, Share)
      repositories/             → PostgreSQL access via pgx
      utils/                    → encryption, logging, helpers
    migrations/                 → SQL migrations (Postgres schema)
```

- **PostgreSQL**: Source of truth for users, files, and share metadata
- **MinIO/S3**: Object store, no direct user access (all via presigned URLs or backend streaming)
- **AES-256-GCM**: Optional server-side encryption for private files
- **Echo**: Fast HTTP framework in Go

## 🛠 Setup & Installation

### Requirements

- Go 1.20+
- PostgreSQL
- MinIO (or any S3-compatible service)
- Docker (recommended for local dev)
- migrate CLI (for schema migrations)

### 1. Run dependencies via Docker

#### Postgres

```bash
docker run -d --name pg -p 5432:5432 \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=securestore \
  postgres:15
```

#### MinIO

```bash
docker run -d -p 9000:9000 -p 9001:9001 \
  --name minio \
  -e MINIO_ROOT_USER=admin \
  -e MINIO_ROOT_PASSWORD=12345678 \
  quay.io/minio/minio server /data --console-address ":9001"
```

### 2. Run migrations

```bash
migrate -path ./migrations -database "postgres://postgres:secret@localhost:5432/securestore?sslmode=disable" up
```

### 3. Configure environment

Create `.env` file:

```bash
APP_PORT=8080
JWT_SECRET=super_secret_jwt_key
DATABASE_URL=postgres://postgres:secret@localhost:5432/securestore?sslmode=disable
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=12345678
MINIO_USE_SSL=false
MINIO_BUCKET=files
FILE_ENC_KEY=YOUR_32_BYTE_KEY_HERE
```

### 4. Build & run

``` bash
go run ./cmd/server
```

## 📡 API Overview

### Authentication

`POST /api/register` -- Create account
`POST /api/login` -- Authenticate & get JWT

### File Management (requires JWT)

`POST /api/files/presigned` -- Generate presigned upload URL
`POST /api/files/:id/finalize` -- Mark upload as complete
`POST /api/files/encrypted` -- Encrypted upload via multipart
`GET /api/files/:id/download` -- Download (decrypt or presigned URL)
`GET /api/files` -- List user files
`DELETE /api/files/:id` -- Mark for deletion

### File Sharing Routes

- `POST /api/shares` -- Create share link (expiry + optional password)
- `GET /api/shares/:token` -- Access share (direct if no password)
- `POST /api/shares/:token/validate` -- Validate password & download
- `DELETE /api/shares/:id` -- Deletes shared link

## ⚙️ Background Jobs

- **CleanupDeletedFiles**: Permanently remove MinIO objects & DB rows
- **ReconcilePendingFiles**: Ensure DB matches MinIO uploads
- **DeleteExpiredShareLinks**: Purge expired shares

All run in independent goroutines with periodic execution.

## 🔐 Security Notes

- User passwords: Hashed with bcrypt
- Share passwords: Optional, stored as bcrypt hash
- Files: AES-256-GCM encryption (optional per upload)
- JWT secret: Required for all authenticated APIs
- Presigned URLs: Time-limited, controlled by backend

## 🔮 Future Enhancements

- Optional processing workers (triggered on upload or scheduled):
  - Virus scanning (ClamAV or external API)
  - Compression (zip/gzip)
  - Embedding generation (vector representation of files)
- Retrieval API:
  - `POST /search` with query → embed query → cosine similarity against stored vectors → return matching file IDs
  - Storage options: Postgres with pgvector, SQLite FTS, or
lightweight vector DB
- Advanced features:
  - Client-side encryption with WebCrypto
  - Per-user encryption keys with KMS integration
  - Audit logs & access tracking
  - Rate limiting & abuse prevention
