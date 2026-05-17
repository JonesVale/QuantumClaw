@echo off
REM ============================================================
REM Docker System Disk Cleanup Script
REM Cleans Docker build cache from system disk (C:\)
REM ============================================================

echo.
echo ===========================================================
echo  Docker System Disk Cleanup
echo ===========================================================
echo.

echo [INFO] Stopping all running containers...
docker-compose down
echo.

echo [INFO] Cleaning ALL Docker build cache...
docker builder prune -af
echo.

echo [INFO] Cleaning dangling images...
docker image prune -a -f
echo.

echo [INFO] Cleaning unused volumes...
docker volume prune -f
echo.

echo [INFO] Cleaning build cache (system-wide)...
docker builder prune -af --filter "until=0s"
echo.

echo [SUCCESS] All Docker caches cleaned!
echo.
echo [INFO] Final disk usage:
docker system df
echo.

echo ===========================================================
echo  Cleanup completed!
echo ===========================================================
echo.
pause
