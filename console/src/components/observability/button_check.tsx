import { useContext, useEffect, useMemo, useState } from 'react'
import {
  Table,
  Drawer,
  Form,
  Input,
  Button,
  Select,
  Modal,
  message,
  Space,
  Tag,
  InputNumber,
  Tooltip,
  Popover,
  Alert,
  Spin,
  Switch
} from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faInfoCircle, faXmark } from '@fortawesome/free-solid-svg-icons'
import { ButtonType } from 'antd/lib/button'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import { ObservabilityCheck, ObservabilityCheckFilter } from './interfaces'
import CSS from 'utils/css'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { CubeSchema, CubeSchemaDimension } from 'interfaces'
import { find, forEach, map } from 'lodash'
import { CubeContext } from '@cubejs-client/react'
import { ResultSet, TimeDimension } from '@cubejs-client/core'
import dayjs from 'dayjs'
import { useAccount } from 'components/login/context_account'
import ReactECharts from 'echarts-for-react'
import { SeriesOption } from 'echarts'

type AddFilterButtonProps = {
  cube?: CubeSchema
  cubeName?: string
  onComplete: (origin: ObservabilityCheckFilter) => void
}

const AddFilterButton = (props: AddFilterButtonProps) => {
  const [form] = Form.useForm()
  const [modalVisible, setModalVisible] = useState(false)

  const onClicked = () => {
    setModalVisible(true)
  }

  // console.log('cube', props.cube)
  return (
    <>
      <Button disabled={props.cube ? false : true} ghost type="primary" block onClick={onClicked}>
        Add
      </Button>
      <Modal
        open={modalVisible}
        title="Add a filter"
        okText="Confirm"
        width={600}
        cancelText="Cancel"
        onCancel={() => {
          setModalVisible(false)
        }}
        onOk={() => {
          form
            .validateFields()
            .then((values: ObservabilityCheckFilter) => {
              form.resetFields()
              setModalVisible(false)
              props.onComplete(values)
            })
            .catch(console.error)
        }}
      >
        <Form form={form} name="form_add_filter" labelCol={{ span: 8 }} wrapperCol={{ span: 14 }}>
          <Form.Item
            label="Dimension"
            name="dimension"
            rules={[
              { required: true, type: 'string' }
              // { validator: mappingValidator }
            ]}
          >
            <Select
              options={map(props.cube?.dimensions, (dimension: any, key) => {
                return {
                  value: props.cubeName + '.' + key,
                  label: <Tooltip title={dimension.title}>{props.cubeName + '.' + key}</Tooltip>
                }
              })}
              onChange={() => {
                // reset fields on change
                form.setFieldsValue({
                  operator: undefined,
                  string_values: undefined,
                  number_values: undefined
                })
              }}
            />
          </Form.Item>

          <Form.Item label="Operator" required dependencies={['dimension']}>
            {(funcs) => {
              const opts: any[] = []
              const dimension: CubeSchemaDimension | undefined = find(
                props.cube?.dimensions,
                (_col, key) => props.cubeName + '.' + key === funcs.getFieldValue('dimension')
              ) as CubeSchemaDimension | undefined

              // string columns
              if (dimension) {
                if (dimension.type === 'string') {
                  opts.push(
                    { value: 'equals', label: 'equals' },
                    { value: 'notEquals', label: 'not equals' },
                    { value: 'set', label: 'is set' },
                    { value: 'notSet', label: 'is not set' },
                    { value: 'startsWith', label: 'starts with' },
                    { value: 'endsWith', label: 'ends with' },
                    { value: 'contains', label: 'contains' },
                    { value: 'notContains', label: 'not contains' }
                  )
                }

                // numbers and dates have same operators
                if (dimension.type === 'number') {
                  opts.push(
                    { value: 'gt', label: 'greater than' },
                    { value: 'gte', label: 'greater than or equals' },
                    { value: 'lt', label: 'less than' },
                    { value: 'lte', label: 'less than or equals' },
                    { value: 'equals', label: 'equals' },
                    { value: 'notEquals', label: 'not equals' },
                    { value: 'set', label: 'is set' },
                    { value: 'notSet', label: 'is not set' }
                  )
                }

                if (dimension.type === 'time') {
                  opts.push(
                    { value: 'inDateRange', label: 'in date range' },
                    { value: 'notInDateRange', label: 'not in date range' },
                    { value: 'beforeDate', label: 'before date' },
                    { value: 'afterDate', label: 'after date' },
                    { value: 'set', label: 'is set' },
                    { value: 'notSet', label: 'is not set' }
                  )
                }

                // boolean
                if (dimension.type === 'boolean') {
                  opts.push(
                    { value: 'equals', label: 'equals' },
                    { value: 'notEquals', label: 'not equals' },
                    { value: 'set', label: 'is set' },
                    { value: 'notSet', label: 'is not set' }
                  )
                }
              }

              return (
                <Form.Item noStyle name="operator" rules={[{ required: true, type: 'string' }]}>
                  <Select placeholder="Select an operator" options={opts} />
                </Form.Item>
              )
            }}
          </Form.Item>

          <Form.Item noStyle dependencies={['dimension', 'operator']}>
            {(funcs) => {
              const operator = funcs.getFieldValue('operator')
              // ignore if operators dont need values
              if (!operator || ['set', 'notSet'].includes(operator)) return <></>

              const dimension: CubeSchemaDimension | undefined = find(
                props.cube?.dimensions,
                (_col, key) => props.cubeName + '.' + key === funcs.getFieldValue('dimension')
              ) as CubeSchemaDimension | undefined
              if (!dimension || dimension.type === 'boolean') return <></>

              const rule: any = { required: true, type: 'array', min: 1, max: 1 }
              // BETWEEN operator nees 2 values
              if (['inDateRange', 'notInDateRange'].includes(operator)) {
                rule.min = 2
                rule.max = 2
                rule.defaultField = { type: 'date' }
              }
              if (['beforeDate', 'afterDate'].includes(operator)) {
                rule.defaultField = { type: 'date' }
              }
              // LIKE operator has no max
              if (['contains', 'notContains'].includes(operator)) {
                rule.max = undefined
              }

              return (
                <Form.Item label="Value(s)" name="values" rules={[rule]}>
                  <Select
                    mode="tags"
                    placeholder={
                      'Input a value & press enter' + (dimension.type === 'time' && ' (YYYY-MM-DD)')
                    }
                  />
                </Form.Item>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

type FiltersInputProps = {
  cube?: CubeSchema
  cubeName?: string
  onChange?: (filters: ObservabilityCheckFilter[]) => void
  value?: ObservabilityCheckFilter[]
}

const FiltersInput = (props: FiltersInputProps) => {
  const removeObservabilityCheckFilter = (index: number) => {
    const filters = props.value ? props.value.slice() : []
    filters.splice(index, 1)
    props.onChange?.(filters)
  }

  // console.log(props.cube)

  return (
    <div>
      {props.value && props.value.length > 0 && (
        <Table
          size="small"
          bordered={false}
          pagination={false}
          rowKey={(record) => record.dimension + record.operator + record.values?.join(', ')}
          showHeader={false}
          className={CSS.margin_b_m}
          columns={[
            {
              title: '',
              key: 'dimension',
              render: (item: ObservabilityCheckFilter) => {
                return (
                  <span>
                    <Space>
                      <span>{item.dimension}</span>
                      <Tag color="blue">
                        <b>{item.operator}</b>
                      </Tag>
                      {item.values && item.values.length > 0 && (
                        <span>{item.values?.join(', ')}</span>
                      )}
                    </Space>
                  </span>
                )
              }
            },
            {
              title: '',
              key: 'remove',
              render: (_text, _record: any, index: number) => {
                return (
                  <div className={CSS.text_right}>
                    <Button
                      type="dashed"
                      size="small"
                      onClick={removeObservabilityCheckFilter.bind(null, index)}
                    >
                      <FontAwesomeIcon icon={faXmark} />
                    </Button>
                  </div>
                )
              }
            }
          ]}
          dataSource={props.value}
        />
      )}

      <AddFilterButton
        cube={props.cube}
        cubeName={props.cubeName}
        onComplete={(values: any) => {
          const filters = props.value ? props.value.slice() : []
          filters.push(values)
          props.onChange?.(filters)
        }}
      />
    </div>
  )
}

type GraphPreviewProps = {
  timezone: string
  timeDimension: string
  measure: string
  filters: ObservabilityCheckFilter[]
  rolling_window_value: number
  rolling_window_unit: 'm' | 'h' | 'd' // minute, hour, day
  threshold_position: string
  threshold_value?: number
}

const GraphPreview = (props: GraphPreviewProps) => {
  const { cubejsApi } = useContext(CubeContext)
  const [isLoading, setIsLoading] = useState(true)
  // const [executedSQL, setExecutedSQL] = useState<ExecutedSQL | undefined>(undefined)
  const [error, setError] = useState<string | undefined>(undefined)
  const [value, setValue] = useState<any[] | undefined>(undefined)

  // compute the date range to get a maximum of 15 data points according to the rolling window unit
  const max_data_points = 15

  const now = dayjs().tz(props.timezone).startOf('day').format('YYYY-MM-DD')
  const fromDate = dayjs()
    .tz(props.timezone)
    .subtract(max_data_points * props.rolling_window_value, props.rolling_window_unit)
    .startOf('day')
    .format('YYYY-MM-DD')

  let granularity = 'day'
  if (props.rolling_window_unit === 'm') granularity = 'minute'
  if (props.rolling_window_unit === 'h') granularity = 'hour'

  const graphQuery = useMemo(() => {
    return {
      measures: [props.measure],
      filters: props.filters?.map((f) => {
        return {
          member: f.dimension,
          operator: f.operator as any,
          values: f.values
        }
      }),
      timeDimensions: [
        {
          dimension: props.timeDimension, // TODO: use the time dimension of the metric
          // "second" | "minute" | "hour" | "day" | "week" | "month" | "quarter" | "year"
          granularity: granularity,
          dateRange: [fromDate, now]
        }
      ] as TimeDimension[],
      timezone: props.timezone
    }
  }, [
    props.measure,
    props.filters,
    props.timeDimension,
    props.timezone,
    props.rolling_window_value,
    props.rolling_window_unit
  ])

  const series = [] as SeriesOption[]
  const lineData: number[] = []
  const dates: string[] = []
  const markLines: any[] = []

  if (props.threshold_value !== undefined) {
    markLines.push({
      name: 'Threshold ' + props.threshold_position,
      yAxis: props.threshold_value,
      label: {
        normal: {
          show: true,
          formatter: '{b}: {c}',
          backgroundColor: '#E91E63',
          padding: [4, 8],
          borderRadius: 4,
          color: '#fff',
          fontWeight: 'bold',
          fontSize: 11,
          position: 'insideEndTop'
        }
      },
      lineStyle: {
        normal: {
          color: '#E91E63',
          width: 1
        }
      }
    })
  }

  // generate the dates for the x axis, according to the rolling window unit
  for (let i = 1; i <= max_data_points; i++) {
    // console.log('sub', max_data_points - i)
    switch (granularity) {
      case 'day':
        dates.push(
          dayjs()
            .tz(props.timezone)
            .subtract(max_data_points - i, props.rolling_window_unit)
            .startOf('day')
            .format('MMM D')
        )
        break
      case 'hour':
        dates.push(
          dayjs()
            .tz(props.timezone)
            .subtract(max_data_points - i, props.rolling_window_unit)
            .startOf('hour')
            .format('MMM D, HH:mm')
        )
        break
      case 'minute':
        dates.push(
          dayjs()
            .tz(props.timezone)
            .subtract(max_data_points - i, props.rolling_window_unit)
            .startOf('minute')
            .format('MMM D, HH:mm')
        )
        break
      default:
    }

    lineData.push(0)
  }

  // console.log('props.rolling_window_unit', props.rolling_window_unit)
  // console.log('props.timezone', props.timezone)
  // console.log('dates', dates)

  value?.forEach((item: any, i: number) => {
    lineData[i] = item.value
    // dates.push(i + '')
  })

  series.push({
    silent: true, // disable hover and cursor:pointer
    type: 'line',
    symbol: 'none',
    cursor: 'default',
    z: 3,
    lineStyle: { type: 'solid', color: '#3D5AFE', width: 1 },
    data: lineData,
    markLine: {
      data: markLines,
      symbol: 'pin',
      label: {
        distance: [20, 8]
      }
    }
  } as SeriesOption)

  // fetch
  useEffect(() => {
    // console.log('KPI: fetch', totalQuery, graphQuery)
    setIsLoading(true)

    Promise.all([
      cubejsApi.sql(graphQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'sql' }),
      cubejsApi.load(graphQuery, { mutexObj: window.MutexCubeJS, mutexKey: 'load' })
    ])
      .then(([sqlQuery, resultSet]: any[]) => {
        const result = resultSet as ResultSet
        // console.log('result', result)
        // console.log('result.series()', result.series())
        setIsLoading(false)
        setError(undefined)
        setValue(result.series()[0]?.series || [])
        console.info('sqlQuery', sqlQuery.sqlQuery.sql.sql)
        // setExecutedSQL({
        //   name: 'Graph',
        //   sql: sqlQuery[0].sqlQuery.sql.sql[0],
        //   args: sqlQuery[0].sqlQuery.sql.sql[1]
        // })
      })
      .catch((error) => {
        setError(error.toString())
        setIsLoading(false)
      })
  }, [graphQuery, cubejsApi])
  return (
    <Spin spinning={isLoading}>
      {error && <Alert message={error} type="error" />}
      <ReactECharts
        option={{
          animationDuration: 150,
          renderer: 'svg',
          grid: {
            top: 20,
            left: 70,
            right: 50,
            bottom: 20, // required to see the bottom line
            show: false,
            containLabel: false
          },
          tooltip: {
            trigger: 'axis'
            // axisPointer: {
            //   type: 'cross'
            // }
          },
          xAxis: [
            {
              type: 'category',
              // boundaryGap: false,
              data: dates,
              // axisLabel: { show: true },
              // axisTick: { show: true },
              axisLine: {
                show: true,
                lineStyle: {
                  color: 'rgba(0,0,0,0.3)',
                  width: 2,
                  type: 'solid'
                }
              }
              // axisTick: { show: false }
            }
          ],
          yAxis: [
            {
              type: 'value',
              min:
                props.threshold_value !== undefined && props.threshold_value < 0
                  ? (props.threshold_value * 1.2).toFixed(2)
                  : undefined,
              max:
                props.threshold_value !== undefined && props.threshold_value > 0
                  ? (props.threshold_value * 1.2).toFixed(2)
                  : undefined,
              axisLabel: { show: true },
              axisLine: {
                show: true,
                lineStyle: {
                  color: 'rgba(0,0,0,0.3)',
                  width: 2,
                  type: 'solid'
                }
              },
              axisTick: { show: true },
              splitLine: { show: true }
            }
          ],
          series: series
        }}
        className="echart"
        style={{ height: 170, cursor: 'default !important' }}
      />
    </Spin>
  )
}

type UpsertCheckButtonProps = {
  observabilityCheck?: ObservabilityCheck
  btnContent: JSX.Element
  onComplete: () => void
  btnType?: ButtonType
  btnSize?: SizeType
}

const UpsertCheckButton = (props: UpsertCheckButtonProps) => {
  const [form] = Form.useForm()
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const workspaceCtx = useCurrentWorkspaceCtx()
  const accountCtx = useAccount()

  const closeDrawer = () => {
    form?.resetFields()
    setDrawerVisible(false)
  }

  const measures = useMemo(() => {
    const options: any[] = []
    forEach(workspaceCtx.cubeSchemasMap, (schema, key) => {
      forEach(schema.measures as any, (def, measure) => {
        const metric = key + '.' + measure
        options.push({
          value: metric,
          label: (
            <Popover zIndex={1500} title={def.title} content={def.description} placement="left">
              <FontAwesomeIcon
                className={CSS.text_blue + ' ' + CSS.opacity_50 + ' ' + CSS.margin_r_xs}
                icon={faInfoCircle}
              />{' '}
              {metric}
            </Popover>
          )
        })
      })
    })

    return options
  }, [workspaceCtx.cubeSchemasMap])

  const onFinish = (values: any) => {
    // console.log('values', values);

    if (loading) return

    setLoading(true)

    if (props.observabilityCheck) {
      values.id = props.observabilityCheck.id
    }

    values.workspace_id = workspaceCtx.workspace.id

    workspaceCtx
      .apiPOST(
        props.observabilityCheck ? '/observabilityCheck.update' : '/observabilityCheck.create',
        values
      )
      .then(() => {
        if (props.observabilityCheck) {
          message.success('The check has successfully been updated.')
        } else {
          message.success('The check has successfully been created.')
          form.resetFields()
        }

        setLoading(false)
        setDrawerVisible(false)
        props.onComplete()
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  const initialValues = Object.assign(
    {
      filters: [],
      rolling_window_value: 1,
      rolling_window_unit: 'h',
      threshold_position: 'below'
    },
    props.observabilityCheck
  )

  // console.log('initialValues', initialValues)
  // console.log('props.observabilityCheck', props.observabilityCheck)
  if (!initialValues.filters) initialValues.filters = []

  // to control which field update will re-render the block: dependencies={[...]}
  // to display an input without a "grid" decoration: noStyle
  // to access fresh values inside a Form.Item: {(funcs) => { const values = funcs.getFieldsValue(); ... }}
  return (
    <>
      <Button type={props.btnType} size={props.btnSize} onClick={() => setDrawerVisible(true)}>
        {props.btnContent}
      </Button>
      {drawerVisible && (
        <Drawer
          title={props.observabilityCheck ? 'Update check' : 'Create a new check'}
          width={800}
          open={true}
          onClose={closeDrawer}
          bodyStyle={{ paddingBottom: 80 }}
          extra={
            <Space>
              <Button loading={loading} onClick={closeDrawer}>
                Cancel
              </Button>
              <Button
                loading={loading}
                onClick={() => {
                  form.submit()
                }}
                type="primary"
              >
                Confirm
              </Button>
            </Space>
          }
        >
          <Form
            form={form}
            initialValues={initialValues}
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 18 }}
            name="groupForm"
            onFinish={onFinish}
          >
            <Form.Item name="name" label="Name" rules={[{ required: true, type: 'string' }]}>
              <Input />
            </Form.Item>

            <Form.Item name="measure" label="Measure" rules={[{ required: true, type: 'string' }]}>
              <Select
                showSearch
                options={measures}
                onChange={(value) => {
                  const cubeName = value?.split('.')[0]
                  const cube = find(workspaceCtx.cubeSchemasMap, (_x, k) =>
                    value?.startsWith(k + '.')
                  )
                  if (!cube) {
                    form.setFieldsValue({ time_dimension: undefined })
                    return
                  }
                  // select the time dimension:
                  const time_dimensions = [
                    'created_at_trunc',
                    'received_at_trunc', // Data_import cube
                    'created_at',
                    'db_created_at'
                  ]
                  const time_dimension = time_dimensions.find((k) => cube.dimensions[k])
                  if (time_dimension === undefined) {
                    form.setFieldsValue({ time_dimension: undefined })
                    return
                  }
                  form.setFieldsValue({ time_dimension: cubeName + '.' + time_dimension })
                }}
              />
            </Form.Item>

            <Form.Item label="Time dimension" required dependencies={['measure']}>
              {(_funcs) => {
                return (
                  <Form.Item
                    name="time_dimension"
                    noStyle
                    rules={[{ required: true, type: 'string' }]}
                  >
                    <Input disabled />
                  </Form.Item>
                )
              }}
            </Form.Item>

            <Form.Item label="Filters" dependencies={['measure']}>
              {(funcs) => {
                let cube = undefined
                let cubeName = undefined
                forEach(workspaceCtx.cubeSchemasMap, (_x, k) => {
                  if (funcs.getFieldValue('measure')?.startsWith(k + '.')) {
                    cube = workspaceCtx.cubeSchemasMap[k]
                    cubeName = k
                  }
                })

                return (
                  <Form.Item
                    name="filters"
                    noStyle
                    rules={[{ required: false, type: 'array', min: 0 }]}
                  >
                    <FiltersInput cube={cube} cubeName={cubeName} />
                  </Form.Item>
                )
              }}
            </Form.Item>

            <Form.Item label="Rolling window" required dependencies={['rolling_window_unit']}>
              {(funcs) => {
                const unit = funcs.getFieldValue('rolling_window_unit')
                const rule: any = { required: true, type: 'number' }
                const inputProps: any = {}
                if (unit === 'm') {
                  rule.min = 15
                  rule.max = 60
                  inputProps['step'] = 15
                }
                if (unit === 'h') {
                  rule.min = 1
                  rule.max = 24
                  inputProps['step'] = 1
                }
                if (unit === 'd') {
                  rule.min = 1
                  rule.max = 7
                  inputProps['step'] = 1
                }

                return (
                  <Form.Item noStyle name="rolling_window_value" rules={[rule]}>
                    <InputNumber
                      {...inputProps}
                      placeholder="Every..."
                      rules={[rule]}
                      addonAfter={
                        <Form.Item
                          noStyle
                          name="rolling_window_unit"
                          rules={[{ required: true, type: 'string' }]}
                        >
                          <Select
                            style={{ width: '150px' }}
                            options={[
                              { value: 'm', label: 'minute' },
                              { value: 'h', label: 'hour' },
                              { value: 'd', label: 'day' }
                            ]}
                            onChange={(value) => {
                              if (value === 'm') form.setFieldsValue({ rolling_window_value: 15 })
                              if (value === 'h') form.setFieldsValue({ rolling_window_value: 1 })
                              if (value === 'd') form.setFieldsValue({ rolling_window_value: 1 })
                            }}
                          />
                        </Form.Item>
                      }
                    />
                  </Form.Item>
                )
              }}
            </Form.Item>

            <Form.Item label="Threshold" required>
              <Form.Item
                noStyle
                name="threshold_value"
                rules={[{ required: true, type: 'number' }]}
              >
                <InputNumber
                  addonBefore={
                    <Form.Item
                      noStyle
                      name="threshold_position"
                      rules={[{ required: true, type: 'string' }]}
                    >
                      <Select
                        style={{ width: '150px' }}
                        options={[
                          { value: 'below', label: 'Below value' },
                          { value: 'above', label: 'Above value' }
                        ]}
                        // optionType="button"
                      />
                    </Form.Item>
                  }
                  placeholder="Enter a value..."
                />
              </Form.Item>
            </Form.Item>

            <Form.Item
              label="Emails to alert"
              name="emails"
              rules={[{ required: false, type: 'array', defaultField: { type: 'email' } }]}
            >
              <Select mode="tags" placeholder="Input an email & press enter" />
            </Form.Item>

            <Form.Item
              label="Activate check"
              name="is_active"
              rules={[{ required: false, type: 'boolean' }]}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>

            <Form.Item noStyle dependencies={['measure', 'filters', 'threshold_value']}>
              {(funcs) => {
                if (!funcs.getFieldValue('measure')) return <></>
                return (
                  <div className={CSS.margin_t_xl}>
                    <GraphPreview
                      measure={funcs.getFieldValue('measure')}
                      timezone={accountCtx.account?.account.timezone || 'UTC'}
                      timeDimension={funcs.getFieldValue('time_dimension')}
                      filters={funcs.getFieldValue('filters')}
                      rolling_window_value={funcs.getFieldValue('rolling_window_value')}
                      rolling_window_unit={funcs.getFieldValue('rolling_window_unit')}
                      threshold_position={funcs.getFieldValue('threshold_position')}
                      threshold_value={funcs.getFieldValue('threshold_value')}
                    />
                  </div>
                )
              }}
            </Form.Item>
          </Form>
        </Drawer>
      )}
    </>
  )
}

export default UpsertCheckButton
