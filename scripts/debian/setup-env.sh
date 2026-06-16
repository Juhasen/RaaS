#!/bin/bash
set -e

# Get repo root relative to scripts/debian directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
REPO_ROOT="$(dirname "$(dirname "$DIR")")"
cd "$REPO_ROOT"

echo "============================================="
echo "   RaaS Environment Configuration Setup"
echo "============================================="
echo ""

# 1. Notification Service (SMTP)
echo "--- Notification Service (SMTP Configuration) ---"
read -p "SMTP Host [smtp.gmail.com]: " smtp_host
smtp_host=${smtp_host:-smtp.gmail.com}

read -p "SMTP Port [587]: " smtp_port
smtp_port=${smtp_port:-587}

read -p "SMTP User (Gmail Address): " smtp_user
read -sp "SMTP Password (App Password): " smtp_password
echo ""
echo ""

# 2. Media Service (Cloudflare R2 / S3)
echo "--- Media Service (Cloudflare R2 / S3 Configuration) ---"
read -p "R2 Account ID: " r2_account_id
read -p "R2 Access Key ID: " r2_access_key_id
read -sp "R2 Secret Access Key: " r2_secret_access_key
echo ""

read -p "R2 Bucket Name [raas-media-bucket]: " r2_bucket_name
r2_bucket_name=${r2_bucket_name:-raas-media-bucket}

read -p "R2 Public URL (e.g. https://pub-xxx.r2.dev): " r2_public_url

echo ""
echo "Creating .env files..."

# Write notification/.env
cat << EOF > notification/.env
SMTP_HOST=$smtp_host
SMTP_PORT=$smtp_port
SMTP_USER=$smtp_user
SMTP_PASSWORD=$smtp_password
EOF
echo "Created notification/.env"

# Write media/.env
cat << EOF > media/.env
# Cloudflare R2 / S3 Configuration for Media Service
R2_ACCOUNT_ID=$r2_account_id
R2_ACCESS_KEY_ID=$r2_access_key_id
R2_SECRET_ACCESS_KEY=$r2_secret_access_key
R2_BUCKET_NAME=$r2_bucket_name
R2_PUBLIC_URL=$r2_public_url
EOF
echo "Created media/.env"

echo ""
echo "Configuration complete!"
echo "============================================="
