const fs = require('fs');
const path = 'web/default/src/routes/stores.$slug.tsx';

const lines = fs.readFileSync(path, 'utf8').split('\n');
console.log('=== lines 200-215 ===');
for (let i = 199; i < Math.min(216, lines.length); i++) {
    console.log('L' + (i+1) + ': ' + lines[i]);
}

// Check Card balance
let stack = 0;
for (let i = 0; i < lines.length; i++) {
    const opens = (lines[i].match(/<Card\b(?!e|H|C|T|D)/g) || []).length;
    const closes = (lines[i].match(/<\/Card>/g) || []).length;
    stack = stack + opens - closes;
}
console.log('\nCard stack:', stack, '(should be 0)');

// Check for unclosed tags
const tags = ['Card','CardContent','CardHeader','CardTitle','CardDescription','div','Button'];
tags.forEach(tag => {
    const opens = lines.filter(l => new RegExp('<' + tag + '\\b(?!e|H|C|T|D)','g').test(l)).length - lines.filter(l => new RegExp('<' + tag + '\\b(?!e|H|C|T|D)[^>]*/>','g').test(l)).length;
    const closes = lines.filter(l => new RegExp('</' + tag + '>','g').test(l)).length;
    if (opens !== closes) {
        console.log(tag + ': ' + opens + ' open, ' + closes + ' close (diff: ' + (opens-closes) + ')');
    }
});
