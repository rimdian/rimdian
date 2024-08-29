const SchemaCompiler = require('@cubejs-backend/schema-compiler/dist/src/index')

// read argument
const args = process.argv.slice(2);
if (args.length < 1) {
    console.error('Usage: node cube-to-query.js <payload>');
    return
}

const payload = JSON.parse(Buffer.from(args[0], 'base64').toString('utf8'));

// payload: base64 encoded JSON string
// {
//     "cubeQuery": {...},
//     "schemas": [...]
// }
// ex: node cube-to-query.js ewogICAgImN1YmVRdWVyeSI6IHsKICAgICAgICAgICAgICAgICJtZWFzdXJlcyI6IFsidmlzaXRvcnMuY291bnQiXSwKICAgICAgICAgICAgICAgICJ0aW1lRGltZW5zaW9ucyI6IFtdLAogICAgICAgICAgICAgICAgImZpbHRlcnMiOiBbCiAgICAgICAgICAgICAgICAgICAgewogICAgICAgICAgICAgICAgICAgICAgICAibWVtYmVyIjogInZpc2l0b3JzLm5hbWUiLAogICAgICAgICAgICAgICAgICAgICAgICAib3BlcmF0b3IiOiAiZXF1YWxzIiwKICAgICAgICAgICAgICAgICAgICAgICAgInZhbHVlcyI6IFtudWxsXQogICAgICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgICAgIF0sCiAgICAgICAgICAgICAgICAidGltZXpvbmUiOiAiQW1lcmljYS9Mb3NfQW5nZWxlcyIKICAgICAgICAgICAgfSwgCiAgICAic2NoZW1hcyI6IFsKICAgICAgICB7CiAgICAgICAgICAgICJmaWxlTmFtZSI6ICJWaXNpdG9ycy5qcyIsCiAgICAgICAgICAgICJjb250ZW50IjogImN1YmUoJ3Zpc2l0b3JzJywge3NxbDogJ3NlbGVjdCAqIGZyb20gdmlzaXRvcnMnLG1lYXN1cmVzOiB7ICBjb3VudDoge3R5cGU6ICdjb3VudCcgIH0sIHVuYm91bmRlZENvdW50OiB7dHlwZTogJ2NvdW50Jyxyb2xsaW5nV2luZG93OiB7ICB0cmFpbGluZzogJ3VuYm91bmRlZCd9ICB9fSwgZGltZW5zaW9uczogeyAgY3JlYXRlZEF0OiB7dHlwZTogJ3RpbWUnLHNxbDogJ2NyZWF0ZWRfYXQnICB9LCAgbmFtZToge3R5cGU6ICdzdHJpbmcnLHNxbDogJ25hbWUnICB9fX0pOyIKICAgICAgICB9CiAgICBdCn0=

class fakeRepo {
    constructor() {
    }

    localPath() {
        return 'none'
    }

    async dataSchemaFiles() {
        return new Promise((resolve) => {
            resolve(payload.schemas)
        })
    }
}

const { compiler, joinGraph, cubeEvaluator } = SchemaCompiler.prepareCompiler(new fakeRepo(), {
    // nativeInstance?: NativeInstance;
    allowNodeRequire: false,
    allowJsDuplicatePropsInSchema: false,
    // maxQueryCacheSize?: number;
    // maxQueryCacheAge?: number;
    compileContext: {},
    // standalone?: boolean;
    // headCommitId?: string;
    // adapter?: string;
})

compiler.compile().then(() => {

    const query = new SchemaCompiler.MysqlQuery({ joinGraph, cubeEvaluator, compiler }, payload.cubeQuery)
    const queryAndParams = query.buildSqlAndParams()

    // output the query and parameters
    console.log(JSON.stringify({
        sql: queryAndParams[0],
        args: queryAndParams[1] || []
    }))
}).catch((e) => {
    console.error(e)
})
