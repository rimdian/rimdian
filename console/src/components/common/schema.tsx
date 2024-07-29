import { Graph } from '@datastructures-js/graph'
import { Popover, Table } from 'antd'
import { CubeSchema, CubeSchemaDimension, CubeSchemaMap, CubeSchemaMeasure } from 'interfaces'
import { forEach } from 'lodash'
import CSS from 'utils/css'
import FormatCurrency from 'utils/format_currency'
import FormatGrowth from 'utils/format_growth'
import FormatNumber from 'utils/format_number'
import FormatPercent from 'utils/format_percent'

export interface DimensionDefinition {
  dimensionName: string
  cubeName: string
  cube: CubeSchema
  dimension: CubeSchemaDimension
  customRender?: (values: any[], currency: string) => JSX.Element | string
}

export interface MeasureDefinition {
  key: string
  measureName: string
  cubeName: string
  cube: CubeSchema
  measure: CubeSchemaMeasure
  customRender?: (currentPeriod: any, previousPeriod: any, currency: string) => JSX.Element | string
  dependsOnMeasures?: string[]
}

export const generateDimensionsMap = (
  graph: Graph<string, boolean>,
  cubeSchemasMap: CubeSchemaMap
) => {
  const result = {} as Record<string, DimensionDefinition>

  forEach(cubeSchemasMap, (cubeSchema, cubeName) => {
    // check if has vertex in graph
    if (!graph.hasVertex(cubeName.toLowerCase())) {
      return
    }

    // add dimensions
    forEach(cubeSchema.dimensions, (dimension, name) => {
      // only accept strings and booleans for now
      if (dimension.type === 'string' || dimension.type === 'boolean') {
        const k = `${cubeName}.${name}`
        result[k] = {
          dimensionName: name,
          cubeName: cubeName,
          cube: cubeSchema,
          dimension: dimension
        } as DimensionDefinition
      }
    })
  })

  return result
}
export const generateMeasuresMap = (
  graph: Graph<string, boolean>,
  cubeSchemasMap: CubeSchemaMap
) => {
  const result = {} as Record<string, MeasureDefinition>

  forEach(cubeSchemasMap, (cubeSchema, cubeName) => {
    // check if has vertex in graph
    if (!graph.hasVertex(cubeName.toLowerCase())) {
      return
    }

    // add measures
    forEach(cubeSchema.measures, (dimension, name) => {
      const k = `${cubeName}.${name}`
      result[k] = {
        key: k,
        measureName: name,
        cubeName: cubeName,
        cube: cubeSchema,
        measure: dimension
      } as MeasureDefinition
    })
  })

  return result
}

export const generateDatabaseGraphForSchema = (
  schemaName: string,
  cubeSchemasMap: CubeSchemaMap
): Graph<string, boolean> => {
  const g = new Graph<string, boolean>()

  if (!cubeSchemasMap[schemaName]) {
    return g
  }

  // recursively add tables linked to tables
  const addTables = (cubeName: string) => {
    // abort if schema does not exist
    if (!cubeSchemasMap[cubeName]) {
      console.error('Schema not found', cubeName)
      return
    }

    const schema = cubeSchemasMap[cubeName]
    const tableName = cubeName.toLowerCase()

    // abort if table already exists
    if (g.hasVertex(tableName)) return

    // add vertex
    g.addVertex(tableName, true)

    // add edges
    if (schema.joins) {
      forEach(schema.joins, (_join, linkedCubeName) => {
        if (!g.hasVertex(linkedCubeName.toLowerCase())) {
          addTables(linkedCubeName) // recursive call
        }

        g.addEdge(tableName, linkedCubeName.toLowerCase())
      })
    }
  }

  addTables(schemaName)

  // add single-way relationships from apps
  forEach(cubeSchemasMap, (cubeSchema, cubeName) => {
    if (cubeSchema.joins) {
      forEach(cubeSchema.joins, (join, linkedCubeName) => {
        // if cube is in the graph, and the relationship is one-to-one or many_to_one
        if (
          g.hasVertex(linkedCubeName.toLowerCase()) &&
          (join.relationship === 'one_to_one' || join.relationship === 'many_to_one')
        ) {
          // add missing tables
          if (!g.hasVertex(cubeName.toLowerCase())) {
            addTables(cubeName) // recursive call
          }

          g.addEdge(cubeName.toLowerCase(), linkedCubeName.toLowerCase())
        }
      })
    }
  })

  return g
}

export const AttributionRoleMeasure = {
  key: 'Session.attribution_roles',
  measureName: 'Session.attribution_roles',
  cubeName: 'Session',
  cube: {
    title: 'Session'
  } as CubeSchema,
  measure: {} as CubeSchemaMeasure,
  dependsOnMeasures: [
    'Session.alone_count',
    'Session.alone_ratio',
    'Session.initiator_count',
    'Session.initiator_ratio',
    'Session.assisting_count',
    'Session.assisting_ratio',
    'Session.closer_count',
    'Session.closer_ratio',
    'Session.alone_linear_conversions_attributed',
    'Session.alone_linear_amount_attributed',
    'Session.initiator_linear_conversions_attributed',
    'Session.initiator_linear_amount_attributed',
    'Session.assisting_linear_conversions_attributed',
    'Session.assisting_linear_amount_attributed',
    'Session.closer_linear_conversions_attributed',
    'Session.closer_linear_amount_attributed'
  ],
  customRender: (currentPeriod: any, previousPeriod: any, currency: string) => {
    const data = [
      {
        key: 'alone',
        title: 'Alone',
        width: (currentPeriod['Session.alone_ratio'] || 0) * 100 + 'px',
        color: '#607D8B',
        contributions: currentPeriod['Session.alone_count'] || 0,
        contributionsRatio: currentPeriod['Session.alone_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.alone_ratio'] || 0,
        linearConversions: currentPeriod['Session.alone_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.alone_linear_amount_attributed'] || 0
      },
      {
        key: 'initiator',
        title: 'Initiator',
        width: (currentPeriod['Session.initiator_ratio'] || 0) * 100 + 'px',
        color: '#00BCD4',
        contributions: currentPeriod['Session.initiator_count'] || 0,
        contributionsRatio: currentPeriod['Session.initiator_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.initiator_ratio'] || 0,
        linearConversions: currentPeriod['Session.initiator_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.initiator_linear_amount_attributed'] || 0
      },
      {
        key: 'assisting',
        title: 'Assisting',
        width: (currentPeriod['Session.assisting_ratio'] || 0) * 100 + 'px',
        color: '#CDDC39',
        contributions: currentPeriod['Session.assisting_count'] || 0,
        contributionsRatio: currentPeriod['Session.assisting_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.assisting_ratio'] || 0,
        linearConversions: currentPeriod['Session.assisting_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.assisting_linear_amount_attributed'] || 0
      },
      {
        key: 'closer',
        title: 'Closer',
        width: (currentPeriod['Session.closer_ratio'] || 0) * 100 + 'px',
        color: '#F06292',
        contributions: currentPeriod['Session.closer_count'] || 0,
        contributionsRatio: currentPeriod['Session.closer_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.closer_ratio'] || 0,
        linearConversions: currentPeriod['Session.closer_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.closer_linear_amount_attributed'] || 0
      }
    ]

    const content = (
      <Table
        rowKey="key"
        dataSource={data}
        size="small"
        pagination={false}
        columns={[
          {
            key: 'title',
            title: '',
            render: (record: any) => record.title
          },
          {
            key: 'bar',
            title: '',
            render: (record: any) => (
              <div>
                <span
                  style={{
                    width: record.width,
                    display: 'inline-block',
                    backgroundColor: record.color,
                    height: '5px'
                  }}
                ></span>
                <div className={CSS.font_size_xxs}>
                  {FormatGrowth(record.contributionsRatio, record.previousContributionsRatio)}
                </div>
              </div>
            )
          },
          {
            key: 'contributionRatio',
            title: '',
            render: (record: any) => FormatPercent(record.contributionsRatio)
          },
          {
            key: 'contributions',
            title: 'Contributions',
            render: (record: any) => FormatNumber(record.contributions)
          },
          {
            key: 'linearConversions',
            title: 'Linear conversions',
            render: (record: any) => FormatNumber(record.linearConversions)
          },
          {
            key: 'linearRevenue',
            title: 'Linear revenue',
            render: (record: any) => FormatCurrency(record.linearRevenue, currency, { light: true })
          }
        ]}
      />
    )
    return (
      <Popover
        content={content}
        title={null}
        trigger={['hover', 'click']}
        placement="left"
        className={CSS.padding_v_m}
      >
        <div style={{ cursor: 'help', width: '100px' }}>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.alone_ratio'] || 0) * 100 + '%',
              backgroundColor: '#607D8B',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.initiator_ratio'] || 0) * 100 + '%',
              backgroundColor: '#00BCD4',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.assisting_ratio'] || 0) * 100 + '%',
              backgroundColor: '#CDDC39',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.closer_ratio'] || 0) * 100 + '%',
              backgroundColor: '#F06292',
              height: '3px'
            }}
          ></div>
        </div>
      </Popover>
    )
  }
} as MeasureDefinition

export const AcquisitionAttributionRoleMeasure = {
  key: 'Session.acquisition_attribution_roles',
  measureName: 'Session.acquisition_attribution_roles',
  cubeName: 'Session',
  cube: {
    title: 'Session'
  } as CubeSchema,
  measure: {
    title: 'Acquisition: roles'
  } as CubeSchemaMeasure,
  dependsOnMeasures: [
    'Session.acquisition_alone_count',
    'Session.acquisition_alone_ratio',
    'Session.acquisition_initiator_count',
    'Session.acquisition_initiator_ratio',
    'Session.acquisition_assisting_count',
    'Session.acquisition_assisting_ratio',
    'Session.acquisition_closer_count',
    'Session.acquisition_closer_ratio',
    'Session.acquisition_alone_linear_conversions_attributed',
    'Session.acquisition_alone_linear_amount_attributed',
    'Session.acquisition_initiator_linear_conversions_attributed',
    'Session.acquisition_initiator_linear_amount_attributed',
    'Session.acquisition_assisting_linear_conversions_attributed',
    'Session.acquisition_assisting_linear_amount_attributed',
    'Session.acquisition_closer_linear_conversions_attributed',
    'Session.acquisition_closer_linear_amount_attributed'
  ],
  customRender: (currentPeriod: any, previousPeriod: any, currency: string) => {
    const data = [
      {
        key: 'alone',
        title: 'Alone',
        width: (currentPeriod['Session.acquisition_alone_ratio'] || 0) * 100 + 'px',
        color: '#607D8B',
        contributions: currentPeriod['Session.acquisition_alone_count'] || 0,
        contributionsRatio: currentPeriod['Session.acquisition_alone_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.acquisition_alone_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.acquisition_alone_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.acquisition_alone_linear_amount_attributed'] || 0
      },
      {
        key: 'initiator',
        title: 'Initiator',
        width: (currentPeriod['Session.acquisition_initiator_ratio'] || 0) * 100 + 'px',
        color: '#00BCD4',
        contributions: currentPeriod['Session.acquisition_initiator_count'] || 0,
        contributionsRatio: currentPeriod['Session.acquisition_initiator_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.acquisition_initiator_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.acquisition_initiator_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.acquisition_initiator_linear_amount_attributed'] || 0
      },
      {
        key: 'assisting',
        title: 'Assisting',
        width: (currentPeriod['Session.acquisition_assisting_ratio'] || 0) * 100 + 'px',
        color: '#CDDC39',
        contributions: currentPeriod['Session.acquisition_assisting_count'] || 0,
        contributionsRatio: currentPeriod['Session.acquisition_assisting_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.acquisition_assisting_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.acquisition_assisting_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.acquisition_assisting_linear_amount_attributed'] || 0
      },
      {
        key: 'closer',
        title: 'Closer',
        width: (currentPeriod['Session.acquisition_closer_ratio'] || 0) * 100 + 'px',
        color: '#F06292',
        contributions: currentPeriod['Session.acquisition_closer_count'] || 0,
        contributionsRatio: currentPeriod['Session.acquisition_closer_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.acquisition_closer_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.acquisition_closer_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.acquisition_closer_linear_amount_attributed'] || 0
      }
    ]

    const content = (
      <Table
        rowKey="key"
        dataSource={data}
        size="small"
        pagination={false}
        columns={[
          {
            key: 'title',
            title: '',
            render: (record: any) => record.title
          },
          {
            key: 'bar',
            title: '',
            render: (record: any) => (
              <div>
                <span
                  style={{
                    width: record.width,
                    display: 'inline-block',
                    backgroundColor: record.color,
                    height: '5px'
                  }}
                ></span>
                <div className={CSS.font_size_xxs}>
                  {FormatGrowth(record.contributionsRatio, record.previousContributionsRatio)}
                </div>
              </div>
            )
          },
          {
            key: 'contributionRatio',
            title: '',
            render: (record: any) => FormatPercent(record.contributionsRatio)
          },
          {
            key: 'contributions',
            title: 'Contributions',
            render: (record: any) => FormatNumber(record.contributions)
          },
          {
            key: 'linearConversions',
            title: 'Linear conversions',
            render: (record: any) => FormatNumber(record.linearConversions)
          },
          {
            key: 'linearRevenue',
            title: 'Linear revenue',
            render: (record: any) => FormatCurrency(record.linearRevenue, currency, { light: true })
          }
        ]}
      />
    )
    return (
      <Popover
        content={content}
        title={null}
        trigger={['hover', 'click']}
        placement="left"
        className={CSS.padding_v_m}
      >
        <div style={{ cursor: 'help', width: '100px' }}>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.acquisition_alone_ratio'] || 0) * 100 + '%',
              backgroundColor: '#607D8B',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.acquisition_initiator_ratio'] || 0) * 100 + '%',
              backgroundColor: '#00BCD4',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.acquisition_assisting_ratio'] || 0) * 100 + '%',
              backgroundColor: '#CDDC39',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.acquisition_closer_ratio'] || 0) * 100 + '%',
              backgroundColor: '#F06292',
              height: '3px'
            }}
          ></div>
        </div>
      </Popover>
    )
  }
} as MeasureDefinition

export const RetentionAttributionRoleMeasure = {
  key: 'Session.retention_attribution_roles',
  measureName: 'Session.retention_attribution_roles',
  cubeName: 'Session',
  cube: {
    title: 'Session'
  } as CubeSchema,
  measure: {
    title: 'Retention: roles'
  } as CubeSchemaMeasure,
  dependsOnMeasures: [
    'Session.retention_alone_count',
    'Session.retention_alone_ratio',
    'Session.retention_initiator_count',
    'Session.retention_initiator_ratio',
    'Session.retention_assisting_count',
    'Session.retention_assisting_ratio',
    'Session.retention_closer_count',
    'Session.retention_closer_ratio',
    'Session.retention_alone_linear_conversions_attributed',
    'Session.retention_alone_linear_amount_attributed',
    'Session.retention_initiator_linear_conversions_attributed',
    'Session.retention_initiator_linear_amount_attributed',
    'Session.retention_assisting_linear_conversions_attributed',
    'Session.retention_assisting_linear_amount_attributed',
    'Session.retention_closer_linear_conversions_attributed',
    'Session.retention_closer_linear_amount_attributed'
  ],
  customRender: (currentPeriod: any, previousPeriod: any, currency: string) => {
    const data = [
      {
        key: 'alone',
        title: 'Alone',
        width: (currentPeriod['Session.retention_alone_ratio'] || 0) * 100 + 'px',
        color: '#607D8B',
        contributions: currentPeriod['Session.retention_alone_count'] || 0,
        contributionsRatio: currentPeriod['Session.retention_alone_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.retention_alone_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.retention_alone_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.retention_alone_linear_amount_attributed'] || 0
      },
      {
        key: 'initiator',
        title: 'Initiator',
        width: (currentPeriod['Session.retention_initiator_ratio'] || 0) * 100 + 'px',
        color: '#00BCD4',
        contributions: currentPeriod['Session.retention_initiator_count'] || 0,
        contributionsRatio: currentPeriod['Session.retention_initiator_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.retention_initiator_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.retention_initiator_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.retention_initiator_linear_amount_attributed'] || 0
      },
      {
        key: 'assisting',
        title: 'Assisting',
        width: (currentPeriod['Session.retention_assisting_ratio'] || 0) * 100 + 'px',
        color: '#CDDC39',
        contributions: currentPeriod['Session.retention_assisting_count'] || 0,
        contributionsRatio: currentPeriod['Session.retention_assisting_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.retention_assisting_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.retention_assisting_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.retention_assisting_linear_amount_attributed'] || 0
      },
      {
        key: 'closer',
        title: 'Closer',
        width: (currentPeriod['Session.retention_closer_ratio'] || 0) * 100 + 'px',
        color: '#F06292',
        contributions: currentPeriod['Session.retention_closer_count'] || 0,
        contributionsRatio: currentPeriod['Session.retention_closer_ratio'] || 0,
        previousContributionsRatio: previousPeriod['Session.retention_closer_ratio'] || 0,
        linearConversions:
          currentPeriod['Session.retention_closer_linear_conversions_attributed'] || 0,
        linearRevenue: currentPeriod['Session.retention_closer_linear_amount_attributed'] || 0
      }
    ]

    const content = (
      <Table
        rowKey="key"
        dataSource={data}
        size="small"
        pagination={false}
        columns={[
          {
            key: 'title',
            title: '',
            render: (record: any) => record.title
          },
          {
            key: 'bar',
            title: '',
            render: (record: any) => (
              <div>
                <span
                  style={{
                    width: record.width,
                    display: 'inline-block',
                    backgroundColor: record.color,
                    height: '5px'
                  }}
                ></span>
                <div className={CSS.font_size_xxs}>
                  {FormatGrowth(record.contributionsRatio, record.previousContributionsRatio)}
                </div>
              </div>
            )
          },
          {
            key: 'contributionRatio',
            title: '',
            render: (record: any) => FormatPercent(record.contributionsRatio)
          },
          {
            key: 'contributions',
            title: 'Contributions',
            render: (record: any) => FormatNumber(record.contributions)
          },
          {
            key: 'linearConversions',
            title: 'Linear conversions',
            render: (record: any) => FormatNumber(record.linearConversions)
          },
          {
            key: 'linearRevenue',
            title: 'Linear revenue',
            render: (record: any) => FormatCurrency(record.linearRevenue, currency, { light: true })
          }
        ]}
      />
    )
    return (
      <Popover
        content={content}
        title={null}
        trigger={['hover', 'click']}
        placement="left"
        className={CSS.padding_v_m}
      >
        <div style={{ cursor: 'help', width: '100px' }}>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.retention_alone_ratio'] || 0) * 100 + '%',
              backgroundColor: '#607D8B',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.retention_initiator_ratio'] || 0) * 100 + '%',
              backgroundColor: '#00BCD4',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.retention_assisting_ratio'] || 0) * 100 + '%',
              backgroundColor: '#CDDC39',
              height: '3px'
            }}
          ></div>
          <div
            style={{
              marginBottom: '2px',
              width: (currentPeriod['Session.retention_closer_ratio'] || 0) * 100 + '%',
              backgroundColor: '#F06292',
              height: '3px'
            }}
          ></div>
        </div>
      </Popover>
    )
  }
} as MeasureDefinition
