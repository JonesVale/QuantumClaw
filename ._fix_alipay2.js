var fs = require('fs');
var c = fs.readFileSync('common/common.go', 'utf8');
c = c.replace('AlipayMinTopUp string', 'AlipayMinTopUp string\n\tAlipaySubject  string');
fs.writeFileSync('common/common.go', c);
console.log('Added AlipaySubject to PaymentSetting');
