import { Table, Button, message, Tooltip, Popconfirm, Space } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faArrowRight, faRefresh, faShuffle } from '@fortawesome/free-solid-svg-icons'
import ListAccounts from 'components/account/block_list_accounts'
import ButtonOrganizationSettings from './button_settings'
import CreateWorkspaceButton from 'components/workspace/button_create'
import { Workspace } from 'interfaces'
import { useCurrentOrganizationCtx } from './context_current_organization'
import Layout from 'components/common/layout'
import { useState } from 'react'
import CSS from 'utils/css'
import Block from 'components/common/block'
import { css } from '@emotion/css'

export const ErrorNotOwner = () => {
  message.error('Only the owner of this organization can perform this task.')
}

const RouteOrganizationDashboard = () => {
  const currentOrgCtx = useCurrentOrganizationCtx()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const resetDemo = (workspace: Workspace) => {
    if (loading) return
    setLoading(true)

    currentOrgCtx
      .apiPOST('/workspace.createOrResetDemo', {
        organization_id: currentOrgCtx.organization.id,
        kind: workspace.demo_kind
      })
      .then(() => {
        message.success('The demo workspace has been resetted!')
        setLoading(false)
        navigate('/orgs/' + currentOrgCtx.organization.id + '/workspaces/' + workspace.id)
      })
      .catch((_) => {
        setLoading(false)
      })
  }

  const workspaces = currentOrgCtx.workspaces.filter((x: Workspace) => !x.deleted_at)

  return (
    <Layout currentOrganization={currentOrgCtx.organization}>
      <div className={CSS.container}>
        <div className={CSS.top}>
          <h1>{currentOrgCtx.organization.name}</h1>
          <div className={CSS.topSeparator}></div>
          <div>
            <Space>
              <ButtonOrganizationSettings
                organization={currentOrgCtx.organization}
                organizations={currentOrgCtx.organizations}
                updateOrganization={currentOrgCtx.updateOrganization}
                apiPOST={currentOrgCtx.apiPOST}
              />

              {currentOrgCtx.organizations.length > 1 && (
                <Button ghost onClick={() => navigate('/orgs')}>
                  <FontAwesomeIcon icon={faShuffle} />
                </Button>
              )}
            </Space>
          </div>
        </div>

        <>
          {workspaces.length === 0 && (
            <div className={css([CSS.blockCTA, CSS.padding_v_l])}>
              <h2>Welcome to Rimdian</h2>

              <p>
                It's time to create your first workspace,
                <br />
                you can generate a demo environment filled with dummy data to discover the platform,
                or configure a real workspace for your business!
              </p>

              <div className={CSS.margin_t_l}>
                <CreateWorkspaceButton
                  imOwner={currentOrgCtx.organization.im_owner}
                  organizationId={currentOrgCtx.organization.id}
                  text={
                    <>
                      Create a new workspace &nbsp;
                      <FontAwesomeIcon icon={faArrowRight} />
                    </>
                  }
                  btnType="primary"
                  btnSize="large"
                  apiPOST={currentOrgCtx.apiPOST}
                  refreshWorkspaces={currentOrgCtx.refreshWorkspaces}
                />
              </div>
            </div>
          )}

          {workspaces.length > 0 && (
            <Block
              classNames={[CSS.margin_b_l]}
              title="Workspaces"
              extra={
                <>
                  <Tooltip title="New workspace">
                    &nbsp;
                    <CreateWorkspaceButton
                      imOwner={currentOrgCtx.organization.im_owner}
                      organizationId={currentOrgCtx.organization.id}
                      text={<>New workspace</>}
                      btnType="link"
                      btnSize="middle"
                      apiPOST={currentOrgCtx.apiPOST}
                      refreshWorkspaces={currentOrgCtx.refreshWorkspaces}
                    />
                  </Tooltip>
                </>
              }
            >
              <Table
                showHeader={false}
                pagination={false}
                dataSource={workspaces}
                rowKey="id"
                columns={[
                  {
                    key: 'name',
                    render: (p: Workspace) => (
                      <div>
                        <div>
                          <b>{p.name}</b>
                        </div>
                        <small>
                          <a href={p.website_url} rel="noreferrer noopener" target="blank">
                            {p.website_url}
                          </a>
                        </small>
                      </div>
                    )
                  },
                  {
                    key: 'actions',
                    render: (p: Workspace) => {
                      let isReady = true
                      if (p.domains.length === 0) isReady = false
                      if (p.has_orders === false && p.has_leads === false) isReady = false

                      return (
                        <div className={CSS.text_right}>
                          {p.is_demo && (
                            <Popconfirm
                              placement="topRight"
                              title="Do you really want to reset the demo with fresh data?"
                              onConfirm={resetDemo.bind(null, p)}
                              okText="Confirm"
                              cancelText="Cancel"
                            >
                              <Button
                                type="default"
                                className={CSS.margin_r_s}
                                icon={<FontAwesomeIcon icon={faRefresh} />}
                                loading={loading}
                              >
                                &nbsp; Reset demo
                              </Button>
                            </Popconfirm>
                          )}
                          {/* {p.demoKind && <Popconfirm placement="topRight" title={t('delete_demo_confirm', "Do you really want to delete this demo?")} onConfirm={deleteDemo.bind(null, p.id)} okText={t('confirm', "Confirm")} cancelText={t('cancel', "Cancel")}>
                            <Button type="default" className={CSS.margin_r_s} icon={<DeleteOutlined />} loading={loading} />
                        </Popconfirm>} */}
                          <Link
                            to={'/orgs/' + currentOrgCtx.organization?.id + '/workspaces/' + p.id}
                          >
                            <Button type="primary">
                              {!isReady ? 'Continue setup' : 'View'} &nbsp;
                              <FontAwesomeIcon icon={faArrowRight} />
                            </Button>
                          </Link>
                        </div>
                      )
                    }
                  }
                ]}
              />
            </Block>
          )}
        </>

        <ListAccounts organization={currentOrgCtx.organization} />
      </div>
    </Layout>
  )
}

export default RouteOrganizationDashboard
