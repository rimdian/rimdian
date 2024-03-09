import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Layout from 'components/common/layout'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import CSS from 'utils/css'
import {
  faCircleCheck,
  faCirclePause,
  faListCheck,
  faTriangleExclamation
} from '@fortawesome/free-solid-svg-icons'
import UpsertCheckButton from './button_check'
import { ObservabilityCheck, ObservabilityIncident } from './interfaces'
import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { map } from 'lodash'
import Block from 'components/common/block'
import { Button, Popconfirm, Spin, Table, Tooltip, message } from 'antd'
import { faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import dayjs from 'dayjs'

const RouteObservability = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  // console.log('schemas', workspaceCtx.cubeSchemasMap)

  // checks
  const {
    isLoading: isLoadingChecks,
    data: checks,
    refetch: refetchChecks,
    isFetching: isFetchingChecks
  } = useQuery<ObservabilityCheck[]>(
    ['observability_checks', workspaceCtx.workspace.id],
    (): Promise<ObservabilityCheck[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET('/observabilityCheck.list?' + '&workspace_id=' + workspaceCtx.workspace.id)
          .then((data: any) => {
            // console.log(data)
            resolve(data as ObservabilityCheck[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  // active incidents
  const {
    isLoading: isLoadingIncidents,
    data: incidents,
    refetch: refetchIncidents,
    isFetching: isFetchingIncidents
  } = useQuery<ObservabilityIncident[]>(
    ['observability_incidents', workspaceCtx.workspace.id],
    (): Promise<ObservabilityIncident[]> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET('/observabilityIncident.list?' + '&workspace_id=' + workspaceCtx.workspace.id)
          .then((data: any) => {
            // console.log(data)
            resolve(data as ObservabilityIncident[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  const deleteCheck = (checkId: string) => {
    workspaceCtx
      .apiPOST('/observabilityCheck.delete', {
        workspace_id: workspaceCtx.workspace.id,
        id: checkId
      })
      .then(() => {
        refetchChecks()
        message.success('Check deleted')
      })
  }

  const checksGroupedByCube = useMemo(() => {
    const checksGroupedByCube: { [key: string]: ObservabilityCheck[] } = {}
    checks?.forEach((check) => {
      const [cubeName, _measureName] = check.measure.split('.')
      if (!checksGroupedByCube[cubeName]) checksGroupedByCube[cubeName] = []
      checksGroupedByCube[cubeName].push(check)
    })
    return checksGroupedByCube
  }, [checks])

  const activeIncidentsByCheckId = useMemo(() => {
    const activeIncidentsByCheckId: { [key: string]: ObservabilityIncident[] } = {}
    incidents?.forEach((incident) => {
      if (!activeIncidentsByCheckId[incident.check_id])
        activeIncidentsByCheckId[incident.check_id] = []
      activeIncidentsByCheckId[incident.check_id].push(incident)
    })
    return activeIncidentsByCheckId
  }, [incidents])

  // console.log('activeIncidentsByCheckId', activeIncidentsByCheckId)
  return (
    <Layout
      currentOrganization={workspaceCtx.organization}
      currentWorkspace={workspaceCtx.workspace}
    >
      <div className={CSS.container}>
        <div className={CSS.top}>
          <h1>Observability</h1>
          {checks && checks?.length > 0 && (
            <>
              <div className={CSS.topSeparator}></div>
              <UpsertCheckButton
                btnContent={<>Create a check</>}
                btnType="primary"
                onComplete={() => {
                  refetchChecks()
                }}
              />
            </>
          )}
        </div>

        {!isLoadingChecks && checks?.length === 0 && (
          <div className={CSS.emptyState.container}>
            <div className={CSS.emptyState.icon}>
              <FontAwesomeIcon icon={faListCheck} />
            </div>
            <div className={CSS.margin_b_xl}>
              <p>
                <b>No checks found.</b>
              </p>
              <p>
                Start monitoring your data volume, freshness & quality to make sure it's accurate
                and ready for business.
              </p>
            </div>

            <UpsertCheckButton
              btnContent={<>Create a check</>}
              btnType="primary"
              onComplete={() => {
                refetchChecks()
              }}
            />
          </div>
        )}

        {map(checksGroupedByCube, (checks, cubeName) => {
          return (
            <Block
              key={cubeName}
              small
              title={cubeName}
              // extra={
              //   <Button size="small" onClick={() => setRefreshAt(dayjs().unix())}>
              //     <FontAwesomeIcon spin={loading} icon={faRefresh} />
              //   </Button>
              // }
            >
              <Table
                size="middle"
                pagination={false}
                dataSource={checks}
                showHeader={false}
                rowKey="id"
                columns={[
                  {
                    title: 'icon',
                    key: 'icon',
                    width: 40,
                    render: (check: ObservabilityCheck) => {
                      if (!check.is_active)
                        return (
                          <FontAwesomeIcon
                            className={CSS.font_size_xl + ' ' + CSS.opacity_30}
                            icon={faCirclePause}
                          />
                        )

                      if (isLoadingIncidents || isFetchingIncidents) return <Spin size="small" />

                      if (activeIncidentsByCheckId[check.id]?.length > 0) {
                        return (
                          <FontAwesomeIcon
                            className={CSS.font_size_xl + ' ' + CSS.text_red}
                            icon={faTriangleExclamation}
                          />
                        )
                      }
                      return (
                        <FontAwesomeIcon
                          className={CSS.font_size_xl + ' ' + CSS.text_green}
                          icon={faCircleCheck}
                        />
                      )
                    }
                  },
                  {
                    title: 'Name',
                    key: 'name',
                    width: 300,
                    render: (check: ObservabilityCheck) => {
                      return (
                        <>
                          {check.name}

                          {activeIncidentsByCheckId[check.id]?.length > 0 && (
                            <>
                              <div
                                className={
                                  CSS.margin_t_xs +
                                  ' ' +
                                  CSS.font_size_xs +
                                  ' ' +
                                  CSS.font_weight_semibold +
                                  ' ' +
                                  CSS.text_red
                                }
                              >
                                Incident started:
                                {' ' +
                                  dayjs(
                                    activeIncidentsByCheckId[check.id][0].first_triggered_at
                                  ).fromNow()}
                              </div>
                              <div
                                className={
                                  CSS.font_size_xs +
                                  ' ' +
                                  CSS.font_weight_semibold +
                                  ' ' +
                                  CSS.padding_r_xs +
                                  ' ' +
                                  CSS.text_red
                                }
                              >
                                Last value:
                                {' ' +
                                  activeIncidentsByCheckId[check.id][0].value +
                                  ' ' +
                                  dayjs(
                                    activeIncidentsByCheckId[check.id][0].last_triggered_at
                                  ).fromNow()}
                              </div>
                            </>
                          )}
                        </>
                      )
                    }
                  },
                  {
                    title: 'rolling_window',
                    key: 'rolling_window',
                    width: 180,
                    render: (check: ObservabilityCheck) => {
                      return (
                        <>
                          <span
                            className={
                              CSS.font_size_xs +
                              ' ' +
                              CSS.font_weight_semibold +
                              ' ' +
                              CSS.padding_r_xs
                            }
                          >
                            Rolling window:
                          </span>

                          <span>
                            {check.rolling_window_value}
                            {check.rolling_window_unit}
                          </span>
                        </>
                      )
                    }
                  },
                  {
                    title: 'threshold',
                    key: 'threshold',
                    width: 180,
                    render: (check: ObservabilityCheck) => {
                      return (
                        <>
                          <span
                            className={
                              CSS.font_size_xs +
                              ' ' +
                              CSS.font_weight_semibold +
                              ' ' +
                              CSS.padding_r_xs
                            }
                          >
                            Threshold:
                          </span>

                          <span>{check.threshold_position + ' ' + check.threshold_value}</span>
                        </>
                      )
                    }
                  },
                  {
                    title: 'next_run',
                    key: 'next_run',
                    render: (check: ObservabilityCheck) => {
                      return (
                        <>
                          <span
                            className={
                              CSS.font_size_xs +
                              ' ' +
                              CSS.font_weight_semibold +
                              ' ' +
                              CSS.padding_r_xs
                            }
                          >
                            Next run:
                          </span>

                          <span>{dayjs(check.next_run_at).fromNow()}</span>
                        </>
                      )
                    }
                  },
                  {
                    title: 'emails',
                    key: 'emails',
                    render: (check: ObservabilityCheck) => {
                      if (check.emails.length === 0) return <></>
                      return (
                        <>
                          <span
                            className={
                              CSS.font_size_xs +
                              ' ' +
                              CSS.font_weight_semibold +
                              ' ' +
                              CSS.padding_r_xs
                            }
                          >
                            Notify:
                          </span>
                          <span>{check.emails.join(', ')}</span>
                        </>
                      )
                    }
                  },
                  {
                    title: 'actions',
                    key: 'actions',
                    width: 100,
                    render: (check: ObservabilityCheck) => {
                      return (
                        <Button.Group>
                          <Tooltip title="Delete check" placement="left">
                            <Popconfirm
                              title="Do you really want to delete this check & its eventual incidents?"
                              okButtonProps={{ danger: true }}
                              okText="Delete"
                              cancelText="No"
                              onConfirm={() => {
                                deleteCheck(check.id)
                              }}
                            >
                              <Button type="text" size="small">
                                <FontAwesomeIcon icon={faTrashCan} />
                              </Button>
                            </Popconfirm>
                          </Tooltip>
                          <UpsertCheckButton
                            observabilityCheck={check}
                            btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
                            btnType="text"
                            btnSize="small"
                            onComplete={() => {
                              refetchChecks()
                            }}
                          />
                        </Button.Group>
                      )
                      // if (!check.is_active)
                      //   return (
                      //     <Button size="small" type="primary" ghost>
                      //       Activate
                      //     </Button>
                      //   )
                      // return check.is_active ? <FontAwesomeIcon icon={faCircleCheck} /> : 'No'
                    }
                  }
                ]}
              />
            </Block>
          )
        })}
      </div>
    </Layout>
  )
}

export default RouteObservability
