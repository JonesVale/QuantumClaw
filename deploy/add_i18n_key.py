import json

files = [
    ('web/default/src/i18n/en.json', 'chat_error_connection', 'Please check your connection or try again.'),
    ('web/default/src/i18n/zh-CN.json', 'chat_error_connection', '请检查你的连接或重试'),
]

for fname, key, val in files:
    with open(fname, 'r', encoding='utf-8') as f:
        data = json.load(f)
    data[key] = val
    with open(fname, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    print(f'Updated {fname}: {key}')
