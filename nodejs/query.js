// test
// {"db": "dev_rmd_acme_demoecommerce", "query": "SELECT * FROM `user` WHERE external_id = ? LIMIT 1;", "args": ["anon-41-11"]}
// node query.js eyJkYiI6ICJkZXZfcm1kX2FjbWVfZGVtb2Vjb21tZXJjZSIsICJxdWVyeSI6ICJTRUxFQ1QgKiBGUk9NIGB1c2VyYCBXSEVSRSBleHRlcm5hbF9pZCA9ID8gTElNSVQgMTsiLCAiYXJncyI6IFsiYW5vbi00MS0xMSJdfQ==

const mysql = require('mysql');

// read argument
const args = process.argv.slice(2);
if (args.length < 1) {
    console.error('Usage: node query.js <payload>');
    return
}

// decode base64
const payload = JSON.parse(Buffer.from(args[0], 'base64').toString('utf8'));

let conn = null;
try {

    let config = {
        host: '127.0.0.1',
        user: 'root',
        database: payload.db,
        insecureAuth: true
    }

    if (process.env.NODE_ENV === 'production') {
        config = {
            host: process.env.DB_HOST,
            user: process.env.DB_USER,
            password: process.env.DB_PASSWORD,
            database: payload.db,
            ssl: {
                ca: Buffer.from(process.env.DB_CA_CERT_BASE64, 'base64').toString('utf8'),
                rejectUnauthorized: false
            }
        }
    }

    conn = mysql.createConnection(config);

    conn.query({ sql: payload.query, values: payload.args }, (err, rows) => {

        if (err) {
            // Important:  prefix the output with 'error: ' so that the client can detect it
            console.log('error: ' + (err.sqlMessage || err.message))
            process.exit();
        }

        conn.end(() => {
            console.log(JSON.stringify(rows));
            process.exit();
        });
    });

} catch (e) {
    console.log(e.message)

    if (conn !== null) {
        conn.end(() => {
            process.exit(1);
        });
    } else {
        process.exit(1);
    }
}