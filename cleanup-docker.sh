#!/bin/bash
# Docker Cleanup Script
# Automatically runs after docker-compose up to prevent disk space issues

echo "========================================="
echo "Docker Cleanup Script"
echo "========================================="

# Show current disk usage before cleanup
echo ""
echo "Before cleanup:"
docker system df

# Clean up unused images, containers, and build cache
echo ""
echo "Cleaning up Docker resources..."
docker system prune -f

# Clean up build cache specifically (most space consuming)
echo ""
echo "Cleaning build cache..."
docker builder prune -f

# Clean up any dangling volumes
echo ""
echo "Cleaning dangling volumes..."
docker volume prune -f

# Show disk usage after cleanup
echo ""
echo "After cleanup:"
docker system df

echo ""
echo "========================================="
echo "Cleanup completed successfully!"
echo "========================================="
