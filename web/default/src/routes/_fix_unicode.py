import os, sys
c = open('models.tsx', 'rb').read().decode('utf-8')
# Replace literal Unicode escape sequences with actual Chinese chars
c = c.replace('\\u91cf\\u5b50\\u8d44\\u6e90', '\u91cf\u5b50\u8d44\u6e90')
open('models.tsx', 'wb').write(c.encode('utf-8'))
print('Fixed: \\u91cf -> 量子资源')
# Verify
c2 = open('models.tsx', 'rb').read().decode('utf-8')
if '\u91cf\u5b50\u8d44\u6e90' in c2:
    print('Verified: Chinese text present')
if '\\u91cf' in c2:
    print('WARNING: escape still exists')
