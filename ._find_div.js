const lines = require('fs').readFileSync('web/default/src/routes/stores.$slug.tsx','utf8').split('\n');

// Find unclosed div by tracking open/close
let stack = [];
for (let i = 0; i < lines.length; i++) {
    const divOpens = (lines[i].match(/<div(\s[^>]*)?>/g) || []).length;
    const divCloses = (lines[i].match(/<\/div>/g) || []).length;
    // Exclude self-closing divs
    const selfClosing = (lines[i].match(/<div[^>]*\/>/g) || []).length;
    divOpens -= selfClosing;
    
    for (let o = 0; o < divOpens; o++) stack.push(i+1);
    for (let c = 0; c < divCloses; c++) {
        if (stack.length > 0) stack.pop();
        else console.log('Extra </div> at line', i+1);
    }
}

if (stack.length > 0) {
    console.log('Unclosed <div> at lines:', stack.join(', '));
    stack.forEach(line => {
        console.log('L' + line + ': ' + lines[line-1].trim().substring(0,120));
    });
} else {
    console.log('All <div> tags balanced');
}
