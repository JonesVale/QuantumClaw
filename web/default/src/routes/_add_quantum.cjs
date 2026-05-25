const fs = require('fs');

function addQuantumSeparation(filepath) {
  let content = fs.readFileSync(filepath, 'utf8');
  
  // Find the providers line and add quantum/ai split
  const providersRegex = /const (\w+) = useMemo\(\(\)=>\[\.\.\.new Set\(.+\.provider\)/;
  const match = content.match(providersRegex);
  
  if (match) {
    const varName = match[1];
    const insertPoint = content.indexOf('])\n  const quantumProviders', content.indexOf('sort(),[all])'));
    
    if (content.includes('quantumProviders')) {
      console.log(filepath + ': already has quantum split');
      return;
    }
    
    // Add after the providers memo closing
    const afterProviders = content.indexOf('sort(),[all])\n', content.indexOf(varName));
    if (afterProviders >= 0) {
      const endOfLine = content.indexOf('\n', afterProviders);
      const insertion = '\n  const aiProviders = useMemo(()=>' + varName + '.filter(p=>![\\'IonQ\\',\\'IBM\\',\\'Rigetti\\'].includes(p)),[' + varName + '])\n  const quantumProviders = useMemo(()=>' + varName + '.filter(p=>[\\'IonQ\\',\\'IBM\\',\\'Rigetti\\'].includes(p)),[' + varName + '])';
      
      content = content.slice(0, endOfLine) + insertion + content.slice(endOfLine);
      
      // Replace references from varName to aiProviders in the sidebar
      // Find the sidebar section where providers are rendered
      const sidebarRegex = new RegExp('\\{' + varName + '\\.map\\(', 'g');
      content = content.replace(sidebarRegex, '{aiProviders.map(');
      
      // Add quantum section after AI providers section
      const quantumSection = `\n            {quantumProviders.length > 0 && (
              <div className="mt-4">
                <div className="text-lg font-bold text-muted-foreground/60 uppercase tracking-[0.15em] px-4 mb-2">量子资源</div>
                {quantumProviders.map(p=>(
                  <button key={p} onClick={()=>setProv(prov===p?'':p)}
                    className={'w-full text-left px-4 py-2.5 rounded-lg text-xl transition-all '+(prov===p?'bg-amber-50 text-amber-800 font-medium':'text-muted-foreground hover:text-foreground hover:bg-muted/30')}>
                    {p}
                  </button>
                ))}
              </div>
            )}`;
      
      // Insert before the closing of the sidebar content
      // Find close marker
      const providersEnd = content.lastIndexOf('{/');
      if (providersEnd >= 0) {
        // Find the end of providers section
        const contextMarker = 'Context';
        const contextIdx = content.indexOf(contextMarker, providersEnd);
        
        if (filepath.includes('pricing')) {
          // No context in pricing, insert before the sidebar closing
          const sidebarEnd = content.indexOf('</div>\n          </div>\n        </div>\n      </div>\n    </div>', content.indexOf('style={{ width: hovered'));
          if (sidebarEnd >= 0) {
            content = content.slice(0, sidebarEnd) + quantumSection + content.slice(sidebarEnd);
          }
        }
      }
      
      fs.writeFileSync(filepath, content, 'utf8');
      console.log(filepath + ': quantum split added');
    }
  } else {
    console.log(filepath + ': no providers memo found');
  }
}

addQuantumSeparation('pricing.tsx');
addQuantumSeparation('rankings.tsx');
