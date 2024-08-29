const SchemaCompiler = require('@cubejs-backend/schema-compiler/dist/src/index')


class myRepo {
    constructor(securityContext) {
        // securityContext should contain the workspace_id
        // this.securityContext = securityContext
    }

    localPath() {
        return 'path/to/schema/files'
    }

    async dataSchemaFiles(includeDependencies) {
        // console.log('Fetching schema files from the API ' + this.securityContext.schema_url)
        return new Promise((resolve, reject) => {
            // console.log('Fetching schema files from the API ' + this.securityContext.schema_url)
            resolve([
                {
                    fileName: 'Visitors.js',
                    content: `cube('visitors', {
            sql: 'select * from visitors',

            measures: {
              count: {
                type: 'count'
              },

              unboundedCount: {
                type: 'count',
                rollingWindow: {
                  trailing: 'unbounded'
                }
              }
            },

            dimensions: {
              createdAt: {
                type: 'time',
                sql: 'created_at'
              },
              name: {
                type: 'string',
                sql: 'name'
              }
            }
        });
          `
                }
            ])
        })
    }
}

// create a timer to measure the time taken to compile the schema
console.time('compile')

const { compiler, joinGraph, cubeEvaluator } = SchemaCompiler.prepareCompiler(new myRepo(undefined), {
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

compiler
    .compile()
    .then(() => {
        console.timeEnd('compile')

        console.time('query')

        const query = new SchemaCompiler.MysqlQuery(
            { joinGraph, cubeEvaluator, compiler },
            {
                measures: ['visitors.count'],
                timeDimensions: [],
                filters: [
                    {
                        member: 'visitors.name',
                        operator: 'equals',
                        values: [null]
                    }
                ],
                timezone: 'America/Los_Angeles'
            }
        )

        const queryAndParams = query.buildSqlAndParams()
        console.log(queryAndParams)

        console.timeEnd('query')
    })
    .catch((e) => {
        console.error(e)
    })
