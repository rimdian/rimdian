const Nunjucks = require('nunjucks')

// read argument
const args = process.argv.slice(2);
if (args.length < 1) {
    console.error('Usage: node nunjucks.js <content> <data>');
    process.exit(1);
}

// decode base64
const template = Buffer.from(args[0], 'base64').toString('ascii');
const data = JSON.parse(Buffer.from(args[1], 'base64').toString('ascii'));

try {
    console.log(Nunjucks.renderString(template, data))
    return
}
catch (e) {
    console.log(e.message)
    return
}
