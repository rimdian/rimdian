import { useState } from 'react'
import { Alert, Button, message, Modal, Table } from 'antd'
import { AppManifest, TableColumn, ExtraColumnsManifest } from 'interfaces'
import { CurrentWorkspaceCtxValue } from 'components/workspace/context_current_workspace'
import { useNavigate } from 'react-router-dom'
import { css } from '@emotion/css'
import CSS from 'utils/css'

type DeleteAppButtonProps = {
  manifest: AppManifest
  workspaceCtx: CurrentWorkspaceCtxValue
}

const DeleteAppButton = (props: DeleteAppButtonProps) => {
  const [drawerVisible, setModalVisible] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()

  const closeModal = () => {
    setModalVisible(false)
  }

  const deleteApp = () => {
    if (isLoading) return
    setIsLoading(true)

    props.workspaceCtx
      .apiPOST('/app.delete', {
        workspace_id: props.workspaceCtx.workspace.id,
        id: props.manifest.id
      })
      .then(() => {
        // reload apps list
        props.workspaceCtx
          .refetchApps()
          .then(() => {
            setIsLoading(false)

            message.success(props.manifest.name + ' has been deleted.')
            // redirect to apps list
            navigate(
              '/orgs/' +
                props.workspaceCtx.organization.id +
                '/workspaces/' +
                props.workspaceCtx.workspace.id +
                '/apps'
            )
          })
          .catch(() => {})
      })
      .catch((e) => {
        setIsLoading(false)
      })
  }

  // console.log('initialValues', initialValues);

  return (
    <>
      <Button
        size="small"
        type="text"
        danger
        ghost
        loading={isLoading}
        onClick={() => setModalVisible(true)}
        // icon={<FontAwesomeIcon icon={faRemove} className={CSS.padding_r_xs} />}
      >
        Delete
      </Button>
      {drawerVisible && (
        <Modal
          title={
            <>
              <img
                src={props.manifest.icon_url}
                className={css(CSS.appIcon, CSS.margin_r_m)}
                style={{ height: 30 }}
                alt=""
              />
              Delete {props.manifest.name}
            </>
          }
          width={640}
          open={true}
          footer={[
            <Button key="a" loading={isLoading} onClick={closeModal} style={{ marginRight: 8 }}>
              Cancel
            </Button>,
            <Button key="b" loading={isLoading} onClick={deleteApp} danger type="primary">
              Delete
            </Button>
          ]}
        >
          <Alert
            message="You are about to delete this app, its tables & extra columns data will be deleted."
            type="warning"
            showIcon
            className={CSS.margin_v_l}
          />
          {props.manifest.app_tables && props.manifest.app_tables?.length > 0 && (
            <>
              <Table
                dataSource={props.manifest.app_tables}
                pagination={false}
                size="small"
                columns={[
                  {
                    title: 'Custom table',
                    dataIndex: 'name',
                    key: 'name',
                    render: (text: string) => <span>{text}</span>
                  },
                  {
                    title: 'Description',
                    dataIndex: 'description',
                    key: 'description',
                    render: (text: string) => <span>{text}</span>
                  }
                ]}
              />
            </>
          )}
          {props.manifest.extra_columns && props.manifest.extra_columns?.length > 0 && (
            <>
              <Table
                dataSource={props.manifest.extra_columns}
                pagination={false}
                size="small"
                className={CSS.margin_t_m}
                columns={[
                  {
                    title: 'Extra columns',
                    key: 'extra_columns',
                    render: (row: ExtraColumnsManifest) => {
                      return row.columns.map((col: TableColumn) => {
                        return (
                          <div key={col.name}>
                            {row.kind}.{col.name}
                          </div>
                        )
                      })
                    }
                  },
                  {
                    title: 'Description',
                    key: 'description',
                    render: (row: ExtraColumnsManifest) => {
                      return row.columns.map((col: TableColumn) => {
                        return <div key={col.name}>{' ' + col.description}</div>
                      })
                    }
                  }
                ]}
              />
            </>
          )}
        </Modal>
      )}
    </>
  )
}

export default DeleteAppButton
