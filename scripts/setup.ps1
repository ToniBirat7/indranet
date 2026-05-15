# IndraNet Windows Dev Environment Setup
#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "=== IndraNet Dev Environment Setup ===" -ForegroundColor Cyan

# Check dependencies
function Check-Command($name) {
    if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
        Write-Error "ERROR: $name is not installed or not in PATH"
        exit 1
    }
}

Check-Command "go"
Check-Command "node"
Check-Command "pnpm"
Check-Command "docker"

Write-Host "✓ Dependencies found" -ForegroundColor Green

# Copy env file
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "✓ Created .env from .env.example — edit with your secrets" -ForegroundColor Green
}

# Install pnpm packages
pnpm install
Write-Host "✓ pnpm packages installed" -ForegroundColor Green

# Install Go dependencies
Set-Location packages/backend
go mod download
Set-Location ../..
Write-Host "✓ Go modules downloaded" -ForegroundColor Green

# Start local services
docker compose up -d
Write-Host "✓ docker-compose services started" -ForegroundColor Green

# Wait for postgres
Write-Host "Waiting for postgres to be ready..."
Start-Sleep -Seconds 5
Set-Location packages/backend
& make migrate
Set-Location ../..
Write-Host "✓ Database migrations applied" -ForegroundColor Green

Write-Host ""
Write-Host "=== Setup complete! ===" -ForegroundColor Cyan
Write-Host "  Backend:  cd packages/backend; make dev"
Write-Host "  Web:      cd packages/web; pnpm dev"
Write-Host "  Client:   cd packages/client; pnpm dev"
