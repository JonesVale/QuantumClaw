const fs = require('fs');
let c = fs.readFileSync('enterprise.tsx', 'utf8');

// Fix indentation: max-w-4xl needs to be deeper than relative z-10
c = c.replace(
  '<div className="relative z-10">\n          <div className="max-w-4xl',
  '<div className="relative z-10">\n              <div className="max-w-4xl'
);

// Fix all closing div indentation  
c = c.replace(
  '          </div>\n        </div>\n        </div>\n      </div>\n      </section>',
  '              </div>\n            </div>\n          </div>\n        </div>\n      </div>\n      </section>'
);

fs.writeFileSync('enterprise.tsx', c, 'utf8');
console.log('Fixed');
