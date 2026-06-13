#!/bin/bash
set -e

# Change to the project root directory (where this script's grandparent is)
cd "$(dirname "$0")/../.."

echo "🚀 Starting Local Dev Swarm Environment..."

# 1. Initialize Docker Swarm if not already active
if ! docker info | grep -q "Swarm: active"; then
    echo "⚙️ Initializing Docker Swarm..."
    docker swarm init
else
    echo "✅ Docker Swarm is already active."
fi

# 2. Create the .env file with dummy credentials if it doesn't exist
if [ ! -f "deploy/dev/.env" ]; then
    echo "📝 Creating deploy/dev/.env with default development credentials..."
    cat <<EOF > deploy/dev/.env
# Development Credentials
PORT=8080
ALLOWED_ORIGINS=http://localhost:3000
PUBLIC_API_BASE_URL=http://localhost:8082/api
PUBLIC_JOIN_BASE_URL=http://localhost:8082

DB_HOST=postgres
DB_PORT=5432
DB_USER=dev_postgres
DB_PASSWORD=dev_postgres_pass
DB_NAME=om_messenger
DB_SSLMODE=disable

JWT_SECRET=dev_secret_key

MINIO_ROOT_USER=dev_minio
MINIO_ROOT_PASSWORD=dev_minio_pass_123
MINIO_ENDPOINT=minio:9000
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=om-avatars

S3_ENDPOINT=minio:9000
S3_USE_SSL=false
S3_BUCKET=om-avatars
S3_ACCESS_KEY=dev_minio
S3_SECRET_KEY=dev_minio_pass_123

REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
EOF
fi

# 3. Build the backend image locally
echo "🏗️ Building backend image locally (om-backend-dev:latest)..."
docker build --progress=plain -t om-backend-dev:latest -f Dockerfile .

# 4. Resolve variables and deploy
echo "📦 Exporting variables and deploying stack..."
export $(cat deploy/dev/.env | grep -v '^#' | xargs)
docker stack deploy -c deploy/dev/docker-compose.yml om-dev

# 5. Wait a moment and check status
sleep 3
echo "✨ Deployment complete! Here are the running services:"
docker service ls --filter name=om-dev

echo "
🎉 Dev Environment is fully ready!
- Backend API: http://localhost:8082
- PostgreSQL: localhost:5432 (dev_postgres/dev_postgres_pass)
- MinIO S3 API: http://localhost:9000 (dev_minio/dev_minio_pass_123)
- MinIO Web Console: http://localhost:9001
- Redis: localhost:6379
"
