import { Button, Drawer, Popover, Spin, Table, TablePaginationConfig, Tooltip } from 'antd'
import { TaskExec, TaskExecJob, TaskExecJobInfoInfo } from 'interfaces'
import Code from 'utils/prism'
import { ButtonType } from 'antd/lib/button'
import { SizeType } from 'antd/lib/config-provider/SizeContext'
import dayjs from 'dayjs'
import { useState } from 'react'
import CSS, { backgroundColorBase } from 'utils/css'
import { useAccount } from 'components/login/context_account'
import { useQuery } from '@tanstack/react-query'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import Block from 'components/common/block'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faInfoCircle } from '@fortawesome/free-solid-svg-icons'
import { BlockDataLog } from 'components/item_timeline/block_list'

interface TaskExecJobsResult {
  task_exec_jobs: TaskExecJob[]
  total: number
  offset: number
  limit: number
}

interface ButtonTaskStateProps {
  taskExec: TaskExec
  accountTimezone: string
  workspaceId: string
  apiGET: (endpoint: string) => Promise<any>
  btnType?: ButtonType
  btnSize?: SizeType
}

const ButtonTaskAbout = (props: ButtonTaskStateProps) => {
  const [visible, setVisible] = useState(false)
  const accountCtx = useAccount()
  const [offset, setOffset] = useState(0)

  const limit = 5

  // fetch jobs
  const { isLoading, data } = useQuery<TaskExecJobsResult>(
    ['task_exec_jobs', props.workspaceId, props.taskExec.id, offset],
    (): Promise<TaskExecJobsResult> => {
      return new Promise((resolve, reject) => {
        props
          .apiGET(
            '/taskExec.jobs?workspace_id=' +
              props.workspaceId +
              '&task_exec_id=' +
              props.taskExec.id +
              '&limit=' +
              limit +
              '&offset=' +
              offset
          )
          .then(resolve)
          .catch(reject)
      })
    },
    {
      enabled: visible
    }
  )

  // compute current page from offset
  const currentPage = offset === 0 ? 1 : offset / limit + 1

  return (
    <>
      <Button size={props.btnSize || 'small'} type={props.btnType} onClick={() => setVisible(true)}>
        About
      </Button>

      {visible && (
        <Drawer
          title="About task"
          open={visible}
          width={900}
          onClose={() => setVisible(false)}
          headerStyle={{ backgroundColor: backgroundColorBase }}
          bodyStyle={{ backgroundColor: backgroundColorBase }}
        >
          <>
            <p>
              <b>Last update: </b> {dayjs(props.taskExec.db_updated_at).fromNow()} -{' '}
              {dayjs(props.taskExec.db_created_at).tz(props.accountTimezone).format('lll')} in{' '}
              {props.accountTimezone}
            </p>

            <Block title="Jobs" small>
              <Table
                dataSource={data?.task_exec_jobs || []}
                size="small"
                loading={isLoading}
                rowKey="id"
                onChange={(pagination: TablePaginationConfig) => {
                  const page = pagination.current || 1
                  const newOffset = page === 0 ? 0 : (page - 1) * limit
                  setOffset(newOffset)
                }}
                pagination={{
                  pageSize: data?.limit || limit,
                  total: data?.total,
                  current: currentPage,
                  hideOnSinglePage: true
                }}
                columns={[
                  {
                    title: 'ID',
                    key: 'id',
                    render: (x: TaskExecJob) => {
                      return <span className={CSS.font_size_xs}>{x.id}</span>
                    }
                  },
                  {
                    title: 'Created at',
                    key: 'created_at',
                    render: (x: TaskExecJob) => {
                      return (
                        <Tooltip
                          title={
                            <span>
                              {dayjs(x.db_created_at)
                                .tz(accountCtx.account?.account.timezone as string)
                                .format('lll')}
                              <br />
                              In {accountCtx.account?.account.timezone}
                            </span>
                          }
                        >
                          <span className={CSS.font_size_xs}>
                            {dayjs(x.db_created_at).format('lll')}
                          </span>
                        </Tooltip>
                      )
                    }
                  },
                  {
                    title: 'Took',
                    key: 'took',
                    render: (x: TaskExecJob) => {
                      if (!x.done_at)
                        return (
                          <Popover
                            content={
                              <RunningJobInfo
                                accountTimezone={accountCtx.account?.account.timezone || 'UTC'}
                                jobId={x.id}
                                taskExecId={props.taskExec.id}
                              />
                            }
                          >
                            <Button
                              type="primary"
                              ghost
                              size="small"
                              icon={
                                <FontAwesomeIcon className={CSS.padding_r_xs} icon={faInfoCircle} />
                              }
                            >
                              Running...
                            </Button>
                          </Popover>
                        )
                      return (
                        <span className={CSS.font_size_xs}>
                          {dayjs.duration(dayjs(x.db_created_at).diff(x.done_at)).humanize()}
                        </span>
                      )
                    }
                  }
                ]}
              />
            </Block>

            <Block title="State" small>
              <Code language="json">{JSON.stringify(props.taskExec.state, null, 2)}</Code>
            </Block>

            <BlockDataLog
              workspaceId={props.workspaceId}
              apiGET={props.apiGET}
              origin={4}
              originId={props.taskExec.id}
            />
          </>
        </Drawer>
      )}
    </>
  )
}

interface RunningJobInfoProps {
  taskExecId: string
  jobId: string
  accountTimezone: string
}

const RunningJobInfo = (props: RunningJobInfoProps) => {
  const workspaceCtx = useCurrentWorkspaceCtx()

  // running job
  const { isLoading, data } = useQuery<TaskExecJobInfoInfo>(
    ['jobInfo', props.jobId, props.taskExecId],
    (): Promise<TaskExecJobInfoInfo> => {
      return new Promise((resolve, reject) => {
        workspaceCtx
          .apiGET(
            '/taskExec.jobInfo?workspace_id=' +
              workspaceCtx.workspace.id +
              '&task_exec_id=' +
              props.taskExecId +
              '&id=' +
              props.jobId
          )
          .then((data: any) => {
            // console.log('job info', data)
            resolve(data as TaskExecJobInfoInfo)
          })
          .catch(reject)
      })
    }
  )

  if (isLoading) return <Spin size="small" />

  if (!data) return <>no data</>

  return (
    <div style={{ width: 600 }}>
      <div>
        <b>Created at: </b> {dayjs(data.create_time).tz(props.accountTimezone).format('lll')} in{' '}
        {props.accountTimezone}
      </div>
      {data.schedule_time && (
        <div>
          <b>Scheduled at: </b> {dayjs(data.schedule_time).tz(props.accountTimezone).format('lll')}{' '}
          in {props.accountTimezone}
        </div>
      )}
      <div>
        <b>Dispatch count: </b> {'' + data.dispatch_count}
      </div>
      <div>
        <b>Response count: </b> {'' + data.response_count}
      </div>
      {data.first_attempt && (
        <>
          <div style={{ marginTop: 20 }}>
            <b>First attempt: </b>
            <div style={{ marginLeft: 20 }}>
              <div>
                <b>Scheduled at: </b>
                {dayjs(data.first_attempt.schedule_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              <div>
                <b>Dispatched at: </b>
                {dayjs(data.first_attempt.dispatch_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              <div>
                <b>Response at: </b>
                {dayjs(data.first_attempt.response_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              {data.first_attempt.response_code && (
                <>
                  <div>
                    <b>Response code: </b>
                    {'' + data.first_attempt.response_code}
                  </div>
                  <div>
                    <b>Response message: </b>
                    {data.first_attempt.response_message}
                  </div>
                </>
              )}
            </div>
          </div>
        </>
      )}
      {data.last_attempt && (
        <>
          <div style={{ marginTop: 20 }}>
            <b>Last attempt: </b>
            <div style={{ marginLeft: 20 }}>
              <div>
                <b>Scheduled at: </b>
                {dayjs(data.last_attempt.schedule_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              <div>
                <b>Dispatched at: </b>
                {dayjs(data.last_attempt.dispatch_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              <div>
                <b>Response at: </b>
                {dayjs(data.last_attempt.response_time)
                  .tz(props.accountTimezone)
                  .format('lll')} in {props.accountTimezone}
              </div>
              {data.last_attempt.response_code && (
                <>
                  <div>
                    <b>Response code: </b>
                    {'' + data.last_attempt.response_code}
                  </div>
                  <div>
                    <b>Response message: </b>
                    {data.last_attempt.response_message}
                  </div>
                </>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
export default ButtonTaskAbout
