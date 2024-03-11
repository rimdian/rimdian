import { Button, Popconfirm, Space, Switch, Table, message } from 'antd'
import { DataHook, DataHookFor } from 'interfaces'
import { useCurrentWorkspaceCtx } from 'components/workspace/context_current_workspace'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'
import UpsertDataHookButton from './button_upsert'
import CSS from 'utils/css'
import { blockCss } from 'components/common/block'
import { useState } from 'react'
import TableTag from 'components/common/partial_table_tag'

const BlockDataHooks = () => {
  const workspaceCtx = useCurrentWorkspaceCtx()
  const [loading, setLoading] = useState(false)

  const toggleDataHookEnabled = (dataHook: DataHook) => {
    if (loading) return
    setLoading(true)

    workspaceCtx
      .apiPOST(
        '/dataHook.update',
        Object.assign(dataHook, {
          workspace_id: workspaceCtx.workspace.id,
          enabled: dataHook.enabled ? false : true
        })
      )
      .then(() => {
        workspaceCtx
          .refreshWorkspace()
          .then(() => {
            message.success('The data hook has been updated!')
            setLoading(false)
          })
          .catch(() => {
            setLoading(false)
          })
      })
      .catch(() => {
        setLoading(false)
      })
  }

  return (
    <>
      <Table
        pagination={false}
        className={blockCss.self}
        dataSource={workspaceCtx.workspace.data_hooks}
        rowKey="id"
        columns={[
          {
            title: 'ID',
            key: 'id',
            render: (hook: DataHook) => hook.id
          },
          {
            title: 'Name',
            key: 'id',
            render: (hook: DataHook) => hook.name
          },
          {
            title: 'Origin',
            key: 'origin',
            render: (hook: DataHook) => {
              if (hook.app_id === 'system') return 'system'
              // find the app
              const app = workspaceCtx.workspace.installed_apps.find((app: any) => {
                return app.id === hook.app_id
              })

              if (app)
                return (
                  <>
                    <Space>
                      <img alt={app.id} src={app.icon_url} width={16} height={16} />
                      {app.name}
                    </Space>
                  </>
                )
              return hook.app_id
            }
          },
          {
            title: 'On',
            key: 'on',
            render: (hook: DataHook) => hook.on
          },
          {
            title: 'For',
            key: 'kind',
            render: (hook: DataHook) => {
              return hook.for.map((x: DataHookFor) => (
                <div key={x.kind}>
                  <TableTag table={x.kind} /> - {x.action}
                </div>
              ))
            }
          },
          {
            title: (
              <div className={CSS.text_right}>
                {/* <UpsertDataHookButton
                  organizationId={workspaceCtx.organization.id}
                  workspaceId={workspaceCtx.workspace.id}
                  btnSize="small"
                  btnType="primary"
                  btnContent={
                    <>
                      <FontAwesomeIcon icon={faPlus} />
                      &nbsp; New data hook
                    </>
                  }
                  apiPOST={workspaceCtx.apiPOST}
                  refreshWorkspace={workspaceCtx.refreshWorkspace}
                /> */}
              </div>
            ),
            key: 'actions',
            className: 'actions',
            render: (row: DataHook) => {
              let title = 'Do you really want to enable this data hook?'
              let okText = 'Enable'
              const okButtonProps: any = {
                loading: loading
              }

              if (row.enabled) {
                title = 'Do you really want to disable this data hook?'
                okButtonProps['danger'] = true
                okText = 'Disable'
              }

              return (
                <div className={CSS.text_right}>
                  <Button.Group>
                    {/* enable/disable switch */}
                    <Popconfirm
                      title={title}
                      onConfirm={toggleDataHookEnabled.bind(null, row)}
                      placement="left"
                      okButtonProps={okButtonProps}
                      okText={okText}
                    >
                      <Switch loading={loading} size="small" checked={row.enabled} />
                    </Popconfirm>

                    {row.app_id === 'system' && (
                      <>
                        <Popconfirm title="Are you sure?" onConfirm={() => {}}>
                          <Button type="primary" size="small">
                            <FontAwesomeIcon icon={faTrashCan} />
                          </Button>
                        </Popconfirm>

                        <UpsertDataHookButton
                          organizationId={workspaceCtx.organization.id}
                          workspaceId={workspaceCtx.workspace.id}
                          domain={row}
                          btnSize="small"
                          btnType="text"
                          btnContent={<FontAwesomeIcon icon={faPenToSquare} />}
                          apiPOST={workspaceCtx.apiPOST}
                          refreshWorkspace={workspaceCtx.refreshWorkspace}
                        />
                      </>
                    )}
                  </Button.Group>
                </div>
              )
            }
          }
        ]}
      />
    </>
  )
}

export default BlockDataHooks
