#!/bin/bash
# Setup script for OM Messenger update system on server
# Run this ONCE on your server after deploying backend

set -e  # Exit on error

echo "🚀 OM Messenger Update System Setup"
echo "===================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Please run as root or with sudo"
    exit 1
fi

# Variables
DOWNLOADS_DIR="/var/www/api-om.wexun.tech/downloads"
BACKUP_DIR="/var/backups/om-messenger-apks"
NGINX_USER="www-data"

# Step 1: Create downloads directory
echo ""
echo "📁 Step 1: Creating downloads directory..."
mkdir -p "$DOWNLOADS_DIR"
mkdir -p "$BACKUP_DIR"

# Step 2: Set permissions
echo "🔐 Step 2: Setting permissions..."
chown -R $NGINX_USER:$NGINX_USER "$DOWNLOADS_DIR"
chmod 755 "$DOWNLOADS_DIR"
chmod 755 "$BACKUP_DIR"

# Step 3: Run database migration
echo "💾 Step 3: Running database migration..."
if command -v psql &> /dev/null; then
    echo "PostgreSQL found. Please run the migration manually:"
    echo "psql -U your_db_user -d om_messenger -f migrations/006_create_app_versions.sql"
else
    echo "⚠️  PostgreSQL CLI not found. Please run migration manually on database."
fi

# Step 4: Create test file
echo "📝 Step 4: Creating test file..."
echo "OM Messenger APK Downloads" > "$DOWNLOADS_DIR/README.txt"
chown $NGINX_USER:$NGINX_USER "$DOWNLOADS_DIR/README.txt"

# Step 5: Test nginx configuration
echo "🔍 Step 5: Checking nginx configuration..."
if command -v nginx &> /dev/null; then
    echo "Testing nginx config..."
    nginx -t
    if [ $? -eq 0 ]; then
        echo "✅ Nginx config is valid"
        echo "Run: systemctl reload nginx  (to apply changes)"
    else
        echo "❌ Nginx config has errors. Please fix before reloading."
    fi
else
    echo "⚠️  Nginx not found. Please configure manually."
fi

# Step 6: Create log directory
echo "📋 Step 6: Setting up logs..."
touch /var/log/nginx/downloads-access.log
touch /var/log/nginx/downloads-error.log
chown $NGINX_USER:adm /var/log/nginx/downloads-*.log

# Summary
echo ""
echo "✅ Setup Complete!"
echo "=================="
echo ""
echo "📁 Downloads directory: $DOWNLOADS_DIR"
echo "💾 Backup directory: $BACKUP_DIR"
echo ""
echo "🔧 Next steps:"
echo "1. Update nginx config with provided configuration"
echo "2. Run database migration: psql -U user -d om_messenger -f migrations/006_create_app_versions.sql"
echo "3. Reload nginx: systemctl reload nginx"
echo "4. Test endpoint: curl https://api-om.wexun.tech/api/version?platform=android"
echo "5. Upload first APK: scp app-release.apk user@server:$DOWNLOADS_DIR/om-messenger-v1.0.0.apk"
echo ""
echo "📖 Full documentation: APK_DEPLOYMENT_WORKFLOW.md"
