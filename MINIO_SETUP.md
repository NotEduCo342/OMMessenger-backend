# MinIO Setup Guide

## Configuration Overview

MinIO is now configured in your `docker-compose.yml` for S3-compatible object storage.

### Security Features ✅
- **Local network only**: MinIO is NOT exposed to the internet
- **Strong credentials**: Set via `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` in `.env`
- **5GB storage quota**: Limited to prevent abuse
- **Docker network isolation**: Only backend container can access MinIO

### Access Details

**Internal (from backend container):**
- S3 API: `http://minio:9000`
- Backend will use this endpoint

**Local management (optional):**
To access the MinIO web console from your machine, uncomment these lines in `docker-compose.yml`:
```yaml
ports:
  - "127.0.0.1:9000:9000"   # S3 API
  - "127.0.0.1:9001:9001"   # Web Console
```
Then visit: `http://localhost:9001` (login with your MINIO_ROOT_USER/PASSWORD)

### Required Environment Variables

Add these to your `.env` file (see `.env.example`):

```bash
# MinIO Configuration
MINIO_ROOT_USER=om_minio_admin
MINIO_ROOT_PASSWORD=your_strong_password_at_least_32_chars_recommended
MINIO_ENDPOINT=minio:9000
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=om-avatars
```

**Important**: Change `MINIO_ROOT_PASSWORD` to a strong random password!

### Starting MinIO

```bash
# Start all services including MinIO
docker-compose up -d

# Check MinIO is running
docker-compose ps

# View MinIO logs
docker-compose logs minio

# Access MinIO shell (for bucket creation)
docker-compose exec minio sh
```

### Creating the Avatars Bucket

After starting MinIO, create the bucket for storing avatars:

```bash
# Method 1: Using MinIO Client (mc) inside container
docker-compose exec minio sh -c "
  mc alias set local http://localhost:9000 \$MINIO_ROOT_USER \$MINIO_ROOT_PASSWORD
  mc mb local/om-avatars --ignore-existing
  mc anonymous set download local/om-avatars
"

# Method 2: Using web console (if ports are exposed)
# Visit http://localhost:9001, login, and create bucket 'om-avatars'
```

### Storage Location

Avatar files are stored in: `./minio-storage/` (gitignored)

### Security Notes

1. **Never expose MinIO ports publicly** — keep them localhost-only or remove the ports entirely
2. **Change default credentials** — use a password manager to generate strong passwords
3. **Backup strategy** — `minio-storage/` directory contains all uploaded files
4. **5GB limit** — enforced at docker volume level; adjust in `docker-compose.yml` if needed

### Next Steps

Once MinIO is configured and running, the backend Go code will be updated to:
- Accept avatar uploads via `POST /users/me/avatar`
- Validate images (format, size)
- Store in MinIO
- Return public URL
- Update user profile with avatar URL
