import { useState } from 'react'
import { Button, Drawer, Tooltip, Tag, Table, Spin, message, Space, Modal, Alert } from 'antd'
import { App, Task, TaskExec, TaskExecList, TaskList } from 'interfaces'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { useQuery } from '@tanstack/react-query'
import { useAccount } from 'components/login/context_account'
import dayjs from 'dayjs'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCircleCheck, faPlay, faRefresh } from '@fortawesome/free-solid-svg-icons'
import ButtonTaskAbout from 'components/task_exec/button_about'
import { faTimesCircle } from '@fortawesome/free-regular-svg-icons'
import CSS, { backgroundColorBase } from 'utils/css'
import { css } from '@emotion/css'
import Block from 'components/common/block'
import ButtonAbortTask from 'components/task_exec/button_abort'
import TextArea from 'antd/lib/input/TextArea'

type TasksAppButtonProps = {
  app: App
  workspaceCtx: CurrentWorkspaceCtxValue
}

const TasksAppButton = (props: TasksAppButtonProps) => {
  const [drawerVisible, setDrawerVisible] = useState(false)
  const accountCtx = useAccount()

  const closeDrawer = () => {
    setDrawerVisible(false)
  }

  // tasks
  const {
    isLoading: loadingTasks,
    data: tasks,
    refetch: refetchTasks,
    isFetching: isFetchingTasks
  } = useQuery<TaskList>(
    ['tasks', props.workspaceCtx.workspace.id, props.app.id],
    (): Promise<TaskList> => {
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiGET(
            '/task.list?workspace_id=' + props.workspaceCtx.workspace.id + '&app_id=' + props.app.id
          )
          .then((data: any) => {
            // console.log('data', data)
            resolve(data as TaskList)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  // task execs
  const {
    isLoading: loadingTaskExecs,
    data: taskExecs,
    refetch: refetchTaskExecs,
    isFetching: isFetchingTaskExecs
  } = useQuery<TaskExecList>(
    ['task_execs', props.workspaceCtx.workspace.id, props.app.id],
    (): Promise<TaskExecList> => {
      return new Promise((resolve, reject) => {
        props.workspaceCtx
          .apiGET(
            '/taskExec.list?workspace_id=' +
              props.workspaceCtx.workspace.id +
              '&app_id=' +
              props.app.id
          )
          .then((data: any) => {
            // console.log('task_execs', data)
            resolve(data as TaskExecList)
          })
          .catch((e) => {
            reject(e)
          })
      })
    }
  )

  // console.log('initialValues', initialValues);
  const alertTypeFromStatus = (status: number) => {
    if (status === -2) {
      return 'error'
    }
    if (status === -1) {
      return 'warning'
    }
    if (status === 0) {
      return 'info'
    }
    if (status === 1) {
      return 'success'
    }
    return 'info'
  }

  return (
    <>
      <Button
        type="primary"
        ghost
        size="small"
        onClick={() => {
          setDrawerVisible(true)
          refetchTasks()
          refetchTaskExecs()
        }}
      >
        Tasks
      </Button>
      {drawerVisible && (
        <Drawer
          title={
            <>
              <img
                src={props.app.manifest.icon_url}
                className={css(CSS.appIcon, CSS.margin_r_m)}
                style={{ height: 30 }}
                alt=""
              />
              Tasks {props.app.manifest.name}
            </>
          }
          width={960}
          open={true}
          onClose={closeDrawer}
          headerStyle={{ backgroundColor: backgroundColorBase }}
          bodyStyle={{ backgroundColor: backgroundColorBase }}
        >
          <Block
            title="Tasks"
            extra={
              <>
                <Button
                  size="small"
                  onClick={() => {
                    refetchTasks()
                  }}
                >
                  <FontAwesomeIcon icon={faRefresh} />
                </Button>
              </>
            }
          >
            <Table
              // showHeader={false}
              pagination={false}
              size="middle"
              dataSource={props.app.manifest.tasks}
              loading={loadingTasks || isFetchingTasks}
              rowKey="id"
              columns={[
                {
                  title: 'Name',
                  key: 'key',
                  // className: 'text-right',
                  render: (record: Task) => {
                    return <Tooltip title={record.id}>{record.name}</Tooltip>
                  }
                },
                {
                  title: 'Active',
                  key: 'value',
                  render: (record: Task) => {
                    const st = tasks?.tasks.find((c) => c.id === record.id)
                    if (!st) {
                      // task not installed
                      return 'not installed'
                    }

                    if (st.is_active) {
                      return <FontAwesomeIcon icon={faCircleCheck} className={CSS.text_green} />
                    } else {
                      return <FontAwesomeIcon icon={faTimesCircle} className={CSS.text_red} />
                    }
                  }
                },
                {
                  title: 'On multiple exec',
                  key: 'value',
                  render: (record: Task) => record.on_multiple_exec.replace('_', ' ')
                },
                {
                  title: 'Cron',
                  key: 'value',
                  render: (record: Task) => {
                    if (!record.is_cron) return '-'
                    return 'Every ' + record.minutes_interval + ' mins'
                  }
                },
                {
                  title: 'Last run',
                  key: 'last_run',
                  render: (record: Task) => {
                    // find last run from tasks
                    const st = tasks?.tasks.find((c) => c.id === record.id)
                    if (!st) {
                      // task not installed
                      return ''
                    }

                    if (!st.last_run) {
                      return <span>never</span>
                    }

                    return (
                      <Tooltip
                        title={
                          dayjs(st.last_run)
                            .tz(accountCtx.account?.account.timezone as string)
                            .format('lll') +
                          ' in ' +
                          accountCtx.account?.account.timezone
                        }
                      >
                        <span>{dayjs(st.last_run).fromNow()}</span>
                      </Tooltip>
                    )
                  }
                },
                // next run
                {
                  title: 'Next run',
                  key: 'next_run',
                  render: (record: Task) => {
                    if (isFetchingTasks || loadingTasks) {
                      return <Spin size="small" />
                    }

                    // find last run from tasks
                    const st = tasks?.tasks.find((c) => c.id === record.id)
                    if (!st) {
                      // st not installed
                      return <b className={CSS.text_red}>not installed</b>
                    }

                    if (!st.is_cron || !st.next_run) {
                      return <span>never</span>
                    }

                    return (
                      <Tooltip
                        title={
                          dayjs(st.next_run)
                            .tz(accountCtx.account?.account.timezone as string)
                            .format('lll') +
                          ' in ' +
                          accountCtx.account?.account.timezone
                        }
                      >
                        <span>{dayjs(st.next_run).fromNow()}</span>
                      </Tooltip>
                    )
                  }
                },
                {
                  title: '',
                  key: 'force',
                  className: CSS.text_right,
                  render: (record: Task) => {
                    // find last run from tasks
                    const st = tasks?.tasks.find((c) => c.id === record.id)
                    if (!st) {
                      // task not installed
                      return ''
                    }

                    if (st.is_active) {
                      return (
                        <RunButton
                          task={st}
                          workspaceCtx={props.workspaceCtx}
                          refetchTasks={refetchTasks}
                          refetchTaskExecs={refetchTaskExecs}
                        />
                      )
                    }
                    return ''
                  }
                }
              ]}
            />
          </Block>

          <Block
            extra={
              <>
                <Button
                  size="small"
                  loading={loadingTaskExecs || isFetchingTaskExecs}
                  onClick={() => {
                    refetchTaskExecs()
                  }}
                >
                  <FontAwesomeIcon icon={faRefresh} />
                </Button>
              </>
            }
            title="Executed tasks"
            classNames={[CSS.margin_t_xl]}
          >
            <Table
              pagination={false}
              dataSource={taskExecs?.task_execs}
              loading={loadingTaskExecs || isFetchingTaskExecs}
              size="small"
              // onChange={onTableChange}
              rowKey="id"
              columns={[
                {
                  title: 'Task name',
                  key: 'kind',
                  render: (x) => (
                    <div>
                      {/* {x.kind === 'webhook' && <Tag color="blue">Webhook</Tag>} */}
                      <Tooltip title={x.id + ' - ' + x.kind}>{x.name}</Tooltip>
                      {x.message && (
                        <Alert
                          message={<small style={{ wordBreak: 'break-all' }}>{x.message}</small>}
                          type={alertTypeFromStatus(x.status)}
                          className={CSS.margin_v_m}
                        />
                      )}
                    </div>
                  )
                },
                {
                  title: 'Status',
                  key: 'status',
                  render: (x: TaskExec) => {
                    let tag = <></>
                    if (x.status) {
                      if (x.status === -2) {
                        tag = <Tag color="red">Aborted</Tag>
                      }
                      if (x.status === -1) {
                        tag = <Tag color="orange">Retrying error...</Tag>
                      }
                      if (x.status === 0) {
                        tag = <Tag color="blue">Processing...</Tag>
                      }
                      if (x.status === 1) {
                        tag = <Tag color="green">Done</Tag>
                      }
                    }
                    return tag
                  }
                },
                {
                  title: 'Created',
                  key: 'createdAt',
                  render: (x: TaskExec) =>
                    dayjs(x.db_created_at)
                      .tz(accountCtx.account?.account.timezone as string)
                      .format('lll')
                },
                {
                  title: 'Updated',
                  key: 'updatedAt',
                  render: (x: TaskExec) => (
                    <Tooltip
                      title={
                        <span>
                          {dayjs(x.db_updated_at)
                            .tz(accountCtx.account?.account.timezone as string)
                            .format('lll')}{' '}
                          in {accountCtx.account?.account.timezone}
                        </span>
                      }
                    >
                      {dayjs(x.db_updated_at)
                        .tz(accountCtx.account?.account.timezone as string)
                        .fromNow()}
                    </Tooltip>
                  )
                },
                {
                  title: '',
                  key: 'actions',
                  className: 'actions',
                  width: 150,
                  render: (row: TaskExec) => (
                    <div className={CSS.text_right}>
                      <Space size="small">
                        <ButtonAbortTask
                          onAbort={() => {
                            refetchTaskExecs().then(() => {
                              message.success('Task successfully aborted')
                            })
                          }}
                          taskExec={row}
                          workspaceId={props.workspaceCtx.workspace.id}
                          apiPOST={props.workspaceCtx.apiPOST}
                        />
                        <ButtonTaskAbout
                          taskExec={row}
                          accountTimezone={accountCtx.account?.account.timezone as string}
                          workspaceId={props.workspaceCtx.workspace.id}
                          apiGET={props.workspaceCtx.apiGET}
                        />
                      </Space>
                    </div>
                  )
                }
              ]}
            />
          </Block>
        </Drawer>
      )}
    </>
  )
}

type RunButtonProps = {
  task: Task
  refetchTasks: any
  refetchTaskExecs: any
  workspaceCtx: CurrentWorkspaceCtxValue
}

const RunButton = (props: RunButtonProps) => {
  const [lauchingTask, setLauchingTask] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [taskState, setTaskState] = useState<string | null>(null)

  const toggleModal = () => {
    setModalVisible(!modalVisible)
  }

  const run = (taskID: string) => {
    let stateObject: any | null = null

    // verify json is valid
    if (taskState && taskState.length > 0) {
      try {
        stateObject = JSON.parse(taskState)
      } catch (e) {
        message.error('Invalid JSON')
        return
      }
    }

    if (lauchingTask) {
      return
    }
    setLauchingTask(true)

    const data: any = {
      workspace_id: props.workspaceCtx.workspace.id,
      id: taskID
      // main_worker_state: taskState
    }

    if (stateObject !== null) {
      data['main_worker_state'] = stateObject
    }
    props.workspaceCtx
      .apiPOST('/task.run', data)
      .then(() => {
        // refetch tasks + task execs
        props
          .refetchTasks()
          .then(() => {
            props
              .refetchTaskExecs()
              .then(() => {
                message.success('Task launched')
                setLauchingTask(false)
                toggleModal()
              })
              .catch((e: Error) => {
                message.error(e.message)
                setLauchingTask(false)
              })
          })
          .catch((e: Error) => {
            message.error(e.message)
            setLauchingTask(false)
          })
      })
      .catch((e) => {
        message.error(e.message)
        setLauchingTask(false)
      })
  }
  return (
    <Tooltip title="Launch manually" placement="left">
      <Button type="primary" ghost size="small" loading={lauchingTask} onClick={toggleModal}>
        <FontAwesomeIcon icon={faPlay} />
      </Button>
      {modalVisible && (
        <Modal
          title="Manual task launch"
          open={modalVisible}
          onOk={run.bind(null, props.task.id)}
          onCancel={toggleModal}
          okText="Launch"
          cancelText="Cancel"
        >
          {/* default JSON app state */}
          <p>(optional) Start main worker with JSON state:</p>
          <TextArea
            style={{ height: 150 }}
            placeholder="{}"
            onChange={(e) => {
              setTaskState(e.target.value)
            }}
          />
        </Modal>
      )}
    </Tooltip>
  )
}

export default TasksAppButton
