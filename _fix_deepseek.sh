#!/bin/sh
# Login to get session
LOGIN=$(wget -qO- --header=Content-Type:application/json --post-data='{"username":"root","password":"***"}' http://localhost:3666/api/user/login 2>/dev/null)
# Extract session cookie from wget response (stored in wget's cookie handling)
wget -qO- --save-cookies=/tmp/cookies.txt --keep-session-cookies --header=Content-Type:application/json --post-data='{"username":"root","password":"***"}' http://localhost:3666/api/user/login > /dev/null 2>&1

# Update DeepSeek channel with correct models
echo "=== Update DeepSeek models ==="
wget -qO- --load-cookies=/tmp/cookies.txt --header=Content-Type:application/json --post-data='{"id":17,"name":"DeepSeek","type":35,"models":"deepseek-chat,deepseek-reasoner,deepseek-v3,deepseek-r1","group":"default","key":"***"}' http://localhost:3666/api/channel

echo ""
echo "=== Verify ==="
wget -qO- --load-cookies=/tmp/cookies.txt "http://localhost:3666/api/channel/?scope=all" 2>/dev/null | grep -o '"id":17,"name":"DeepSeek","models":"[^"]*"'
rm -f /tmp/cookies.txt
