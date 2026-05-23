const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/layout/app-layout.tsx';
let c = fs.readFileSync(path, 'utf-8');

// Remove any remaining hardcoded language dropdown items
// They look like: <DropdownMenuItem onClick={() => switchLanguage('X')}> ... 3 lines
const pattern = /                <DropdownMenuItem onClick={\(\) => switchLanguage\('[^']+'\)}>\n                  .+\n                <\/DropdownMenuItem>\n/g;
c = c.replace(pattern, '');

fs.writeFileSync(path, c, 'utf-8');

// Verify
const c2 = fs.readFileSync(path, 'utf-8');
const englishLeft = c2.includes("switchLanguage('English')");
const chineseLeft = c2.includes("switchLanguage('中文简体')");
if (!englishLeft && !chineseLeft) {
  console.log('SUCCESS: all hardcoded items removed');
} else {
  console.log('ISSUE: English=' + englishLeft + ' Chinese=' + chineseLeft);
}
