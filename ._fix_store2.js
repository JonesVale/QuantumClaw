let lines = require('fs').readFileSync('web/default/src/routes/stores.$slug.tsx','utf8').split('\n');

// Show the unclosed div context
for (let i = 68; i < 78; i++) {
    console.log('L' + (i+1) + ': ' + lines[i]);
}

// Fix: add missing </div> after line 71 (or wherever it's needed)
// The div opens at line 71 but doesn't close
console.log('\nLine 71 opens:', lines[70].trim());
