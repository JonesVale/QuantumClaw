#!/bin/sh
TOKEN="***"
echo "=== Set Category ==="
wget -qO- --header="Content-Type: application/json" --header="Authorization: Bearer $TOKEN" --post-data='{"id":1,"category":"paid"}' http://localhost:3666/api/channel/category
echo ""
echo "=== Set Free Category ==="
wget -qO- --header="Content-Type: application/json" --header="Authorization: Bearer $TOKEN" --post-data='{"id":14,"category":"free"}' http://localhost:3666/api/channel/category
echo ""
echo "=== Verify ==="
wget -qO- --header="Authorization: Bearer $TOKEN" "http://localhost:3666/api/channel/?scope=all" | grep -o '"id":[0-9]*,"type":[0-9]*,"name":"[^"]*","category":"[^"]*"' | head -5
