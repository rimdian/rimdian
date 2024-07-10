import { Table, Row, Col, Spin, Tag, Tooltip } from 'antd'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faAngleRight, faBars } from '@fortawesome/free-solid-svg-icons'
import { ColumnInformationSchema, TableInformationSchema, TablesDescriptions } from './schemas'
import { useSearchParams } from 'react-router-dom'
import { useMemo } from 'react'
import AlphaSort from 'utils/alpha_sort'
import numbro from 'numbro'
import { get } from 'lodash'
import CSS, { borderColorSecondary, colorPrimary } from 'utils/css'
import Block, { blockCss } from 'components/common/block'
import { css } from '@emotion/css'
import { kpiCss } from 'components/common/partial_kpi'

const menuCSS = {
  item: css({
    padding: CSS.S + ' ' + CSS.S,
    cursor: 'pointer',
    wordBreak: 'break-all',
    '&:not(:last-child)': {
      borderBottom: 'solid 1px ' + borderColorSecondary
    },
    '&:hover': {
      borderBottom: 'solid 1px rgba(78, 108, 255, 0.4)'
    }
  }),
  itemSelected: css({
    color: colorPrimary,
    borderBottom: 'solid 1px rgba(78, 108, 255, 0.4)'
  })
}

type BlockDBSchemasProps = {
  data?: TableInformationSchema[]
  isLoading: boolean
  isFetching: boolean
}

const BlockDBSchemas = (props: BlockDBSchemasProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [params, setSearchParams] = useSearchParams()

  const selectedTableName = params.get('table') || (props.data ? 'user' : '')

  const selectedTable = useMemo(() => {
    return props.data?.find((t) => t.name === selectedTableName)
  }, [props.data, selectedTableName])

  const columns = useMemo(() => {
    if (!props.data || !selectedTable) return []
    return selectedTable.columns
  }, [props.data, selectedTable])

  return (
    <div className={CSS.margin_t_m}>
      <Row gutter={24}>
        <Col span={6}>
          <Block
            title="Tables"
            small
            extra={
              <>
                {(props.isLoading || props.isFetching) && (
                  <Spin style={{ display: 'inline-block' }} size="small" />
                )}
              </>
            }
          >
            {props.data &&
              props.data.map((table: TableInformationSchema) => (
                <div
                  key={table.name}
                  className={css([
                    menuCSS.item,
                    selectedTableName === table.name && menuCSS.itemSelected
                  ])}
                  onClick={() =>
                    setSearchParams({
                      tab: 'tables',
                      table: table.name
                    })
                  }
                >
                  <span className={CSS.pull_right}>
                    <FontAwesomeIcon icon={faAngleRight} />
                  </span>
                  <Tooltip
                    title={
                      table.storage_type === 'COLUMNSTORE' ? 'Columnstore' : 'In-memory rowstore'
                    }
                    placement="topLeft"
                  >
                    <FontAwesomeIcon
                      style={{ opacity: 0.5 }}
                      rotation={table.storage_type === 'COLUMNSTORE' ? 90 : undefined}
                      icon={faBars}
                    />
                    &nbsp;&nbsp;
                  </Tooltip>
                  {table.name} &nbsp;&nbsp;
                  <Tag>
                    {numbro(table.rows || 0).format({
                      average: true,
                      totalLength: 3,
                      trimMantissa: true
                    })}
                  </Tag>
                </div>
              ))}
          </Block>
        </Col>

        <Col span={18}>
          <Block grid small>
            <div className={kpiCss.self}>
              <div className={kpiCss.title}>Table | View</div>
              <div className={kpiCss.value}>
                {props.isLoading && <Spin size="small" />}
                {!props.isLoading &&
                  (selectedTable?.type === 'VIEW'
                    ? 'View ' + selectedTable.name
                    : 'Table ' + selectedTable?.name)}
              </div>
            </div>
            <div className={kpiCss.self}>
              <div className={kpiCss.title}>Storage type</div>
              <div className={kpiCss.value}>
                {props.isLoading && <Spin size="small" />}
                {!props.isLoading &&
                  (selectedTable?.storage_type === 'COLUMNSTORE'
                    ? 'Columnstore'
                    : 'In-memory Rowstore')}
              </div>
            </div>
            <div className={kpiCss.self}>
              <div className={kpiCss.title}>Memory used</div>
              <div className={kpiCss.value}>
                {props.isLoading && <Spin size="small" />}
                {!props.isLoading &&
                  numbro(selectedTable?.memory_use || 0).format({
                    output: 'byte',
                    base: 'binary',
                    mantissa: 2,
                    spaceSeparated: true
                  })}
              </div>
            </div>
            <div className={kpiCss.self}>
              <div className={kpiCss.title}>Rows</div>
              <div className={kpiCss.value}>
                {props.isLoading && <Spin size="small" />}
                {!props.isLoading &&
                  numbro(selectedTable?.rows || 0).format({
                    average: true,
                    mantissa: 2,
                    optionalMantissa: true
                  })}
              </div>
            </div>
          </Block>

          <Table
            className={blockCss.self}
            pagination={false}
            dataSource={columns}
            loading={props.isLoading}
            size="middle"
            rowKey="name"
            columns={[
              {
                key: 'position',
                title: '',
                sorter: AlphaSort('name'),
                sortDirections: ['ascend', 'descend'],
                render: (x: ColumnInformationSchema) => x.position
              },
              {
                key: 'name',
                title: 'Columns',
                sorter: AlphaSort('name'),
                sortDirections: ['ascend', 'descend'],
                render: (x: ColumnInformationSchema) => {
                  switch (x.column_key) {
                    case 'UNI':
                      return (
                        <>
                          {x.name}{' '}
                          <Tag className={CSS.margin_l_m} color="purple">
                            Unique
                          </Tag>
                        </>
                      )
                    case 'PRI':
                      if (selectedTable?.storage_type === 'COLUMNSTORE')
                        return (
                          <>
                            {x.name}{' '}
                            <Tag className={CSS.margin_l_m} color="gold">
                              Storage sort
                            </Tag>
                          </>
                        )
                      else
                        return (
                          <>
                            {x.name}{' '}
                            <Tag className={CSS.margin_l_m} color="magenta">
                              Primary
                            </Tag>
                          </>
                        )
                    case 'MUL':
                      return (
                        <>
                          {x.name}{' '}
                          <Tag className={CSS.margin_l_m} color="green">
                            Indexed
                          </Tag>
                        </>
                      )
                    default:
                      return x.name
                  }
                }
              },
              {
                key: 'desc',
                title: 'Description',
                render: (x: ColumnInformationSchema) => {
                  // is custom dimension
                  if (x.name.charAt(0) === '_') {
                    let extraColumn: any
                    workspaceCtx.workspace.installed_apps.forEach((app) => {
                      app.extra_columns?.forEach((colsManifest) => {
                        colsManifest.columns.forEach((col) => {
                          if (col.name === x.name) {
                            extraColumn = col
                          }
                        })
                      })
                    })
                    return (
                      <>
                        <Tag color="cyan">Extra column</Tag>
                        <br />
                        {extraColumn?.description}
                      </>
                    )
                  }

                  return get(TablesDescriptions, selectedTableName + '.columns.' + x.name, '')
                }
              },
              {
                key: 'type',
                title: 'Type & Charset',
                width: 300,
                render: (x: ColumnInformationSchema) => {
                  let defaultValue = null

                  switch (x.data_type) {
                    case 'double':
                      // return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR'}).format(Number(x.defaultValue))
                      if (x.default_value !== null && x.default_value !== undefined) {
                        defaultValue = numbro(x.default_value).format({
                          thousandSeparated: true
                        })
                      }
                      break
                    case 'tinyint':
                      if (x.default_value === '1') defaultValue = 'true'
                      if (x.default_value === '0') defaultValue = 'false'
                      break
                    default:
                      defaultValue = x.default_value
                  }

                  return (
                    <>
                      {x.column_type} {x.nullable === 'NO' && <Tag color="red">Not Null</Tag>}{' '}
                      {defaultValue !== null && 'default ' + defaultValue} {x.extra}{' '}
                      {x.character_set && <Tag color="geekblue">{x.character_set}</Tag>}
                    </>
                  )
                }
              }

              // {
              //   key: 'actions',
              //   width: 160,
              //   title: (
              //     <div className={GlobalCSS.text_right}>
              //       {props.isFetching && (
              //         <span className="margin-r-m">
              //           <Spin size="small" />
              //         </span>
              //       )}
              //       {TablesWithCustomColumns.includes(selectedTableName) && (
              //         <CreateCustomColumnsButton
              //           workspaceId={workspaceCtx.workspace.id}
              //           btnSize="small"
              //           btnType="primary"
              //           tableName={selectedTableName}
              //           customDimensions={workspaceCtx.workspace.custom_columns}
              //           btnContent={
              //             <>
              //               <FontAwesomeIcon icon={faPlus} />
              //               &nbsp; Add
              //             </>
              //           }
              //           apiPOST={workspaceCtx.apiPOST}
              //           onComplete={() => {
              //             workspaceCtx.refreshWorkspace()
              //             refetch()
              //           }}
              //         />
              //       )}
              //     </div>
              //   ),
              //   render: (x: ColumnInformationSchema) => (
              //     <div className={GlobalCSS.text_right}>
              //       {x.name.charAt(0) === '_' && selectedTable && (
              //         <DeleteCustomColumnsButton
              //           table={selectedTable.name}
              //           column={x.name}
              //           workspaceId={workspaceCtx.workspace.id}
              //           apiPOST={workspaceCtx.apiPOST}
              //           onComplete={() => {
              //             workspaceCtx.refreshWorkspace()
              //             refetch()
              //           }}
              //           btnSize="small"
              //         />
              //       )}
              //     </div>
              //   )
              // }
            ]}
          />
        </Col>
      </Row>

      {/* {!workspaceCtx.workspace.customDimensions.length && <div className='block-cta padding-v-l'>
            <p>In Rimdian, a Custom Dimension is a custom column persisted in the database tables.</p>
            <p>It is used to store extra data specific to your business case.</p>

            <div className="margin-t-l">
                <CreateCustomColumnsButton
                    workspaceId={workspaceCtx.workspace.id}
                    btnSize="middle"
                    btnType='primary'
                    customDimensions={workspaceCtx.workspace.customDimensions}
                    btnContent={<><FontAwesomeIcon icon={faPlus} />&nbsp; New custom dimension</>}
                    apiPOST={workspaceCtx.apiPOST}
                    onComplete={() => {
                        workspaceCtx.refreshWorkspace()
                    }}
                />
            </div>
        </div>} */}
    </div>
  )
}

export default BlockDBSchemas
