import { ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { Row, Table, Col } from 'antd'
import numbro from 'numbro'
import ReactECharts from 'echarts-for-react'
import * as echarts from 'echarts'
import { CountriesMap } from 'utils/countries_timezones'
import FormatGrowth from 'utils/format_growth'
import worldMap from 'utils/world_map'
import Block from 'components/common/block'
import CSS from 'utils/css'
import { useRimdianCube } from './context_cube'

echarts.registerMap('world', worldMap as any)

export type UsersPerCountryProps = {
  workspaceId: string
  timezone: string
  refreshKey: string
  dateFrom: string
  dateTo: string
  dateFromPrevious: string
  dateToPrevious: string
  tooltip?: ReactNode | string
}

export const UsersPerCountry = (props: UsersPerCountryProps) => {
  const { cubeApi } = useRimdianCube()
  const refreshKeyRef = useRef('')
  const [loading, setLoading] = useState<boolean>(true)
  const [tableData, setTableData] = useState<any[]>([])
  const [total, setTotal] = useState<number>(0)
  const [error, setError] = useState<string | undefined>(undefined)

  const query = useMemo(() => {
    return {
      measures: ['Session.unique_users'],
      dimensions: ['User.country'],
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

    // console.log('query', query)
    setLoading(true)

    cubeApi
      .load(query)
      .then((resultSet) => {
        const [currentSet, previousSet] = resultSet.decompose()

        // use .pivot() as tablePivot() returns the measure named with the date range
        // which is boring to process
        const previousPivot = previousSet.pivot({
          x: ['User.country'],
          y: ['Session.unique_users', 'measures']
        })

        const currentSetPivot = currentSet.pivot({
          x: ['User.country'],
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

        setTotal(inc)
        setTableData(data)
        setLoading(false)
        setError(undefined)
      })
      .catch((error) => {
        setError(error.toString())
      })
  }, [query, cubeApi, props.refreshKey])

  let title = 'Users per country'

  if (error) {
    title = `${title} - ${error.toString()}`
  }

  return (
    <Block title={title} small classNames={[CSS.margin_t_l]}>
      <Row gutter={24}>
        <Col span={12}>
          <Table
            rowKey="name"
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
                  return CountriesMap[x.name] ? CountriesMap[x.name].name : 'unknown...'
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
              style={{ height: 200 }}
              option={{
                tooltip: {
                  trigger: 'item',
                  formatter: (params: any) => {
                    // only accepts a string, could be HTML in string
                    return `${params.name}: ${
                      params.data
                        ? numbro(params.data.percentage).format({
                            average: false,
                            output: 'percent',
                            mantissa: 2,
                            optionalMantissa: true
                          })
                        : 0
                    }`
                  }
                },
                geo: {
                  map: 'world',
                  left: 0,
                  top: 12,
                  right: 12,
                  bottom: 12,
                  scaleLimit: {
                    min: 0,
                    max: total
                  },
                  label: {
                    show: false,
                    color: '#4E6CFF',
                    fontSize: 10
                  },
                  roam: false,
                  zoom: 1,
                  itemStyle: {
                    areaColor: '#FFFFFF',
                    borderColor: '#CFD8DC',
                    borderWidth: 1
                  },
                  emphasis: {
                    itemStyle: {
                      areaColor: '#4E6CFF',
                      shadowOffsetX: 0,
                      shadowOffsetY: 0,
                      shadowBlur: 5,
                      borderWidth: 0,
                      // borderColor: '#4E6CFF',
                      shadowColor: 'rgba(0, 0, 0, 0.1)'
                    },
                    label: {
                      show: false,
                      color: '#4E6CFF',
                      fontSize: 10
                    }
                  },
                  zlevel: 1
                },

                visualMap: {
                  min: 0,
                  max: total,
                  show: false,
                  inRange: {
                    color: [
                      'rgba(78, 108, 255, 0.3)',
                      'rgba(78, 108, 255, 0.4)',
                      'rgba(78, 108, 255, 0.5)',
                      'rgba(78, 108, 255, 0.6)',
                      'rgba(78, 108, 255, 0.7)',
                      'rgba(78, 108, 255, 0.8)',
                      'rgba(78, 108, 255, 0.9)',
                      'rgba(78, 108, 255, 1)'
                    ]
                  }
                },
                series: [
                  {
                    name: 'country',
                    type: 'map',
                    geoIndex: 0,
                    data: tableData,
                    zlevel: 3
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
