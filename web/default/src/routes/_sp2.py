content = open('models.tsx', 'rb').read().decode('utf-8')

quantum_section = (
    '            {sidebarOpenGroups.providers && quantumProviders.length > 0 && (\r\n'
    '              <div className="mt-4">\r\n'
    '                <div className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em] px-4 mb-2">\u91cf\u5b50\u8d44\u6e90</div>\r\n'
    '                {quantumProviders.map(([p,c])=>(\r\n'
    '                  <button key={p} onClick={()=>setProv(prov===p?\'\':p)}\r\n'
    "                    className={'w-full flex items-center justify-between px-4 py-2.5 rounded-lg text-xl transition-all '+(prov===p?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>\r\n"
    '                    <span>{p}</span>\r\n'
    '                    <span className="text-lg text-muted-foreground/40">{c}</span>\r\n'
    '                  </button>\r\n'
    '                ))}\r\n'
    '              </div>\r\n'
    '            )}\r\n'
    '          </div>\r\n'
    '          <div className="my-3 border-t border-border/10" />\r\n'
    '          {/* Context */}'
)

close_marker = '</div>\r\n          <div className="my-3 border-t border-border/10" />\r\n          {/* Context */}'
idx = content.find(close_marker)

if idx >= 0:
    content = content[:idx] + quantum_section + content[idx + len(close_marker):]
    print('Quantum section: INSERTED')
else:
    print('CLOSE MARKER: NOT FOUND')
    print('Searching...')
    alt_idx = content.find('{/* Context */}')
    if alt_idx >= 0:
        print('Found context at', alt_idx)
        print(repr(content[alt_idx-120:alt_idx]))

open('models.tsx', 'wb').write(content.encode('utf-8'))
print('DONE')
