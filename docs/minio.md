Yes — here’s a practical starter setup.

A couple of notes first. MinIO’s current official docs emphasize container deployment and recommend TLS for network encryption in real environments. The MinIO Go SDK supports the usual S3-style flow: create a client, optionally create a bucket, upload with `PutObject`, and download with `GetObject`. For MySQL, `VARCHAR`, `BIGINT`, `JSON`, and temporal types are all standard choices for this kind of metadata table. ([MinIO AIStor Documentation][1])

## 1) Docker Compose to self-host MinIO

This is a simple single-node setup that is good for local development or a small internal service. A single-drive deployment does not provide erasure coding, so for production durability you would add replication, backups, and a stronger storage layout. ([MinIO AIStor Documentation][2])

Create `docker-compose.yml`:

```yaml
version: "3.9"

services:
  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    ports:
      - "9000:9000"   # S3 API
      - "9001:9001"   # MinIO Console
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data

volumes:
  minio_data:
```

Run it:

```bash
docker compose up -d
```

Endpoints:

* S3 API: `http://localhost:9000`
* Console: `http://localhost:9001`

For production, change the default credentials immediately and enable HTTPS/TLS. MinIO’s docs state TLS is the basis for secure network encryption. ([MinIO AIStor Documentation][3])

A matching `.env` style set of values you can use in your app:

```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_BUCKET=app-files
MINIO_USE_SSL=false
```

---

## 2) MySQL schema for MinIO object metadata

This schema stores the object identity, file metadata, integrity fields, and some app-level information. I’m using `DATETIME(6)` instead of `TIMESTAMP` to avoid the narrower `TIMESTAMP` range and to keep this future-proof. MySQL documents that `TIMESTAMP` has a more limited range than `DATETIME`. ([MySQL][4])

```sql
CREATE TABLE file_objects (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    bucket_name VARCHAR(100) NOT NULL,
    object_key VARCHAR(500) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    stored_filename VARCHAR(255) DEFAULT NULL,
    content_type VARCHAR(150) DEFAULT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    etag VARCHAR(64) DEFAULT NULL,
    sha256 CHAR(64) DEFAULT NULL,
    storage_class VARCHAR(50) DEFAULT NULL,
    metadata_json JSON DEFAULT NULL,
    uploaded_by_user_id BIGINT UNSIGNED DEFAULT NULL,
    is_public TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_bucket_object (bucket_name, object_key),
    KEY idx_uploaded_by_user_id (uploaded_by_user_id),
    KEY idx_created_at (created_at),
    KEY idx_original_filename (original_filename)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

Recommended meanings:

* `bucket_name`: MinIO bucket, such as `app-files`
* `object_key`: full object path, such as `users/42/2026/03/uuid-report.pdf`
* `original_filename`: name from the user upload
* `stored_filename`: sanitized name if you want to keep one separately
* `etag`: object ETag returned by MinIO after upload
* `sha256`: optional stronger checksum you calculate in your app
* `metadata_json`: custom metadata such as document type, tags, or tenant info

A common object key pattern is:

```text
{entity}/{entity_id}/{yyyy}/{mm}/{uuid}-{safe_filename}
```

Example:

```text
users/42/2026/03/550e8400-e29b-41d4-a716-446655440000-report.pdf
```

---

## 3) Sample Go flow: upload + save metadata + download

This example uses:

* MinIO Go SDK for object storage
* MySQL driver for metadata
* upload from a local file
* download to a local file

The MinIO Go SDK examples and API reference show the same core methods used here: `New`, `BucketExists`, `MakeBucket`, `PutObject`, and `GetObject`. ([GitHub][5])

### `go.mod`

```go
module example/minioapp

go 1.23

require (
    github.com/go-sql-driver/mysql v1.8.1
    github.com/google/uuid v1.6.0
    github.com/minio/minio-go/v7 v7.0.76
)
```

### `main.go`

```go
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	MySQLDSN string
}

type FileRecord struct {
	ID               uint64
	BucketName       string
	ObjectKey        string
	OriginalFilename string
	StoredFilename   string
	ContentType      string
	FileSize         int64
	ETag             string
	SHA256           string
	UploadedByUserID sql.NullInt64
	CreatedAt        time.Time
}

func main() {
	cfg := Config{
		MinioEndpoint:  env("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: env("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: env("MINIO_SECRET_KEY", "minioadmin123"),
		MinioBucket:    env("MINIO_BUCKET", "app-files"),
		MinioUseSSL:    env("MINIO_USE_SSL", "false") == "true",
		MySQLDSN:       env("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/appdb?parseTime=true&charset=utf8mb4"),
	}

	ctx := context.Background()

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalf("create minio client: %v", err)
	}

	if err := ensureBucket(ctx, minioClient, cfg.MinioBucket); err != nil {
		log.Fatalf("ensure bucket: %v", err)
	}

	// Example upload
	localSourcePath := "./sample.pdf"
	record, err := UploadFile(ctx, db, minioClient, cfg.MinioBucket, localSourcePath, 42)
	if err != nil {
		log.Fatalf("upload file: %v", err)
	}

	fmt.Printf("uploaded:\n")
	fmt.Printf("  id: %d\n", record.ID)
	fmt.Printf("  bucket: %s\n", record.BucketName)
	fmt.Printf("  object_key: %s\n", record.ObjectKey)
	fmt.Printf("  etag: %s\n", record.ETag)
	fmt.Printf("  sha256: %s\n", record.SHA256)

	// Example download
	targetPath := "./downloaded-" + record.StoredFilename
	if err := DownloadFile(ctx, minioClient, record.BucketName, record.ObjectKey, targetPath); err != nil {
		log.Fatalf("download file: %v", err)
	}

	fmt.Printf("downloaded to: %s\n", targetPath)
}

func UploadFile(
	ctx context.Context,
	db *sql.DB,
	minioClient *minio.Client,
	bucketName string,
	localPath string,
	uploadedByUserID int64,
) (*FileRecord, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat source file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("source path is a directory")
	}

	originalFilename := filepath.Base(localPath)
	safeName := sanitizeFilename(originalFilename)
	objectKey := buildObjectKey(uploadedByUserID, safeName)

	contentType := detectContentType(originalFilename)
	sha256Hex, err := hashFileSHA256(localPath)
	if err != nil {
		return nil, fmt.Errorf("sha256 file: %w", err)
	}

	// Reopen because hashing already read from disk separately.
	uploadFile, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("reopen source file: %w", err)
	}
	defer uploadFile.Close()

	opts := minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"original-filename": originalFilename,
			"sha256":            sha256Hex,
		},
	}

	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectKey, uploadFile, info.Size(), opts)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	record := &FileRecord{
		BucketName:       bucketName,
		ObjectKey:        objectKey,
		OriginalFilename: originalFilename,
		StoredFilename:   safeName,
		ContentType:      contentType,
		FileSize:         info.Size(),
		ETag:             uploadInfo.ETag,
		SHA256:           sha256Hex,
		UploadedByUserID: sql.NullInt64{Int64: uploadedByUserID, Valid: true},
	}

	res, err := db.ExecContext(ctx, `
		INSERT INTO file_objects
		(
			bucket_name,
			object_key,
			original_filename,
			stored_filename,
			content_type,
			file_size,
			etag,
			sha256,
			uploaded_by_user_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.BucketName,
		record.ObjectKey,
		record.OriginalFilename,
		record.StoredFilename,
		record.ContentType,
		record.FileSize,
		record.ETag,
		record.SHA256,
		record.UploadedByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert metadata: %w", err)
	}

	lastID, err := res.LastInsertId()
	if err == nil {
		record.ID = uint64(lastID)
	}

	return record, nil
}

func DownloadFile(
	ctx context.Context,
	minioClient *minio.Client,
	bucketName string,
	objectKey string,
	targetPath string,
) error {
	object, err := minioClient.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	defer object.Close()

	// Force a read early so "not found" and similar errors appear now.
	if _, err := object.Stat(); err != nil {
		return fmt.Errorf("stat remote object: %w", err)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, object); err != nil {
		return fmt.Errorf("copy object to file: %w", err)
	}

	return nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func buildObjectKey(userID int64, filename string) string {
	now := time.Now().UTC()
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	id := uuid.NewString()
	return fmt.Sprintf(
		"users/%d/%04d/%02d/%s-%s%s",
		userID,
		now.Year(),
		int(now.Month()),
		id,
		base,
		ext,
	)
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, name)

	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "application/octet-stream"
	}
	t := mime.TypeByExtension(ext)
	if t == "" {
		return "application/octet-stream"
	}
	return t
}

func hashFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
```

---

## Minimal SQL migration flow

Run your DB and then:

```sql
CREATE DATABASE appdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE appdb;

CREATE TABLE file_objects (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    bucket_name VARCHAR(100) NOT NULL,
    object_key VARCHAR(500) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    stored_filename VARCHAR(255) DEFAULT NULL,
    content_type VARCHAR(150) DEFAULT NULL,
    file_size BIGINT UNSIGNED NOT NULL,
    etag VARCHAR(64) DEFAULT NULL,
    sha256 CHAR(64) DEFAULT NULL,
    storage_class VARCHAR(50) DEFAULT NULL,
    metadata_json JSON DEFAULT NULL,
    uploaded_by_user_id BIGINT UNSIGNED DEFAULT NULL,
    is_public TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_bucket_object (bucket_name, object_key),
    KEY idx_uploaded_by_user_id (uploaded_by_user_id),
    KEY idx_created_at (created_at),
    KEY idx_original_filename (original_filename)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## How the flow works

Upload flow:

1. User sends file to your Go backend.
2. Backend generates an object key.
3. Backend uploads file bytes to MinIO.
4. Backend stores MinIO metadata in MySQL.
5. Backend returns your own file record ID.

Download flow:

1. Client requests file by your DB ID.
2. Backend loads `bucket_name` and `object_key` from MySQL.
3. Backend streams the object from MinIO to the client, or saves it locally.

A production app usually does not expose raw object keys directly unless it is deliberate. Instead, it uses your own file ID and authorization checks before download.

---

## Better production shape

For a more real deployment:

* use TLS on MinIO
* use a non-root MinIO user/policy for the app
* keep MySQL and MinIO on private networking
* store checksum and validate after upload
* back up both MySQL and MinIO data
* consider presigned URLs for browser downloads
* avoid using original filenames as object keys

MinIO’s docs emphasize TLS for secure transport, and the SDK examples show the client can be configured for secure or insecure connections depending on your endpoint. ([MinIO AIStor Documentation][3])

If you want, I can turn this into a small runnable Go API with endpoints like `POST /files` and `GET /files/:id`.

[1]: https://docs.min.io/enterprise/aistor-object-store/installation/container/?utm_source=chatgpt.com "Container | AIStor Object Store Documentation - MinIO"
[2]: https://docs.min.io/enterprise/aistor-object-store/reference/aistor-server/thresholds/?utm_source=chatgpt.com "Thresholds and Limits - AIStor Object Store - MinIO"
[3]: https://docs.min.io/enterprise/aistor-object-store/installation/linux/network-encryption/?utm_source=chatgpt.com "Enable Network Encryption | MinIO AIStor Documentation"
[4]: https://dev.mysql.com/doc/refman/8.4/en/datetime.html?utm_source=chatgpt.com "13.2.2 The DATE, DATETIME, and TIMESTAMP Types"
[5]: https://github.com/minio/minio-go?utm_source=chatgpt.com "MinIO Go client SDK for S3 compatible object storage"
