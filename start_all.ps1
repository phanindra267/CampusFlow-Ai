$ErrorActionPreference = "Stop"

Write-Host "========================================="
Write-Host "CampusCare AI - Starting Local Environment"
Write-Host "========================================="

# 1. Frontend
Write-Host "`n[1] Starting Next.js Frontend Development Server..."
Start-Process -NoNewWindow -FilePath "npm.cmd" -ArgumentList "run", "dev" -WorkingDirectory "frontend"
Write-Host "Frontend is launching on http://localhost:3000"

# 2. Backend
Write-Host "`n[2] Recompiling and Starting Go API..."
go build -o .\bin\api.exe .\cmd\api\main.go
Start-Process -NoNewWindow -FilePath ".\bin\api.exe" -WorkingDirectory "."
Write-Host "Backend API is launching on http://localhost:8080/api/v1/health/live"

Write-Host "`nAll processes initiated. Keep this terminal open or check http://localhost:3000!"
