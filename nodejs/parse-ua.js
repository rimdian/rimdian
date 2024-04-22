const UAParser = require('ua-parser-js');

// read argument
const args = process.argv.slice(2);
if (args.length < 1) {
    console.error('Usage: node parse-ua.js <user_agent>');
    return
}
// decode base64
const decoded = Buffer.from(args[0], 'base64').toString('utf8');


try {
    const result = UAParser(decoded);
    // { ua: '', browser: {}, cpu: {}, device: {}, engine: {}, os: {} }

    // wtf this lib has no support for desktop?!
    if (!result.device.type) {
        result.device.type = 'desktop';
    }

    console.log(JSON.stringify(result))
    return
} catch (e) {
    console.error('parse error:', e.message)
    return
}