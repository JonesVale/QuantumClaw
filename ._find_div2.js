let lines = require('fs').readFileSync('web/default/src/routes/stores.$slug.tsx','utf8').split('\n');
let stack = [];
for (let i = 0; i < lines.length; i++) {
    let opens = (lines[i].match(/<div(\s[^>]*)?>/g) || []).length;
    let closes = (lines[i].match(/<\/div>/g) || []).length;
    for (let o = 0; o < opens; o++) stack.push(i+1);
    for (let c = 0; c < closes; c++) if (stack.length > 0) stack.pop();
}
if (stack.length > 0) console.log('Unclosed divs:', stack.join(','));
else console.log('All divs balanced');
