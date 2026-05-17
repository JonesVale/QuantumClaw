# Docker Cleanup Script
# Automatically runs after docker-compose up to prevent disk space issues
# Usage: Run this script after docker-compose build

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Docker Cleanup Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

# Show current disk usage before cleanup
Write-Host ""
Write-Host "Before cleanup:" -ForegroundColor Yellow
docker system df

# Clean up unused images, containers, and build cache
Write-Host ""
Write-Host "Cleaning up Docker resources..." -ForegroundColor Green
docker system prune -f

# Clean up build cache specifically (most space consuming)
Write-Host ""
Write-Host "Cleaning build cache..." -ForegroundColor Green
docker builder prune -f

# Clean up any dangling volumes
Write-Host ""
Write-Host "Cleaning dangling volumes..." -ForegroundColor Green
docker volume prune -f

# Show disk usage after cleanup
Write-Host ""
Write-Host "After cleanup:" -ForegroundColor Yellow
docker system df

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Cleanup completed successfully!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan
