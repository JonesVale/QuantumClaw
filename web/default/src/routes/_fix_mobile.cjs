const fs = require('fs');
let c = fs.readFileSync('models.tsx', 'utf8');

// 1) Add break-words to model descriptions
c = c.replace('max-w-prose', 'max-w-prose break-words');

// 2) Fix padding wrapper: use responsive Tailwind on mobile + state-based on desktop
c = c.replace(
  'className="transition-all duration-200" style={{ paddingLeft: hovered ? \'304px\' : \'72px\' }}',
  'className="max-w-[100vw] overflow-x-hidden" style={{ paddingLeft: hovered ? \'304px\' : \'72px\' }}'
);

// 3) Same fix for pricing/rankings/apps padding pattern
// Find all other pages with same structure later

// 4) Add overflow-hidden to body to prevent horizontal scroll
c = c.replace(
  'min-h-screen bg-background" style={{backgroundImage:',
  'min-h-screen bg-background overflow-x-hidden" style={{backgroundImage:'
);

fs.writeFileSync('models.tsx', c, 'utf8');
console.log('Fixed mobile overflow');
