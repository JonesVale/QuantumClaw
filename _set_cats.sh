#!/bin/sh
echo '{"id":1,"category":"paid"}' | wget -qO- --header=Content-Type:application/json --header='Authorization: Bearer qc-ra-2c3bb30b853e45e3' --post-file=/dev/stdin http://localhost:3666/api/channel/category
echo ""
echo '{"id":14,"category":"free"}' | wget -qO- --header=Content-Type:application/json --header='Authorization: Bearer qc-ra-2c3bb30b853e45e3' --post-file=/dev/stdin http://localhost:3666/api/channel/category
echo ""
echo '{"id":22,"category":"custom"}' | wget -qO- --header=Content-Type:application/json --header='Authorization: Bearer qc-ra-2c3bb30b853e45e3' --post-file=/dev/stdin http://localhost:3666/api/channel/category
echo ""
echo "=== Verify ==="
wget -qO- --header='Authorization: Bearer qc-ra-2c3bb30b853e45e3' 'http://localhost:3666/api/channel/?scope=all' | grep -o '"id":[0-9]*,"name":"[^"]*","category":"[^"]*"' | head -5