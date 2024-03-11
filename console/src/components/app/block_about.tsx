import { Tooltip, Tabs, Tag, Table, Collapse, Alert } from 'antd'
import {
  AppManifest,
  App,
  AppTable,
  ExtraColumnsManifest,
  TableJoin,
  DataHook,
  SqlQuery,
  DataHookFor
} from 'interfaces'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCircleCheck } from '@fortawesome/free-solid-svg-icons'
import { ItemKindTagsColors } from 'utils/colors'
import Code from 'utils/prism'
import CSS from 'utils/css'
import { css } from '@emotion/css'
import { useMemo } from 'react'
import TableTag from 'components/common/partial_table_tag'

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
                            const joinedTo = table.joins
                              ? table.joins.find((j: TableJoin) => j.local_column === row.name)
                              : null

                            return (
                              <>
                                <b>{row.name}</b>
                                {row.is_required && (
                                  <>
                                    <span className={CSS.padding_l_s}>
                                      <Tag color="lime">Required</Tag>
                                    </span>
                                    {joinedTo && (
                                      <i className={CSS.font_size_xs}>
                                        {joinedTo.relationship === 'one_to_one' && 'one to one'}
                                        {joinedTo.relationship === 'one_to_many' && 'one to many'}
                                        {joinedTo.relationship === 'many_to_one' && 'many to one'}
                                        <Tag color="blue" className={CSS.margin_l_s}>
                                          {joinedTo.external_table}.{joinedTo.external_column}
                                        </Tag>
                                      </i>
                                    )}
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
    const sqlQueriesTab = {
      key: 'sqlQuery',
      label: (
        <span>
          SQL queries ({props.manifest.sql_queries ? props.manifest.sql_queries.length : 0})
        </span>
      ),
      children: (
        <>
          <div className={CSS.margin_v_l}>
            <Table
              // showHeader={false}
              pagination={false}
              size="middle"
              dataSource={props.manifest.sql_queries}
              className={CSS.margin_t_l}
              rowKey="id"
              columns={[
                {
                  title: 'Name',
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
                  title: 'Query',
                  key: 'query',
                  render: (record: SqlQuery) => {
                    if (record.query === '*') {
                      return (
                        <Alert type="warning" message="Query has SELECT * access on all tables." />
                      )
                    }
                    return (
                      <div>
                        <div className={CSS.font_size_xs} style={{ wordBreak: 'break-all' }}>
                          <Code language="sql">{record.query}</Code>
                        </div>
                        {record.test_args && (
                          <div>
                            <div className={CSS.padding_v_m}>
                              <b>Test args:</b>
                            </div>
                            <Code language="json">{JSON.stringify(record.test_args, null, 2)}</Code>
                          </div>
                        )}
                      </div>
                    )
                  }
                }
              ]}
            />
          </div>
        </>
      )
    }
    const items = []

    if (extraColumnsCount > 0) items.push(extraColumnsTab)
    if (props.manifest.app_tables?.length) items.push(customTablesTab)
    if (props.manifest.data_hooks?.length) items.push(dataHooksTab)
    if (props.manifest.sql_queries?.length) items.push(sqlQueriesTab)
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
