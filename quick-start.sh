#!/bin/bash

set -e

echo "🚀 Mini Cloud - S3 Integration Quick Start"
echo "=========================================="
echo ""

if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

echo "1️⃣  Starting MinIO..."
docker-compose up -d
echo "✅ MinIO started!"
echo ""

echo "2️⃣  MinIO Web Console:"
echo "   🌐 http://localhost:9001"
echo "   👤 Login: minioadmin"
echo "   🔑 Password: minioadmin"
echo ""

echo "3️⃣  Downloading Go dependencies..."
go mod download
go mod tidy
echo "✅ Dependencies downloaded!"
echo ""

echo "4️⃣  Ready to start the application!"
echo "   Run: make dev        (with S3)"
echo "   Or:  make local      (with local storage)"
echo "   Or:  go run main.go server"
echo ""

echo "5️⃣  Test the API:"
echo "   Upload:   curl -X POST -F \"file=@file.txt\" http://localhost:8080/upload"
echo "   List:     curl http://localhost:8080/files"
echo "   Download: curl http://localhost:8080/download/{id}"
echo "   Delete:   curl -X DELETE http://localhost:8080/files/{id}"
echo ""

echo "📚 For more details, see S3_SETUP.md"
