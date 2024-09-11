import { Tooltip, Tabs, Tag, Table, Collapse, Alert } from 'antd'
import {
  AppManifest,
  App,
  AppTable,
  ExtraColumnsManifest,
  DataHook,
  SqlQuery,
  DataHookFor,
  CubeSchema
} from 'interfaces'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCircleCheck } from '@fortawesome/free-solid-svg-icons'
import { ItemKindTagsColors } from 'utils/colors'
import Code from 'utils/prism'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import { useMemo } from 'react'
import TableTag from 'components/common/partial_table_tag'
import { map, size } from 'lodash'
import { faCircleXmark } from '@fortawesome/free-solid-svg-icons'

type BlockAboutAppProps = {
  manifest: AppManifest
  installedApp?: App
}

const BlockAboutApp = (props: BlockAboutAppProps) => {
  let extraColumnsCount = 0

  // console.log('manifest', props.manifest)
  props.manifest.extra_columns?.forEach((table: ExtraColumnsManifest) => {
    extraColumnsCount += table.columns.length
  })

  const tabs = useMemo(() => {
    const extraColumnsTab = {
      key: 'extraColumns',
      label: <span>Extra columns ({extraColumnsCount})</span>,
      children: (
        <>
          {props.manifest.extra_columns?.length === 0 ? (
            <div className={CSS.margin_v_l}>
              <i>No extra columns</i>
            </div>
          ) : (
            <div className={CSS.margin_v_l}>
              {props.manifest.extra_columns?.map((col: any) => (
                <div key={col.kind}>
                  <Table
                    size="small"
                    pagination={false}
                    rowKey={(row: any) => col.kind + row.name}
                    columns={[
                      {
                        title: 'Table & column',
                        key: 'name',
                        render: (row: any) => (
                          <Tooltip title={row.description}>
                            <Tag color={ItemKindTagsColors[col.kind]}>{col.kind}</Tag>
                            {' ' + row.name}
                          </Tooltip>
                        )
                      },
                      {
                        title: 'Type',
                        key: 'type',
                        render: (row: any) => (
                          <span>
                            {row.type}
                            {row.size ? '(' + row.size + ')' : ''}
                            {row.extra_definition ? ' ' + row.extra_definition : ''}
                          </span>
                        )
                      },
                      // {
                      //   title: 'Default',
                      //   key: 'default',
                      //   render: (row: any) => (
                      //     <span>

                      //       {row.is_required && (
                      //         <FontAwesomeIcon icon={faCircleCheck} className={CSS.text_green} />
                      //       )}
                      //     </span>
                      //   )
                      // },
                      {
                        title: 'Required?',
                        key: 'required',
                        render: (row: any) => (
                          <span>
                            {row.is_required && (
                              <FontAwesomeIcon icon={faCircleCheck} className={CSS.text_green} />
                            )}
                          </span>
                        )
                      }
                    ]}
                    dataSource={col.columns}
                  />
                </div>
              ))}
            </div>
          )}
        </>
      )
    }

    const customTablesTab = {
      key: 'customTable',
      label: (
        <span>
          Custom tables ({props.manifest.app_tables ? props.manifest.app_tables.length : 0})
        </span>
      ),
      children: (
        <>
          {!props.manifest.app_tables || props.manifest.app_tables.length === 0 ? (
            <div className={CSS.margin_v_l}>
              <i>No custom tables</i>
            </div>
          ) : (
            <>
              <Collapse
                className={CSS.margin_t_l}
                bordered={false}
                accordion
                defaultActiveKey={[props.manifest.app_tables[0].name]}
              >
                {props.manifest.app_tables.map((table: AppTable) => (
                  <Collapse.Panel header={table.name} key={table.name}>
                    <Table
                      dataSource={table.columns}
                      size="small"
                      pagination={false}
                      showHeader={false}
                      rowKey="name"
                      columns={[
                        {
                          title: 'Column',
                          key: 'name',
                          render: (row: any) => {
                            return (
                              <>
                                <b>{row.name}</b>
                                {row.is_required && (
                                  <>
                                    <span className={CSS.padding_l_s}>
                                      <Tag color="lime">Required</Tag>
                                    </span>
                                  </>
                                )}
                              </>
                            )
                          }
                        },
                        {
                          title: 'Desc',
                          key: 'desc',
                          render: (row: any) => row.description
                        },
                        {
                          title: 'Type',
                          key: 'type',
                          render: (row: any) => (
                            <span>
                              {row.type}
                              {row.size ? '(' + row.size + ')' : ''}
                              {row.extra_definition ? ' ' + row.extra_definition : ''}
                            </span>
                          )
                        }
                      ]}
                    />
                  </Collapse.Panel>
                ))}
              </Collapse>
            </>
          )}
        </>
      )
    }

    const cubeSchemasTab = {
      key: 'cubeSchemas',
      label: (
        <span>
          Cube schemas ({props.manifest.cube_schemas ? size(props.manifest.cube_schemas) : 0})
        </span>
      ),
      children: (
        <>
          {!props.manifest.cube_schemas || !size(props.manifest.cube_schemas) ? (
            <div className={CSS.margin_v_l}>
              <i>No cube schemas</i>
            </div>
          ) : (
            <>
              <Collapse className={CSS.margin_t_l} bordered={false} accordion defaultActiveKey="">
                {map(props.manifest.cube_schemas, (schema: CubeSchema, cubeName: string) => (
                  <Collapse.Panel header={cubeName} key={cubeName}>
                    TODO
                    {/* <Table
                      dataSource={schema.columns}
                      size="small"
                      pagination={false}
                      showHeader={false}
                      rowKey="name"
                      columns={[
                        {
                          title: 'Column',
                          key: 'name',
                          render: (row: any) => {
                            return (
                              <>
                                <b>{row.name}</b>
                                {row.is_required && (
                                  <>
                                    <span className={CSS.padding_l_s}>
                                      <Tag color="lime">Required</Tag>
                                    </span>
                                  </>
                                )}
                              </>
                            )
                          }
                        },
                        {
                          title: 'Desc',
                          key: 'desc',
                          render: (row: any) => row.description
                        },
                        {
                          title: 'Type',
                          key: 'type',
                          render: (row: any) => (
                            <span>
                              {row.type}
                              {row.size ? '(' + row.size + ')' : ''}
                              {row.extra_definition ? ' ' + row.extra_definition : ''}
                            </span>
                          )
                        }
                      ]}
                    /> */}
                  </Collapse.Panel>
                ))}
              </Collapse>
            </>
          )}
        </>
      )
    }

    const tasksTab = {
      key: 'task',
      label: <span>Tasks ({props.manifest.tasks ? props.manifest.tasks.length : 0})</span>,
      children: (
        <Table
          // showHeader={false}
          pagination={false}
          size="middle"
          dataSource={props.manifest.tasks}
          className={CSS.margin_t_l}
          rowKey="external_id"
          columns={[
            {
              title: 'Name',
              key: 'key',
              // className: 'text-right',
              render: (record) => {
                return <Tooltip title={record.external_id}>{record.name}</Tooltip>
              }
            },
            {
              title: 'On multiple exec',
              key: 'value',
              render: (record) => record.on_multiple_exec.replace('_', ' ')
            },
            {
              title: 'Cron',
              key: 'value',
              render: (record) => {
                if (!record.is_cron) return '-'
                return 'Every ' + record.minutes_interval + ' mins'
              }
            }
          ]}
        />
      )
    }

    const dataHooksTab = {
      key: 'dataHook',
      label: (
        <span>Data hooks ({props.manifest.data_hooks ? props.manifest.data_hooks.length : 0})</span>
      ),
      children: (
        <>
          <div className={CSS.margin_v_l}>
            <Table
              // showHeader={false}
              pagination={false}
              size="middle"
              dataSource={props.manifest.data_hooks}
              className={CSS.margin_t_l}
              rowKey="id"
              columns={[
                {
                  title: 'Name',
                  key: 'key',
                  // className: 'text-right',
                  render: (record: DataHook) => {
                    return <Tooltip title={record.id}>{record.name}</Tooltip>
                  }
                },
                {
                  title: 'On',
                  key: 'on',
                  render: (record: DataHook) => record.on
                },
                {
                  title: 'For',
                  key: 'for',
                  render: (record: DataHook) => {
                    return record.for.map((x: DataHookFor) => (
                      <div key={x.kind}>
                        <TableTag table={x.kind} /> - {x.action}
                      </div>
                    ))
                  }
                }
              ]}
            />
          </div>
        </>
      )
    }

    let sqlAccessCount = 0
    if (props.manifest.sql_access) {
      sqlAccessCount += props.manifest.sql_access.predefined_queries?.length || 0
      sqlAccessCount += props.manifest.sql_access.tables_permissions?.length || 0
    }

    const tablePermissions = props.manifest.sql_access?.tables_permissions || []

    // append app_tables to tablePermissions
    if (props.manifest.app_tables) {
      props.manifest.app_tables.forEach((table) => {
        tablePermissions.push({
          table: table.name,
          read: true,
          write: true
        })
      })
    }

    const predefinedQueries = props.manifest.sql_access?.predefined_queries || []

    const sqlAccessTab = {
      key: 'sqlAccess',
      label: <span>SQL access ({sqlAccessCount})</span>,
      children: (
        <>
          <div className={CSS.margin_v_l}>
            {tablePermissions.length > 0 && (
              <Table
                pagination={false}
                size="middle"
                dataSource={tablePermissions}
                className={CSS.margin_t_l}
                rowKey="table"
                columns={[
                  {
                    title: 'Tables access',
                    key: 'table',
                    render: (record) => {
                      return <TableTag table={record.table} />
                    }
                  },
                  // read
                  {
                    title: (
                      <>
                        Read <div className={CSS.font_size_xxs}>SELECT</div>
                      </>
                    ),
                    key: 'read',
                    render: (record) => {
                      if (record.read) {
                        return <FontAwesomeIcon icon={faCircleCheck} className={CSS.text_green} />
                      }
                      return (
                        <FontAwesomeIcon
                          icon={faCircleXmark}
                          className={CSS.text_stone + ' ' + CSS.opacity_30}
                        />
                      )
                    }
                  },
                  // write
                  {
                    title: (
                      <>
                        Write <div className={CSS.font_size_xxs}>INSERT, UPDATE, DELETE</div>
                      </>
                    ),
                    key: 'write',
                    render: (record) => {
                      if (record.write) {
                        return <FontAwesomeIcon icon={faCircleCheck} className={CSS.text_green} />
                      }
                      return (
                        <FontAwesomeIcon
                          icon={faCircleXmark}
                          className={CSS.text_stone + ' ' + CSS.opacity_30}
                        />
                      )
                    }
                  }
                ]}
              />
            )}
            {predefinedQueries.length > 0 && (
              <Table
                // showHeader={false}
                pagination={false}
                size="middle"
                dataSource={predefinedQueries}
                className={CSS.margin_t_l}
                rowKey="id"
                columns={[
                  {
                    title: 'Predefined queries',
                    key: 'id',
                    width: '40%',
                    // className: 'text-right',
                    render: (record: SqlQuery) => {
                      return (
                        <div>
                          <p>
                            <Tag color="green">{record.type}</Tag>
                            <b>
                              <Tooltip title={record.id}>{record.name}</Tooltip>
                            </b>
                          </p>
                          {record.description}
                        </div>
                      )
                    }
                  },
                  {
                    title: 'SQL',
                    key: 'query',
                    render: (record: SqlQuery) => {
                      if (record.query === '*') {
                        return (
                          <Alert
                            type="warning"
                            message="Query has SELECT * access on all tables."
                          />
                        )
                      }
                      return (
                        <div>
                          <div className={CSS.font_size_xs}>
                            <Code language="sql" style={{ overflowWrap: 'anywhere' }}>
                              {record.query}
                            </Code>
                          </div>
                          {record.test_args && (
                            <div>
                              <div className={CSS.padding_v_m}>
                                <b>Test args:</b>
                              </div>
                              <Code language="json">
                                {JSON.stringify(record.test_args, null, 2)}
                              </Code>
                            </div>
                          )}
                        </div>
                      )
                    }
                  }
                ]}
              />
            )}
          </div>
        </>
      )
    }

    const items = []

    if (extraColumnsCount > 0) items.push(extraColumnsTab)
    if (props.manifest.app_tables?.length) items.push(customTablesTab)
    if (size(props.manifest.cube_schemas) > 0) items.push(cubeSchemasTab)
    if (props.manifest.data_hooks?.length) items.push(dataHooksTab)
    if (
      props.manifest.sql_access?.tables_permissions ||
      props.manifest.sql_access?.predefined_queries
    )
      items.push(sqlAccessTab)
    if (props.manifest.tasks?.length) items.push(tasksTab)

    items.push({
      key: 'manifest',
      label: 'App manifest',
      children: (
        <>
          <Code language="json">{JSON.stringify(props.manifest, null, 2)}</Code>
        </>
      )
    })

    if (props.installedApp) {
      items.push({
        key: 'state',
        label: 'State',
        children: (
          <>
            <Code language="json">{JSON.stringify(props.installedApp.state, null, 2)}</Code>
          </>
        )
      })
    }

    return items
  }, [props.installedApp, props.manifest, extraColumnsCount])

  return (
    <>
      <p>
        <b>Author</b>: {props.manifest.author}
        {props.manifest.is_native && (
          <Tooltip title="Native app">
            <FontAwesomeIcon
              icon={faCircleCheck}
              className={css([CSS.text_green, CSS.margin_l_s])}
            />
          </Tooltip>
        )}
      </p>
      <p>
        <b>Homepage</b>: <a href={props.manifest.homepage}>{props.manifest.homepage}</a>
      </p>
      <p>
        <b>Webhooks endpoint</b>: {props.manifest.webhook_endpoint}
      </p>
      <p>
        <b>Version</b>: {props.manifest.version}
      </p>
      <p>{props.manifest.description}</p>

      <Tabs defaultActiveKey="extraColumns" items={tabs} />
    </>
  )
}

export default BlockAboutApp
