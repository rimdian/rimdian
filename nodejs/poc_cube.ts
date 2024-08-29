import { MysqlQuery, prepareCompiler } from '@cubejs-backend/schema-compiler'
import { FileContent, SchemaFileRepository } from '@cubejs-backend/shared'

class myRepo implements SchemaFileRepository {
  constructor(securityContext: any) {
    // securityContext should contain the workspace_id
    // this.securityContext = securityContext
  }

  localPath(): string {
    return 'path/to/schema/files'
  }

  async dataSchemaFiles(includeDependencies?: boolean | undefined): Promise<FileContent[]> {
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
  // return [
  //     { fileName: 'Users', content: Users },
  //     { fileName: 'Sessions', content: Sessions },
  //     { fileName: 'Devices', content: Devices },
  //     { fileName: 'Orders', content: Orders },
  //     { fileName: 'Carts', content: Carts },
  //     ...
  // ];
}

// create a timer to measure the time taken to compile the schema
console.time('compile')

const { compiler, joinGraph, cubeEvaluator } = prepareCompiler(new myRepo(undefined))

compiler
  .compile()
  .then(() => {
    console.timeEnd('compile')

    console.time('query')

    const query = new MysqlQuery(
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
