@echo off
title Start MinIO, hmber, and Docker Compose
echo ============================================
echo 🚀 Starting all services...
echo ============================================

REM --- Start MinIO in a new window ---
echo Starting MinIO server...
start "MinIO Server" cmd /c ".\minio.exe server E:\"
timeout /t 3 >nul

REM --- Run Docker Compose ---
echo Starting Docker containers...
docker compose up -d
timeout /t 10 >nul

REM --- Start hmber in a new window ---
echo Starting hmber service...
start "hmber Service" cmd /c ".\hmber.exe"



echo ============================================
echo ✅ All services started successfully!
echo ============================================

pause
