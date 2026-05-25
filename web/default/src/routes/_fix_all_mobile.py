import os
files = ['pricing.tsx', 'rankings.tsx', 'apps.tsx']
for f in files:
    c = open(f, 'rb').read().decode('utf-8')
    c = c.replace("style={{ paddingLeft: hovered ? '304px' : '72px' }}", "className='max-w-[100vw] overflow-x-hidden' style={{ paddingLeft: hovered ? '304px' : '72px' }}")
    c = c.replace('min-h-screen bg-background" style={{backgroundImage:', 'min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:')
    open(f, 'wb').write(c.encode('utf-8'))
    print(f + ': done')
