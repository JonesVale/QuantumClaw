TOKEN=***
echo '{"id":1,"category":"paid"}' | wget -qO- --header=Content-Type:application/json --header="Authorization: Bearer $TOKEN" --post-file=/dev/stdin http://localhost:3666/api/channel/category