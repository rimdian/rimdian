import { Alert, Badge, Button, Modal, Popover, Progress, Spin } from 'antd'
import { Workspace, TaskExec, TaskExecList } from 'interfaces'
import { difference, get, toArray } from 'lodash'
import dayjs from 'dayjs'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import ButtonTaskAbout from './button_about'
import { css } from '@emotion/css'
import CSS, { shadowBase } from 'utils/css'
import { useEffect, useState } from 'react'

type BadgeRunningTasksProps = {
  organizationId: string
  workspaceId: string
  workspaces: Workspace[]
  accountTimezone: string
  refetchSegments: () => Promise<any>
  apiGET: (endpoint: string) => Promise<any>
}

const badgeCSS = {
  self: css({
    backgroundColor: '#F3F6FC',
    position: 'fixed',
    bottom: CSS.L,
    right: CSS.L,
    boxShadow: shadowBase
  })
}

export const BadgeRunningTasks = (props: BadgeRunningTasksProps) => {
  const navigate = useNavigate()
  const [refreshRate, setRefreshRate] = useState(5) // every 5 secs by default
  const [refreshingSegments, setRefreshingSegments] = useState<string[]>([])

  // running tasks
  const { isLoading, data, refetch, isFetching } = useQuery<TaskExec[]>(
    ['workspace', props.workspaceId, 'runningTasks'],
    (): Promise<TaskExec[]> => {
      return new Promise((resolve, reject) => {
        props
          .apiGET('/taskExec.list?status=0&workspace_id=' + props.workspaceId)
          .then((data: TaskExecList) => {
            // check if a task is recomputing segments
            const segmentBeingRefreshed = data.task_execs
              .filter((x: TaskExec) => x.task_id === 'recompute_segment')
              .map((x: TaskExec) => x.id)

            // check if a new segment is being refreshed or a segment is done refreshing
            if (difference(refreshingSegments, segmentBeingRefreshed).length > 0) {
              // console.log('refreshing segments')
              props.refetchSegments()
            }

            setRefreshingSegments(segmentBeingRefreshed)

            // increase refresh rate when a task is running
            if (data && data.task_execs?.length > 0) {
              if (data.task_execs.find((x: any) => !x.is_done) && refreshRate > 5) {
                setRefreshRate(5)
              } else if (refreshRate < 30) {
                setRefreshRate(30)
              }
            }
            resolve(data.task_execs as TaskExec[])
          })
          .catch((e) => {
            reject(e)
          })
      })
    },
    {
      enabled: props.workspaceId && props.workspaces.length > 0 ? true : false
    }
  )

  // create a dynamic interval to refresh data
  useEffect(() => {
    const interval = setInterval(() => {
      refetch()
    }, refreshRate * 1000)
    return () => clearInterval(interval)
  }, [refreshRate])

  // refresh data on open
  const onOpenChange = (visible: boolean) => {
    if (visible) {
      refetch()
    }
  }

  if (data && data.length > 0) {
    const demoTask = data.find(
      (x) => x.task_id === 'system_generate_demo' && (x.status === 0 || x.status === -1)
    )

    if (demoTask) {
      const workers = toArray(get(demoTask, 'state.workers', {}))
      const status = get(workers, '[0].status', 'unknown')
      let totalDataLogs = 0
      const processedDataLogs = get(workers, '[0].processed_data_logs', 0)

      if (workers.length > 0) {
        workers.forEach((w, i) => {
          if (i > 0) {
            totalDataLogs += w.total_data_logs
          }
        }, 0)
      }

      return (
        <Modal
          open={true}
          closable={false}
          footer={false}
          centered={true}
          title={
            <>
              Generating demo...
              <span className={CSS.pull_right}>
                <Spin size="small" spinning={isFetching} />
              </span>
            </>
          }
        >
          <Alert
            type="info"
            showIcon={true}
            className={CSS.margin_b_l}
            message="We are actually generating thousands of fake user visits... it might take 2 minutes."
          />
          <p>
            <b>Task ID:</b> {demoTask.id}
            <span className={CSS.pull_right}>
              <ButtonTaskAbout
                workspaceId={props.workspaceId}
                apiGET={props.apiGET}
                taskExec={demoTask}
                accountTimezone={props.accountTimezone}
              />
            </span>
          </p>
          <p>
            <b>Status:</b> {status}
          </p>
          <p>
            <b>Last update:</b> {dayjs(demoTask.db_updated_at).fromNow()} -{' '}
            {dayjs(demoTask.db_created_at).tz(props.accountTimezone).format('lll')} in{' '}
            {props.accountTimezone}
          </p>
          {status === 'loading' && workers.length > 1 && (
            <>
              <p>
                <b>Workers:</b>
              </p>
              <table>
                <tbody>
                  {workers.map((w: any, k: number) => {
                    if (k === 0) return ''
                    const totalDays: number = w.to_day - w.from_day
                    const daysProcessed = w.current_day - w.from_day
                    const percent = daysProcessed === 0 ? 0 : (daysProcessed * 100) / totalDays
                    return (
                      <tr key={k}>
                        <td style={{ width: 40 }}>{k}</td>
                        <td>
                          <Progress
                            percent={percent}
                            strokeColor={{ from: '#00C9FF', to: '#92FE9D' }}
                            size="small"
                          />
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </>
          )}
          {status === 'processing' && (
            <>
              <p>
                <b>Data logs processed: </b> {processedDataLogs}/{totalDataLogs}
              </p>
              <Progress
                percent={Math.floor((processedDataLogs * 100) / totalDataLogs)}
                strokeColor={{ from: '#00C9FF', to: '#92FE9D' }}
                size="small"
              />
            </>
          )}
        </Modal>
      )
    }

    return (
      <Popover
        onOpenChange={onOpenChange}
        title="Running tasks"
        content={
          <Spin spinning={isLoading || isFetching}>
            {data.map((x: TaskExec) => (
              <table className={CSS.margin_v_s} key={x.id}>
                <tbody>
                  <tr>
                    <th style={{ padding: '2px 20px 2px 0' }}>{x.name}</th>
                    <td style={{ padding: '2px 20px 2px 0' }}>
                      Started {dayjs(x.db_created_at).fromNow()}
                    </td>
                    <td>
                      <ButtonTaskAbout
                        workspaceId={props.workspaceId}
                        apiGET={props.apiGET}
                        taskExec={x}
                        accountTimezone={props.accountTimezone}
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            ))}
          </Spin>
        }
        placement="topRight"
      >
        <div className={badgeCSS.self}>
          <Badge count={data.length} color="#03A9F4">
            <Button
              type="primary"
              ghost
              onClick={() =>
                navigate(
                  '/orgs/' +
                    props.organizationId +
                    '/workspaces/' +
                    props.workspaceId +
                    '/system/tasks?isDone=false'
                )
              }
            >
              <Spin size="small" />
              {/* <span style={{fontSize: 14, animation: 'loadingCircle 1s infinite linear'}}>
                            <svg viewBox="0 0 1024 1024" focusable="false" data-icon="loading" width="1em" height="1em" fill="currentColor" aria-hidden="true"><path d="M988 548c-19.9 0-36-16.1-36-36 0-59.4-11.6-117-34.6-171.3a440.45 440.45 0 00-94.3-139.9 437.71 437.71 0 00-139.9-94.3C629 83.6 571.4 72 512 72c-19.9 0-36-16.1-36-36s16.1-36 36-36c69.1 0 136.2 13.5 199.3 40.3C772.3 66 827 103 874 150c47 47 83.9 101.8 109.7 162.7 26.7 63.1 40.2 130.2 40.2 199.3.1 19.9-16 36-35.9 36z"></path></svg>
                        </span> */}
              {/* <FontAwesomeIcon color='#4E6CFF' icon={faForward} /> */}
            </Button>
          </Badge>
        </div>
      </Popover>
    )
  }

  return <></>
}
