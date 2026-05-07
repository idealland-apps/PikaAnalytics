# PowerShell build script for Windows

Write-Host "Building PikaAnalytics for production..." -ForegroundColor Cyan

# Check if we're in the right directory
if (-not (Test-Path "README.md")) {
    Write-Host "ERROR: Please run this script from the project root directory" -ForegroundColor Red
    exit 1
}

# Create dist directory
Write-Host "Creating distribution directory..." -ForegroundColor Blue
if (Test-Path "dist") {
    Remove-Item "dist" -Recurse -Force
}
New-Item -ItemType Directory -Path "dist" | Out-Null

# Build frontend
Write-Host "Building frontend..." -ForegroundColor Blue
Set-Location "frontend"
npm install
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Frontend build failed" -ForegroundColor Red
    exit 1
}

# Copy frontend build to backend
Write-Host "Preparing backend with frontend assets..." -ForegroundColor Blue
Set-Location "../backend"
if (Test-Path "frontend") {
    Remove-Item "frontend" -Recurse -Force
}
New-Item -ItemType Directory -Path "frontend" | Out-Null
Copy-Item "../frontend/build/*" "frontend/" -Recurse

# Build Go binary
Write-Host "Building Go binary..." -ForegroundColor Blue
go mod tidy
go build -o "../dist/pikaanalytics.exe" main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go build failed" -ForegroundColor Red
    exit 1
}

# Copy necessary files to dist
Set-Location ".."
Copy-Item "backend/frontend" "dist/" -Recurse
if (Test-Path "backend/pikaanalytics.db") {
    Copy-Item "backend/pikaanalytics.db" "dist/"
} else {
    Write-Host "No existing database found, will be created on first run" -ForegroundColor Yellow
}

Write-Host "SUCCESS: Build completed successfully!" -ForegroundColor Green
Write-Host "Production files are in the 'dist' directory" -ForegroundColor Green
Write-Host "To run: cd dist; ./pikaanalytics.exe" -ForegroundColor Green