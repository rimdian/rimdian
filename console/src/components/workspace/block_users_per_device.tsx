import { ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Row, Table, Col } from 'antd'
import { Query } from '@cubejs-client/core'
import numbro from 'numbro'
import ReactECharts from 'echarts-for-react'
import FormatGrowth from 'utils/format_growth'
import { PaletteCarbon } from 'utils/colors'
import Block from 'components/common/block'
import CSS from 'utils/css'
import { useRimdianCube } from './context_cube'

export type UsersPerDeviceProps = {
  workspaceId: string
  timezone: string
  refreshKey: string
  dateFrom: string
  dateTo: string
  dateFromPrevious: string
  dateToPrevious: string
  tooltip?: ReactNode | string
}

export const UsersPerDevice = (props: UsersPerDeviceProps) => {
  const { cubeApi } = useRimdianCube()
  const refreshKeyRef = useRef('')
  const [loading, setLoading] = useState<boolean>(true)
  const [tableData, setTableData] = useState<any[]>([])
  const [error, setError] = useState<string | undefined>(undefined)

  const query = useMemo(() => {
    return {
      measures: ['Session.unique_users'],
      dimensions: ['Device.device_type'],
      filters: [
        {
          member: 'Device.device_type',
          operator: 'set'
        }
      ],
      timeDimensions: [
        {
          dimension: 'Session.created_at_trunc',
          granularity: undefined, // total count
          compareDateRange: [
            [props.dateFrom, props.dateTo],
            [props.dateFromPrevious, props.dateToPrevious]
          ]
        }
      ],
      timezone: props.timezone,
      renewQuery: props.refreshKey !== refreshKeyRef.current
    }
  }, [
    props.dateFrom,
    props.dateTo,
    props.timezone,
    props.dateFromPrevious,
    props.dateToPrevious,
    props.refreshKey
  ])

  useEffect(() => {
    if (props.refreshKey === refreshKeyRef.current) {
      return
    } else {
      refreshKeyRef.current = props.refreshKey
    }

    setLoading(true)

    cubeApi
      .load(query as Query)
      .then((resultSet) => {
        const [currentSet, previousSet] = resultSet.decompose()

        // use .pivot() as tablePivot() returns the measure named with the date range
        // which is boring to process
        const previousPivot = previousSet.pivot({
          x: ['Device.device_type'],
          y: ['Session.unique_users', 'measures']
        })

        const currentSetPivot = currentSet.pivot({
          x: ['Device.device_type'],
          y: ['Session.unique_users', 'measures']
        })

        let inc = 0
        let previousValue
        const data: any[] = []

        currentSetPivot.forEach((row) => {
          inc += row.yValuesArray[0][1]
        })

        currentSetPivot.forEach((row) => {
          previousValue = previousPivot.find((prevRow) => prevRow.xValues[0] === row.xValues[0])

          data.push({
            name: row.xValues[0],
            value: row.yValuesArray[0][1],
            previousValue: previousValue ? previousValue.yValuesArray[0][1] : 0,
            percentage: row.yValuesArray[0][1] === 0 ? 0 : row.yValuesArray[0][1] / inc,
            growth: previousValue
              ? (row.yValuesArray[0][1] - previousValue.yValuesArray[0][1]) /
                previousValue.yValuesArray[0][1]
              : 100
          })
        })

        setTableData(data)
        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [query, cubeApi, props.refreshKey])

  // debug sql query
  //   cubeApi
  //     .sql(query)
  //     .then((q: any) => {
  //       console.log(q[0].sqlQuery.sql.sql)
  //     })
  //     .catch((error) => {
  //       console.log(error)
  //     })

  let title = 'Users per device'

  if (error) {
    title = `${title} - ${error.toString()}`
  }

  return (
    <Block title={title} small classNames={[CSS.margin_t_l]}>
      <Row gutter={24}>
        <Col span={12}>
          <Table
            rowKey="name"
            className="small-table"
            size="small"
            showHeader={false}
            loading={loading}
            dataSource={tableData}
            pagination={{
              pageSize: 5,
              showLessItems: true,
              hideOnSinglePage: true,
              showSizeChanger: false
            }}
            columns={[
              {
                title: '',
                key: 'name',
                render: (x: any) => {
                  return <span style={{ textTransform: 'capitalize' }}>{x.name}</span>
                }
              },
              {
                title: '',
                key: 'percentage',
                width: 70,
                className: 'text-right',
                sorter: (a: any, b: any) => {
                  if (a.percentage < b.percentage) {
                    return -1
                  }
                  if (a.percentage > b.percentage) {
                    return 1
                  }
                  return 0
                },
                sortOrder: 'descend',
                render: (x: any) =>
                  numbro(x.percentage).format({
                    average: false,
                    output: 'percent',
                    mantissa: 2,
                    optionalMantissa: true
                  })
              },
              {
                title: '',
                key: 'diff',
                width: 80,
                className: 'text-right',
                render: (x: any) => FormatGrowth(x.value, x.previousValue)
              }
            ]}
          />
        </Col>
        <Col span={12}>
          {!loading && (
            <ReactECharts
              style={{ height: 150 }}
              option={{
                color: PaletteCarbon,
                tooltip: {
                  trigger: 'item'
                },
                series: [
                  {
                    name: 'device',
                    type: 'pie',
                    top: 12,
                    bottom: 12,
                    right: 12,
                    radius: ['50%', '70%'],
                    // avoidLabelOverlap: false,
                    itemStyle: {
                      borderRadius: 4,
                      borderColor: '#fff',
                      borderWidth: 2
                    },
                    label: {
                      position: 'outer',
                      alignTo: 'labelLine',
                      bleedMargin: 5
                    },
                    data: tableData
                  }
                ]
              }}
            />
          )}
        </Col>
      </Row>
    </Block>
  )
}
