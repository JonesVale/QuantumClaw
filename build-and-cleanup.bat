@echo off
REM ============================================================
REM QuantumClaw Smart Build Script
REM Automatically builds and cleans Docker cache to save space
REM ============================================================

echo.
echo ===========================================================
echo  QuantumClaw Smart Build & Cleanup Script
echo ===========================================================
echo.

REM Check if Docker is running
docker info >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: Docker is not running. Please start Docker Desktop first.
    echo.
    pause
    exit /b 1
)

REM Show disk usage before build
echo [INFO] Current Docker disk usage:
docker system df
echo.

REM Clean cache before build to save space
echo [INFO] Cleaning cache before build...
docker builder prune -f
echo.

REM Build the project
echo [INFO] Building Docker images...
docker-compose build
echo.

REM Check if build succeeded
if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

REM Start the containers
echo [INFO] Starting containers...
docker-compose up -d
echo.

REM Clean up build cache after successful build
echo [INFO] Cleaning build cache after build...
docker builder prune -f
echo.

REM Clean up unused images
echo [INFO] Cleaning unused images...
docker image prune -f
echo.

REM Show final disk usage
echo [SUCCESS] Build completed! Final Docker disk usage:
docker system df
echo.

echo ===========================================================
echo  Build and cleanup completed successfully!
echo ===========================================================
echo.
echo You can now access the application at:
echo   http://localhost:3000
echo.
pause
