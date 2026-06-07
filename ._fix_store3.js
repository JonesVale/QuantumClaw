const path = 'web/default/src/routes/stores.$slug.tsx';
let content = require('fs').readFileSync(path, 'utf8');

// Add </div> at the end to balance the one unclosed div
// Insert before the closing ) and } of the component
content = content.replace('</Card>\n          ))}\n        </div>\n      )\n    )\n  )\n}\n',
                        '</Card>\n          ))}\n        </div>\n      </div>\n      )\n    )\n  )\n}\n');

require('fs').writeFileSync(path, content);
console.log('Fixed: added missing </div>');

// Now build to verify
const { execSync } = require('child_process');
try {
    const result = execSync('cmd /c "cd /d H:\\AiData\\openclaw\\workspace\\QuantumClaw\\web\\default && npx rsbuild build 2>&1 | findstr /v warning | findstr /v info | findstr /v rsbuild"', { timeout: 120000 });
    console.log('Output:', result.stdout.toString().substring(0, 500));
} catch(e) {
    console.log('Build result:', e.stderr ? e.stderr.toString().substring(0, 500) : e.stdout?.toString().substring(0,300));
}
