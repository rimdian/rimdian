// const CubejsServer = require('@cubejs-backend/server');
// const MysqlDriver = require('@cubejs-backend/mysql-driver');
// require('dotenv').config();
// const axios = require('axios');
// const https = require('https');
// // const Nunjucks = require('nunjucks');
// const UAParser = require('ua-parser-js');
// const Joi = require('@hapi/joi');
// const express = require('express');
// // const Paseto = require('paseto.js');
// const _ = require('lodash');

// const refreshSchemaEveryXsecs = process.env.NO_CACHE ? 1 : 60;
// const lastSchemaUpdate = {}
// // dont wait for graceful shutdown in dev
// const gracefulTimeout = process.env.CUBEJS_DB_USER === 'dev' ? 0 : 30 * 1000

// // force port 4444 because the golang PORT will be 8080
// process.env.PORT = 4444;

// // espace single quotes to avoid Cube SQL injection
// const espaceQuotes = (str) => {
//     return str.replace(/'/g, "\\'")
// }

// const parseUA = (req, res) => {
//     // console.log('req.body', req.body);
//     // validate body
//     const schema = Joi.object({
//         user_agent: Joi.string().required()
//     });

//     const { error } = schema.validate(req.body);
//     if (error) {
//         res.status(400).send(error.details[0].message);
//         return;
//     }

//     const { user_agent } = req.body;

//     try {
//         const result = UAParser(user_agent);
//         // { ua: '', browser: {}, cpu: {}, device: {}, engine: {}, os: {} }

//         // wtf this lib has no support for desktop?!
//         if (!result.device.type) {
//             result.device.type = 'desktop';
//         }
//         return res.status(200).send(result);
//     } catch (e) {
//         console.error('parse error:', e.message)
//         return res.status(500).send(e.message);
//     }
// };

// // const compileTemplate = (req, res) => {
// //     // validate body
// //     const schema = Joi.object({
// //         template: Joi.string().required(),
// //         data: Joi.object().required()
// //     });

// //     const { error } = schema.validate(req.body);
// //     if (error) {
// //         return res.status(400).send(error.details[0].message);
// //     }
// //     const { template, data } = req.body;

// //     try {
// //         const result = Nunjucks.renderString(template, data);
// //         return res.status(200).send(result);
// //     } catch (e) {
// //         console.error(e);
// //         return res.status(500).send(e.message);
// //     }
// // }



// // each workspace has its own cache
// // class SchemaCache {
// //     constructor() {
// //         // cache of schema files, key is workspace_id
// //         this.cache = {};
// //         this.cacheTimeoutInSeconds = refreshSchemaEveryXsecs
// //     }

// //     set(workspaceId, schemaFiles) {
// //         this.cache[workspaceId] = {
// //             files: schemaFiles,
// //             lastUpdated: new Date()
// //         };
// //     }

// //     get(workspaceId) {
// //         // return null if the workspace is not in the cache or the cache is expired
// //         if (!this.cache[workspaceId] || this.cache[workspaceId].lastUpdated < new Date(new Date() - 1000 * this.cacheTimeoutInSeconds)) {
// //             return null;
// //         }

// //         return this.cache[workspaceId].files;
// //     }
// // }

// // singleton
// // const schemaCache = new SchemaCache();

// // // fetches schema files from the CM API
// // class CMFileRepository {

// //     constructor(securityContext, schemaCache) {
// //         // securityContext should contain the :
// //         // workspace_id
// //         // API schema_url (the schema url already includes the workspace_id and the token)
// //         this.securityContext = securityContext;
// //         this.schemaCache = schemaCache;
// //         this.isFetching = false
// //     }

// //     // if is already fetching, wait for the fetch to finish to avoid spamming the API
// //     async fetchSchemaFiles() {
// //         while (this.isFetching) {
// //             // test every 100ms
// //             await new Promise(resolve => setTimeout(resolve, 100));
// //         }
// //         return this.requestSchemaFiles();
// //     }

// //     // fetch the schema files from the API with axios
// //     async requestSchemaFiles() {
// //         this.isFetching = true
// //         try {
// //             // console.log('Fetching schema files from the API')
// //             const response = await axios.get(this.securityContext.schema_url, {
// //                 httpsAgent: agent,
// //                 headers: {
// //                     'Content-Type': 'application/json',
// //                     'Accept': 'application/json'
// //                 }
// //             });
// //             // console.log('res', response)
// //             this.isFetching = false
// //             if (response.status !== 200) {
// //                 throw new Error(`Failed to fetch schema files from the API: ${response.status}`);
// //             }
// //             if (!response.data) {
// //                 throw new Error('Failed to fetch schema files from the API: no data');
// //             }
// //             return response.data;
// //         } catch (error) {
// //             console.log(error);
// //             this.isFetching = false
// //             return null;
// //         }
// //     }

// // returns hardcoded schemas
// // in the future, should connect to the DB to request custom tables/columns
// class CMFileRepository {

//     constructor(securityContext) {
//         // securityContext should contain the workspace_id
//         this.securityContext = securityContext;
//     }

//     // return a promise
//     // VERSION 0.31 OF CUBEJS DOESNT WORK WITH ASYNC DATA SCHEMA FILES
//     async dataSchemaFiles() {
//         return new Promise((resolve, reject) => {
//             // console.log('Fetching schema files from the API ' + this.securityContext.schema_url)

//             axios.get(this.securityContext.schema_url, {
//                 // httpsAgent that doesnt verificate the certificate
//                 httpsAgent: new https.Agent({
//                     keepAlive: true,
//                     timeout: 10 * 1000, //10secs
//                     rejectUnauthorized: false
//                 }),

//                 headers: {
//                     'Content-Type': 'application/json',
//                     'Accept': 'application/json'
//                 }
//             }).then((response) => {
//                 // console.log('res', response.data)

//                 if (response.status !== 200) {
//                     return reject(new Error(`Failed to fetch schema files from the API: ${response.status}`));
//                 }
//                 if (!response.data) {
//                     return reject(new Error('Failed to fetch schema files from the API: no data'));
//                 }

//                 // convert JSON files to JS "function" files
//                 const result = []

//                 response.data.forEach((file) => {
//                     // console.log('file', file)
//                     // console.log('file.content', file.content)
//                     // console.log('file.content', JSON.stringify(file.content))
//                     const contentSchema = JSON.parse(file.content)
//                     const measures = []
//                     const dimensions = []
//                     const segments = []
//                     const joins = []

//                     // build measures
//                     _.forEach(contentSchema.measures, (measure, key) => {
//                         const properties = [
//                             `title: '${espaceQuotes(measure.title)}'`,
//                             `type: '${measure.type}'`,
//                             `description: '${espaceQuotes(measure.description)}'`
//                         ]

//                         if (measure.sql && measure.sql !== '') {
//                             properties.push(`sql: \`${espaceQuotes(measure.sql)}\``)
//                         }
//                         if (measure.drillMembers && measure.drillMembers.length > 0) {
//                             properties.push(`drillMembers: [${measure.drillMembers.map((drillMember) => `'${espaceQuotes(drillMember)}'`).join(', ')}]`)
//                         }
//                         if (measure.filters && measure.filters.length > 0) {

//                             properties.push(`filters: [${measure.filters.map((filter) => `{sql: \`${espaceQuotes(filter.sql)}\`}`).join(', ')}]`)
//                         }
//                         if (measure.format && measure.format !== '') {
//                             properties.push(`format: '${espaceQuotes(measure.format)}'`)
//                         }
//                         if (measure.rollingWindow) {
//                             const rollingWindowProperties = []
//                             if (measure.rollingWindow.trailing && measure.rollingWindow.trailing !== '') {
//                                 rollingWindowProperties.push(`trailing: \`${espaceQuotes(measure.rollingWindow.trailing)}\``)
//                             }
//                             if (measure.rollingWindow.leading && measure.rollingWindow.leading !== '') {
//                                 rollingWindowProperties.push(`leading: \`${espaceQuotes(measure.rollingWindow.leading)}\``)
//                             }
//                             if (measure.rollingWindow.offset && measure.rollingWindow.offset !== '') {
//                                 rollingWindowProperties.push(`offset: \`${espaceQuotes(measure.rollingWindow.offset)}\``)
//                             }
//                             properties.push(`rollingWindow: {${rollingWindowProperties.join(', ')}}`)
//                         }
//                         if (measure.shown && !measure.shown) {
//                             properties.push(`shown: false`)
//                         }
//                         if (measure.meta) {
//                             // JSON encode the meta object
//                             properties.push(`meta: '${espaceQuotes(JSON.stringify(measure.meta))}'`)
//                         }

//                         measures.push(`${key}: {${properties.join(', ')}}`)
//                     })

//                     // build dimensions
//                     _.forEach(contentSchema.dimensions, (dimension, key) => {

//                         const properties = [
//                             `title: '${espaceQuotes(dimension.title)}'`,
//                             `type: '${dimension.type}'`,
//                             `description: '${espaceQuotes(dimension.description)}'`
//                         ]

//                         if (dimension.sql && dimension.sql !== '') {
//                             properties.push(`sql: \`${espaceQuotes(dimension.sql)}\``)
//                         }
//                         if (dimension.primaryKey) {
//                             properties.push(`primaryKey: true`)
//                         }
//                         if (dimension.format && dimension.format !== '') {
//                             properties.push(`format: '${espaceQuotes(dimension.format)}'`)
//                         }
//                         if (dimension.shown && !dimension.shown) {
//                             properties.push(`shown: false`)
//                         }
//                         if (dimension.meta) {
//                             // JSON encode the meta object
//                             properties.push(`meta: '${espaceQuotes(JSON.stringify(dimension.meta))}'`)
//                         }
//                         if (dimension.subquery) {
//                             properties.push(`subquery: true`)
//                         }
//                         if (dimension.propagateFiltersToSubQuery) {
//                             properties.push(`propagateFiltersToSubQuery: true`)
//                         }
//                         if (dimension.case) {
//                             const caseProperties = []
//                             if (dimension.case.when && dimension.case.when.length > 0) {
//                                 caseProperties.push(`when: [${dimension.case.when.map((when) => `{sql: \`${espaceQuotes(when.sql)}\`, label: '${espaceQuotes(when.label)}'}`).join(', ')}]`)
//                             }
//                             if (dimension.case.else) {
//                                 caseProperties.push(`else: {label: '${espaceQuotes(dimension.case.else.label)}'}`)
//                             }
//                             properties.push(`case: {${caseProperties.join(', ')}}`)
//                         }

//                         dimensions.push(`${key}: {${properties.join(', ')}}`)
//                     })

//                     // build segments
//                     _.forEach(contentSchema.segments, (segment, key) => {
//                         segments.push(`${key}: {sql: \`${espaceQuotes(segment.sql)}\`}`)
//                     })

//                     // build joins
//                     _.forEach(contentSchema.joins, (join, key) => {
//                         joins.push(`${key}: {sql: \`${espaceQuotes(join.sql)}\`, relationship: '${espaceQuotes(join.relationship)}'}`)
//                     })

//                     const content = `cube(\`${file.fileName}\`, {sql: '${contentSchema.sql}', title: '${espaceQuotes(contentSchema.title)}', description: '${espaceQuotes(contentSchema.description)}', segments: {` + segments.join(`,`) + `}, joins: {` + joins.join(`,`) + `}, measures: {` + measures.join(`,`) + `}, dimensions: {` + dimensions.join(`,`) + `}});`

//                     // if (file.fileName === 'Session') {
//                     //     console.log(content)
//                     // }
//                     result.push({
//                         fileName: file.fileName,
//                         content: content
//                     })
//                 })
//                 resolve(result)
//                 // resolve(response.data);
//             }).catch(reject)
//         })
//     };
//     // return [
//     //     { fileName: 'Users', content: Users },
//     //     { fileName: 'Sessions', content: Sessions },
//     //     { fileName: 'Devices', content: Devices },
//     //     { fileName: 'Orders', content: Orders },
//     //     { fileName: 'Carts', content: Carts },
//     // ];
// }

// // multitenancy docs: https://cube.dev/events/multitenancy-workshop.pdf
// const options = {
//     // embed nodejs functions with the cubejs server
//     initApp: (app) => {
//         // app.use(function (req, res, next) {
//         //     res.header("Access-Control-Allow-Origin", "*");
//         //     res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
//         //     next();
//         // }),
//         app.use(express.json());
//         app.use(express.urlencoded({ extended: true }));
//         // app.post('/compile-template', compileTemplate);
//         app.post('/parse-ua', parseUA);
//     },
//     // required settings to allow the console talking directly to cubejs
//     // http: {
//     //     cors: {
//     //         origin: '*',
//     //         allowedHeaders: '*',
//     //     },
//     // },
//     telemetry: false,
//     cacheAndQueueDriver: 'memory', // disable cubestore, singlestore is fast enough
//     scheduledRefreshTimer: false,

//     // called once per tenant
//     // Used to tell Cube which database type is used to store data for a tenant.
//     dbType: 'mysql',

//     // call on each request
//     // Used to tell Cube which tenant is making the current request.
//     contextToAppId: ({ securityContext }) => {
//         if (!securityContext) return 'CUBEJS_APP_ANONYMOUS'
//         else return `CUBEJS_APP_${securityContext.workspace_id}`
//     },

//     contextToOrchestratorId: ({ securityContext }) => `CUBEJS_APP_${securityContext.workspace_id}`,

//     // called once per tenant
//     // Used to tell Cube which database schema to use to store pre-aggregations for a tenant
//     preAggregationsSchema: ({ securityContext }) => {
//         if (!securityContext) return 'CUBEJS_APP_ANONYMOUS'
//         else return `pre_aggregations_${securityContext.workspace_id}`
//     },

//     // called once per datasource
//     // Used to tell Cube which database driver is used for a data source
//     driverFactory: ({ securityContext }) => {
//         const cfg = {
//             readOnly: true,
//             database: process.env.DB_PREFIX + securityContext.workspace_id,
//         }
//         // we don't use ssl in local dev
//         if (process.env.DB_CA_CERT_BASE64 && process.env.DB_CA_CERT_BASE64 !== "") {
//             cfg.ssl = { ca: Buffer.from(process.env.DB_CA_CERT_BASE64, 'base64').toString('utf8') }
//         }
//         return new MysqlDriver(cfg)
//     },

//     schemaVersion: ({ securityContext }) => {
//         if (!securityContext || !securityContext.workspace_id) return 0

//         if (!lastSchemaUpdate[securityContext.workspace_id]) {
//             lastSchemaUpdate[securityContext.workspace_id] = new Date().getTime()
//             return lastSchemaUpdate[securityContext.workspace_id]
//         }

//         // check if the lastSchemaUpdate is older than the cache timeout
//         if (lastSchemaUpdate[securityContext.workspace_id] < new Date(new Date() - 1000 * refreshSchemaEveryXsecs)) {
//             // update the lastSchemaUpdate
//             lastSchemaUpdate[securityContext.workspace_id] = new Date().getTime()
//         }

//         return lastSchemaUpdate[securityContext.workspace_id]
//     },

//     // called once per tenant
//     // Used to tell Cube which data schema files to use for a tenant.
//     repositoryFactory: ({ securityContext }) => new CMFileRepository(securityContext),

//     scheduledRefreshContexts: async () => [
//         // {
//         //     securityContext: {},
//         // },
//     ],


//     // Called by a refresh worker
//     // Allows to build pre-aggregations for all tenants.
//     // scheduledRefreshContexts: () => [
//     //     { securityContext: { tenant: ‘avocado’ } },
//     //     { securityContext: { tenant: ‘mango’ } },
//     // ],

//     // Called on each request
//     // Used to validate or update a request before it’s executed. Often, used to control data access for each tenant
//     // queryRewrite: (query, { securityContext }) => {
//     //     if (securityContext.role !== `admin`) throw new Error(‘...);
//     //     query.filters.push(...);
//     // },


//     // dont use paseto, the console talks to the api that generates a proper cubejs JWT
//     // checkAuth: (req, auth) => {
//     //     console.log('auth', auth);
//     //     console.log('process.env.SECRET_KEY', process.env.SECRET_KEY);
//     //     const encoder = new Paseto.V2();

//     //     encoder.decrypt(auth.token, process.env.SECRET_KEY).then((data) => {
//     //         console.log('data is ', data);
//     //         req.securityContext = data;
//     //         return
//     //     }).catch((err) => {
//     //         console.log(err);
//     //         throw new Error('Could not decrypt token');
//     //     })
//     // }
// }

// const server = new CubejsServer(options);

// server.listen().then(({ port }) => {
//     console.log(`🚀 Cube.js server is listening on ${port}`);
// }).catch(e => {
//     console.error('Fatal error during server start: ');
//     console.error(e.stack || e);
// });

// const closeServer = (code) => {
//     return new Promise(() => {
//         setTimeout(() => {
//             server.close(() => {
//                 console.log('Forcefully shutting down.')
//                 process.exit(code)
//             })
//         }, gracefulTimeout)
//     })
// }

// async function shutdown(signal) {
//     console.log(`Received ${signal}. Shutting down.`);
//     // force to wait 30 secs to make sure the go server is not sending any more requests
//     if (process.env.NODE_ENV === 'production') {
//         console.log('Waiting 30 secs before shutting down.')
//         await closeServer(1)
//     } else {
//         await closeServer(0)
//     }
// }

// process.on('SIGTERM', shutdown)
// process.on("SIGINT", shutdown); // Atom, VSCode, WebStorm or Terminal Ctrl+C